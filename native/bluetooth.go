package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include <stdlib.h>
#include "panel.h"
*/
import "C"

import (
	"hatwmpanel/internal/connectivity/bluetooth"
	"unsafe"
)

func (p *Panel) SetBluetoothDevices(devices []bluetooth.Device) {
	if p == nil || p.ptr == nil {
		return
	}
	cDevices := make([]C.HatWMPanelBluetoothDevice, len(devices))
	addresses := make([]*C.char, len(devices))
	names := make([]*C.char, len(devices))
	for i, device := range devices {
		addresses[i] = C.CString(device.Address)
		names[i] = C.CString(device.Name)
		connected := 0
		if device.Connected {
			connected = 1
		}
		cDevices[i] = C.HatWMPanelBluetoothDevice{
			address:   addresses[i],
			name:      names[i],
			connected: C.int(connected),
		}
	}
	defer func() {
		for i := range addresses {
			C.free(unsafe.Pointer(addresses[i]))
			C.free(unsafe.Pointer(names[i]))
		}
	}()
	var ptr *C.HatWMPanelBluetoothDevice
	if len(cDevices) > 0 {
		ptr = &cDevices[0]
	}
	C.hatwm_panel_set_bluetooth_devices(p.ptr, ptr, C.int(len(cDevices)))
}

func (p *Panel) TakeBluetoothMenuRequest() bool {
	return p != nil && p.ptr != nil &&
		C.hatwm_panel_take_bluetooth_menu_request(p.ptr) != 0
}

func (p *Panel) ShowBluetoothMenu() bool {
	return p != nil && p.ptr != nil &&
		C.hatwm_panel_show_bluetooth_menu(p.ptr) != 0
}

func (p *Panel) TakeBluetoothClick() int {
	if p == nil || p.ptr == nil {
		return -1
	}
	return int(C.hatwm_panel_take_bluetooth_click(p.ptr)) - 1
}

func (p *Panel) TakeBluetoothDisconnect() bool {
	return p != nil && p.ptr != nil &&
		C.hatwm_panel_take_bluetooth_disconnect(p.ptr) != 0
}
