package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include <stdlib.h>
#include "panel.h"
*/
import "C"

import (
	traymodule "hatwmpanel/internal/desktop/tray"
	"slices"
	"unsafe"

	"github.com/godbus/dbus/v5"
)

func (p *Panel) SetTray(items []traymodule.TrayItem) {
	if p == nil || p.ptr == nil {
		return
	}
	if sameTrayItems(p.trayItems, items) {
		return
	}
	p.trayItems = append(p.trayItems[:0], items...)
	C.hatwm_panel_set_tray(p.ptr, C.int(len(items)))
	for i, item := range items {
		if len(item.Icon) == 0 || item.IconW <= 0 || item.IconH <= 0 {
			if item.IconPath != "" {
				path := C.CString(item.IconPath)
				C.hatwm_panel_set_tray_icon_file(p.ptr, C.int(i), path)
				C.free(unsafe.Pointer(path))
			}
			continue
		}
		C.hatwm_panel_set_tray_icon(
			p.ptr,
			C.int(i),
			(*C.uint32_t)(unsafe.Pointer(&item.Icon[0])),
			C.int(item.IconW),
			C.int(item.IconH),
		)
	}
	C.hatwm_panel_redraw(p.ptr)
}

func sameTrayItems(a, b []traymodule.TrayItem) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Service != b[i].Service || a[i].Path != b[i].Path ||
			a[i].IconW != b[i].IconW ||
			a[i].IconH != b[i].IconH || a[i].IconPath != b[i].IconPath ||
			!slices.Equal(a[i].Icon, b[i].Icon) {
			return false
		}
	}
	return true
}

func (p *Panel) TakeTrayClick() (int, int) {
	if p == nil || p.ptr == nil {
		return -1, 0
	}
	index := int(C.hatwm_panel_take_tray_click(p.ptr))
	button := int(C.hatwm_panel_take_tray_button(p.ptr))
	return index, button
}

func (p *Panel) ShowTrayMenu(service string, path dbus.ObjectPath) bool {
	if p == nil || p.ptr == nil || service == "" || !path.IsValid() || path == "/" {
		return false
	}
	cService := C.CString(service)
	cPath := C.CString(string(path))
	defer C.free(unsafe.Pointer(cService))
	defer C.free(unsafe.Pointer(cPath))
	return C.hatwm_panel_show_tray_menu(p.ptr, cService, cPath) != 0
}
