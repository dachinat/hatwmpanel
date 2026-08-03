package tray

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/godbus/dbus/v5"
)

func TestItemsReturnsSnapshot(t *testing.T) {
	tray := &Tray{items: []TrayItem{{Service: "original"}}}
	items := tray.Items()
	items[0].Service = "changed"

	if got := tray.Items()[0].Service; got != "original" {
		t.Fatalf("Items() exposed internal state: got %q", got)
	}
	if items := (*Tray)(nil).Items(); items != nil {
		t.Fatalf("nil Tray.Items() = %#v, want nil", items)
	}
}

func TestFindThemeIconInRootsUsesRootAndSizePriority(t *testing.T) {
	preferredRoot := t.TempDir()
	fallbackRoot := t.TempDir()
	iconName := "hatwm-test-icon"

	paths := []string{
		filepath.Join(preferredRoot, "16x16", "apps", iconName+".png"),
		filepath.Join(preferredRoot, "22x22", "apps", iconName+".svg"),
		filepath.Join(fallbackRoot, "24x24", "apps", iconName+".svg"),
	}
	for _, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("icon"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if got := findThemeIconInRoots(iconName, []string{preferredRoot, fallbackRoot}); got != paths[1] {
		t.Fatalf("findThemeIconInRoots() = %q, want %q", got, paths[1])
	}
	if got := findThemeIconInRoots("missing", []string{preferredRoot}); got != "" {
		t.Fatalf("missing icon path = %q, want empty", got)
	}
}

func TestMenuValidatesItemAndActivation(t *testing.T) {
	tray := &Tray{items: []TrayItem{
		{Service: "org.example.Regular", Menu: dbus.ObjectPath("/Menu")},
		{Service: "org.example.Menu", Menu: dbus.ObjectPath("/Menu"), ItemIsMenu: true},
		{Service: "org.example.Missing", Menu: dbus.ObjectPath("/")},
	}}

	if _, _, ok := tray.Menu(0, 1); ok {
		t.Fatal("regular item exposed its menu for a left click")
	}
	if service, path, ok := tray.Menu(0, 3); !ok ||
		service != "org.example.Regular" || path != "/Menu" {
		t.Fatalf("right-click Menu() = %q, %q, %v", service, path, ok)
	}
	if _, _, ok := tray.Menu(1, 1); !ok {
		t.Fatal("menu-only item rejected a left click")
	}
	if _, _, ok := tray.Menu(2, 3); ok {
		t.Fatal("item with root menu path was accepted")
	}
	if _, _, ok := tray.Menu(-1, 3); ok {
		t.Fatal("negative tray index was accepted")
	}
}

func TestWatcherPropertiesReflectRegisteredItems(t *testing.T) {
	tray := &Tray{items: []TrayItem{{
		Service: "org.example.Item", Path: dbus.ObjectPath("/StatusNotifierItem"),
	}}}
	watcher := &trayWatcher{tray: tray}

	variant, dbusErr := watcher.Get("", "RegisteredStatusNotifierItems")
	if dbusErr != nil {
		t.Fatal(dbusErr)
	}
	items, ok := variant.Value().([]string)
	if !ok || len(items) != 1 || items[0] != "org.example.Item/StatusNotifierItem" {
		t.Fatalf("registered items = %#v", variant.Value())
	}
	if _, dbusErr := watcher.Get("", "MissingProperty"); dbusErr == nil {
		t.Fatal("unknown watcher property returned no D-Bus error")
	}
	properties, dbusErr := watcher.GetAll("")
	if dbusErr != nil || len(properties) != 3 {
		t.Fatalf("GetAll() = %#v, %v", properties, dbusErr)
	}
	if dbusErr := watcher.Set("", "ProtocolVersion", dbus.MakeVariant(int32(1))); dbusErr == nil {
		t.Fatal("read-only watcher property accepted a write")
	}
}
