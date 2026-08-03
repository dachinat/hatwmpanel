package native

import (
	traymodule "hatwmpanel/internal/desktop/tray"
	"testing"

	"github.com/godbus/dbus/v5"
)

func TestSameTrayItemsComparesRenderedIdentityAndIcons(t *testing.T) {
	base := []traymodule.TrayItem{{
		Service: "org.example.Item",
		Path:    dbus.ObjectPath("/StatusNotifierItem"),
		Icon:    []uint32{0xff0000ff, 0x00ff00ff},
		IconW:   2,
		IconH:   1,
	}}
	copyItems := []traymodule.TrayItem{{
		Service: base[0].Service, Path: base[0].Path,
		Icon: append([]uint32(nil), base[0].Icon...), IconW: 2, IconH: 1,
	}}
	if !sameTrayItems(base, copyItems) {
		t.Fatal("equivalent tray items were considered different")
	}

	copyItems[0].Icon[0] = 0
	if sameTrayItems(base, copyItems) {
		t.Fatal("changed tray icon was considered unchanged")
	}
	if sameTrayItems(base, append(copyItems, traymodule.TrayItem{})) {
		t.Fatal("different tray item counts were considered equal")
	}
}
