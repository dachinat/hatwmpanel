package tray

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

const (
	trayWatcherPath = dbus.ObjectPath("/StatusNotifierWatcher")
	trayWatcherIF   = "org.kde.StatusNotifierWatcher"
)

type TrayItem struct {
	Service    string
	Path       dbus.ObjectPath
	Icon       []uint32
	IconW      int
	IconH      int
	IconPath   string
	Status     string
	ItemIsMenu bool
	Menu       dbus.ObjectPath
}

type statusNotifierPixmap struct {
	Width  int32
	Height int32
	Data   []byte
}

type Tray struct {
	conn         *dbus.Conn
	mu           sync.RWMutex
	items        []TrayItem
	fallbackIcon string
}

type trayWatcher struct{ tray *Tray }

func NewTray() (*Tray, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, err
	}
	t := &Tray{
		conn:         conn,
		fallbackIcon: findThemeIcon("application-x-executable"),
	}
	reply, err := conn.RequestName(trayWatcherIF, dbus.NameFlagDoNotQueue)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		conn.Close()
		return nil, nil // another watcher already owns the tray
	}
	w := &trayWatcher{tray: t}
	if err := conn.Export(w, trayWatcherPath, trayWatcherIF); err != nil {
		conn.Close()
		return nil, err
	}
	if err := conn.Export(w, trayWatcherPath, "org.freedesktop.DBus.Properties"); err != nil {
		conn.Close()
		return nil, err
	}
	node := &introspect.Node{
		Name: string(trayWatcherPath),
		Interfaces: []introspect.Interface{
			{
				Name: trayWatcherIF,
				Methods: []introspect.Method{
					{Name: "RegisterStatusNotifierItem", Args: []introspect.Arg{{Name: "service", Type: "s", Direction: "in"}}},
					{Name: "RegisterStatusNotifierHost", Args: []introspect.Arg{{Name: "service", Type: "s", Direction: "in"}}},
				},
				Signals: []introspect.Signal{
					{Name: "StatusNotifierItemRegistered", Args: []introspect.Arg{{Name: "service", Type: "s"}}},
					{Name: "StatusNotifierItemUnregistered", Args: []introspect.Arg{{Name: "service", Type: "s"}}},
					{Name: "StatusNotifierHostRegistered"},
				},
				Properties: []introspect.Property{
					{Name: "RegisteredStatusNotifierItems", Type: "as", Access: "read"},
					{Name: "IsStatusNotifierHostRegistered", Type: "b", Access: "read"},
					{Name: "ProtocolVersion", Type: "i", Access: "read"},
				},
			},
			{
				Name: "org.freedesktop.DBus.Properties",
				Methods: []introspect.Method{
					{Name: "Get", Args: []introspect.Arg{
						{Name: "interface", Type: "s", Direction: "in"},
						{Name: "property", Type: "s", Direction: "in"},
						{Name: "value", Type: "v", Direction: "out"},
					}},
					{Name: "GetAll", Args: []introspect.Arg{
						{Name: "interface", Type: "s", Direction: "in"},
						{Name: "properties", Type: "a{sv}", Direction: "out"},
					}},
					{Name: "Set", Args: []introspect.Arg{
						{Name: "interface", Type: "s", Direction: "in"},
						{Name: "property", Type: "s", Direction: "in"},
						{Name: "value", Type: "v", Direction: "in"},
					}},
				},
			},
			introspect.IntrospectData,
		},
	}
	if err := conn.Export(
		introspect.NewIntrospectable(node),
		trayWatcherPath,
		"org.freedesktop.DBus.Introspectable",
	); err != nil {
		conn.Close()
		return nil, err
	}
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus"),
		dbus.WithMatchMember("NameOwnerChanged"),
	); err != nil {
		conn.Close()
		return nil, err
	}
	for _, member := range []string{"NewIcon", "NewAttentionIcon", "NewOverlayIcon", "NewStatus"} {
		if err := conn.AddMatchSignal(
			dbus.WithMatchInterface("org.kde.StatusNotifierItem"),
			dbus.WithMatchMember(member),
		); err != nil {
			conn.Close()
			return nil, err
		}
	}
	signals := make(chan *dbus.Signal, 16)
	conn.Signal(signals)
	go t.watchNames(signals)
	_ = conn.Emit(trayWatcherPath, trayWatcherIF+".StatusNotifierHostRegistered")
	return t, nil
}

func (t *Tray) Close() {
	if t != nil && t.conn != nil {
		t.conn.Close()
	}
}

func (t *Tray) Items() []TrayItem {
	if t == nil {
		return nil
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return append([]TrayItem(nil), t.items...)
}

func (t *Tray) Interact(index, button, x, y int) {
	items := t.Items()
	if index < 0 || index >= len(items) {
		return
	}
	item := items[index]
	go func() {
		method := "Activate"
		if button == 2 {
			method = "SecondaryActivate"
		} else if button == 3 || item.ItemIsMenu {
			method = "ContextMenu"
		}
		call := t.conn.Object(item.Service, item.Path).Call(
			"org.kde.StatusNotifierItem."+method, 0, int32(x), int32(y))
		if call.Err != nil {
			slog.Debug("tray interaction failed", "service", item.Service,
				"method", method, "error", call.Err)
		}
	}()
}

func (t *Tray) register(sender dbus.Sender, service string) *dbus.Error {
	busName := string(sender)
	path := dbus.ObjectPath("/StatusNotifierItem")
	if strings.HasPrefix(service, "/") {
		path = dbus.ObjectPath(service)
	} else if service != "" {
		busName = service
	}
	if !path.IsValid() {
		return dbus.MakeFailedError(dbus.ErrMsgInvalidArg)
	}

	item := TrayItem{
		Service:  busName,
		Path:     path,
		IconPath: t.fallbackIcon,
	}
	t.mu.Lock()
	for _, existing := range t.items {
		if existing.Service == item.Service && existing.Path == item.Path {
			t.mu.Unlock()
			return nil
		}
	}
	t.items = append(t.items, item)
	t.mu.Unlock()
	_ = t.conn.Emit(trayWatcherPath, trayWatcherIF+".StatusNotifierItemRegistered", busName+string(path))
	// RegisterStatusNotifierItem is commonly a synchronous D-Bus call. Do not
	// call back into the item for properties until its registration call has
	// returned, otherwise single-threaded clients such as blueman cannot answer
	// and the item is permanently left with its fallback marker.
	go t.loadItem(item)
	return nil
}

func (t *Tray) loadItem(item TrayItem) {
	item.Status = t.itemStringProperty(item, "Status")
	item.ItemIsMenu = t.itemBoolProperty(item, "ItemIsMenu")
	item.Menu = t.itemObjectPathProperty(item, "Menu")
	item.IconPath = t.itemIconPath(item)
	if item.IconPath == "" {
		item.Icon, item.IconW, item.IconH = t.itemIcon(item)
		if len(item.Icon) == 0 {
			item.IconPath = t.fallbackIcon
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	for i := range t.items {
		if t.items[i].Service == item.Service && t.items[i].Path == item.Path {
			t.items[i] = item
			return
		}
	}
}

func (t *Tray) itemObjectPathProperty(item TrayItem, property string) dbus.ObjectPath {
	v, err := t.conn.Object(item.Service, item.Path).
		GetProperty("org.kde.StatusNotifierItem." + property)
	if err != nil {
		return ""
	}
	value, _ := v.Value().(dbus.ObjectPath)
	if !value.IsValid() {
		return ""
	}
	return value
}

func (t *Tray) Menu(index, button int) (string, dbus.ObjectPath, bool) {
	items := t.Items()
	if index < 0 || index >= len(items) ||
		(button != 3 && !items[index].ItemIsMenu) ||
		items[index].Menu == "" || items[index].Menu == "/" {
		return "", "", false
	}
	return items[index].Service, items[index].Menu, true
}

func (t *Tray) itemStringProperty(item TrayItem, property string) string {
	v, err := t.conn.Object(item.Service, item.Path).
		GetProperty("org.kde.StatusNotifierItem." + property)
	if err != nil {
		return ""
	}
	value, _ := v.Value().(string)
	return strings.TrimSpace(value)
}

func (t *Tray) itemBoolProperty(item TrayItem, property string) bool {
	v, err := t.conn.Object(item.Service, item.Path).
		GetProperty("org.kde.StatusNotifierItem." + property)
	if err != nil {
		return false
	}
	value, _ := v.Value().(bool)
	return value
}

func (t *Tray) itemIconPath(item TrayItem) string {
	property := "IconName"
	if item.Status == "NeedsAttention" {
		property = "AttentionIconName"
	}
	v, err := t.conn.Object(item.Service, item.Path).
		GetProperty("org.kde.StatusNotifierItem." + property)
	if err != nil {
		if property != "IconName" {
			copy := item
			copy.Status = ""
			return t.itemIconPath(copy)
		}
		return ""
	}
	name, ok := v.Value().(string)
	name = strings.TrimSpace(name)
	if !ok || name == "" {
		return ""
	}
	if filepath.IsAbs(name) {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}
	if themePath := t.itemStringProperty(item, "IconThemePath"); themePath != "" {
		if path := findThemeIconInRoots(name, []string{themePath}); path != "" {
			return path
		}
	}
	return findThemeIcon(name)
}

func findThemeIcon(name string) string {
	name = strings.TrimSuffix(strings.TrimSuffix(name, ".svg"), ".png")
	if !strings.HasSuffix(name, "-symbolic") {
		if path := findThemeIconExact(name + "-symbolic"); path != "" {
			return path
		}
	}
	return findThemeIconExact(name)
}

func findThemeIconExact(name string) string {
	home, _ := os.UserHomeDir()
	roots := []string{
		filepath.Join(home, ".icons"),
		filepath.Join(home, ".local", "share", "icons"),
		"/usr/share/icons/Reversal-dark",
		"/usr/share/icons/hicolor",
		"/usr/share/pixmaps",
		"/usr/share/icons",
	}
	return findThemeIconInRoots(name, roots)
}

func findThemeIconInRoots(name string, roots []string) string {
	bestPath, bestScore := "", -1
	for rootIndex, root := range roots {
		_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext != ".svg" && ext != ".png" {
				return nil
			}
			if strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())) != name {
				return nil
			}
			score := 1000 - rootIndex*100
			slashed := filepath.ToSlash(path)
			switch {
			case strings.Contains(slashed, "/22"):
				score += 30
			case strings.Contains(slashed, "/24"):
				score += 20
			case strings.Contains(slashed, "/16"):
				score += 10
			}
			if ext == ".svg" {
				score++
			}
			if score > bestScore {
				bestPath, bestScore = path, score
			}
			return nil
		})
		if bestPath != "" {
			break
		}
	}
	return bestPath
}

func (t *Tray) itemIcon(item TrayItem) ([]uint32, int, int) {
	property := "IconPixmap"
	if item.Status == "NeedsAttention" {
		property = "AttentionIconPixmap"
	}
	v, err := t.conn.Object(item.Service, item.Path).
		GetProperty("org.kde.StatusNotifierItem." + property)
	if err != nil {
		if property != "IconPixmap" {
			copy := item
			copy.Status = ""
			return t.itemIcon(copy)
		}
		return nil, 0, 0
	}
	var pixmaps []statusNotifierPixmap
	if err := v.Store(&pixmaps); err != nil {
		return nil, 0, 0
	}

	// Pick the image nearest the panel's usual tray size. StatusNotifierItem
	// pixels are non-premultiplied ARGB in network byte order; Cairo expects
	// native-endian, premultiplied ARGB32.
	best := -1
	bestDistance := int(^uint(0) >> 1)
	for i, pixmap := range pixmaps {
		if pixmap.Width <= 0 || pixmap.Height <= 0 ||
			len(pixmap.Data) != int(pixmap.Width*pixmap.Height*4) {
			continue
		}
		distance := int(pixmap.Width) - 22
		if distance < 0 {
			distance = -distance
		}
		if distance < bestDistance {
			best, bestDistance = i, distance
		}
	}
	if best < 0 {
		return nil, 0, 0
	}

	pixmap := pixmaps[best]
	pixels := make([]uint32, pixmap.Width*pixmap.Height)
	for i := range pixels {
		a := uint32(pixmap.Data[i*4])
		r := uint32(pixmap.Data[i*4+1]) * a / 255
		g := uint32(pixmap.Data[i*4+2]) * a / 255
		b := uint32(pixmap.Data[i*4+3]) * a / 255
		pixels[i] = a<<24 | r<<16 | g<<8 | b
	}
	return pixels, int(pixmap.Width), int(pixmap.Height)
}

func (t *Tray) watchNames(signals <-chan *dbus.Signal) {
	for signal := range signals {
		if signal == nil {
			continue
		}
		if signal.Name == "org.kde.StatusNotifierItem.NewIcon" ||
			signal.Name == "org.kde.StatusNotifierItem.NewAttentionIcon" ||
			signal.Name == "org.kde.StatusNotifierItem.NewOverlayIcon" ||
			signal.Name == "org.kde.StatusNotifierItem.NewStatus" {
			t.refreshIcon(signal.Sender, signal.Path)
			continue
		}
		if signal.Name != "org.freedesktop.DBus.NameOwnerChanged" || len(signal.Body) != 3 {
			continue
		}
		name, nameOK := signal.Body[0].(string)
		newOwner, ownerOK := signal.Body[2].(string)
		if !nameOK || !ownerOK || newOwner != "" {
			continue
		}
		t.mu.Lock()
		for i := len(t.items) - 1; i >= 0; i-- {
			if t.items[i].Service == name {
				removed := t.items[i]
				t.items = append(t.items[:i], t.items[i+1:]...)
				_ = t.conn.Emit(trayWatcherPath, trayWatcherIF+".StatusNotifierItemUnregistered",
					removed.Service+string(removed.Path))
			}
		}
		t.mu.Unlock()
	}
}

func (t *Tray) refreshIcon(sender string, path dbus.ObjectPath) {
	items := t.Items()
	for i, item := range items {
		if item.Path != path {
			continue
		}
		matches := item.Service == sender
		if !matches && !strings.HasPrefix(item.Service, ":") {
			var owner string
			err := t.conn.BusObject().Call(
				"org.freedesktop.DBus.GetNameOwner", 0, item.Service,
			).Store(&owner)
			if err == nil {
				matches = owner == sender
			}
		}
		if !matches {
			continue
		}
		item.Status = t.itemStringProperty(item, "Status")
		item.ItemIsMenu = t.itemBoolProperty(item, "ItemIsMenu")
		iconPath := t.itemIconPath(item)
		var icon []uint32
		var width, height int
		if iconPath == "" {
			icon, width, height = t.itemIcon(item)
			if len(icon) == 0 {
				iconPath = t.fallbackIcon
			}
		}
		t.mu.Lock()
		if i < len(t.items) && t.items[i].Service == item.Service && t.items[i].Path == item.Path {
			t.items[i].Icon = icon
			t.items[i].IconW = width
			t.items[i].IconH = height
			t.items[i].IconPath = iconPath
			t.items[i].Status = item.Status
			t.items[i].ItemIsMenu = item.ItemIsMenu
		}
		t.mu.Unlock()
		return
	}
}

func (w *trayWatcher) RegisterStatusNotifierItem(sender dbus.Sender, service string) *dbus.Error {
	return w.tray.register(sender, service)
}

func (w *trayWatcher) RegisterStatusNotifierHost(_ dbus.Sender, _ string) *dbus.Error {
	return nil
}

func (w *trayWatcher) Get(_ string, property string) (dbus.Variant, *dbus.Error) {
	switch property {
	case "RegisteredStatusNotifierItems":
		items := w.tray.Items()
		names := make([]string, len(items))
		for i, item := range items {
			names[i] = item.Service + string(item.Path)
		}
		return dbus.MakeVariant(names), nil
	case "IsStatusNotifierHostRegistered":
		return dbus.MakeVariant(true), nil
	case "ProtocolVersion":
		return dbus.MakeVariant(int32(0)), nil
	default:
		return dbus.Variant{}, dbus.NewError(
			"org.freedesktop.DBus.Error.UnknownProperty",
			[]any{"unknown property " + property},
		)
	}
}

func (w *trayWatcher) GetAll(_ string) (map[string]dbus.Variant, *dbus.Error) {
	items, _ := w.Get("", "RegisteredStatusNotifierItems")
	return map[string]dbus.Variant{
		"RegisteredStatusNotifierItems":  items,
		"IsStatusNotifierHostRegistered": dbus.MakeVariant(true),
		"ProtocolVersion":                dbus.MakeVariant(int32(0)),
	}, nil
}

func (w *trayWatcher) Set(_ string, _ string, _ dbus.Variant) *dbus.Error {
	return dbus.NewError("org.freedesktop.DBus.Error.PropertyReadOnly", nil)
}

func (t *Tray) Run(ctx context.Context) {
	if t == nil {
		return
	}
	<-ctx.Done()
}
