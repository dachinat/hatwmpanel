package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include <stdlib.h>
#include "panel.h"
*/
import "C"

import (
	"hatwmpanel/internal/connectivity/network"
	"unsafe"
)

func (p *Panel) SetNetworks(networks []network.Network) {
	if p == nil || p.ptr == nil {
		return
	}
	cNetworks := make([]C.HatWMPanelNetwork, len(networks))
	ssids := make([]*C.char, len(networks))
	for i, item := range networks {
		ssids[i] = C.CString(item.SSID)
		secured, active, wired := 0, 0, 0
		if item.Secured() {
			secured = 1
		}
		if item.Active {
			active = 1
		}
		if item.Wired() {
			wired = 1
		}
		cNetworks[i] = C.HatWMPanelNetwork{
			ssid:    ssids[i],
			signal:  C.int(item.Signal),
			secured: C.int(secured),
			active:  C.int(active),
			wired:   C.int(wired),
		}
	}
	defer func() {
		for _, ssid := range ssids {
			C.free(unsafe.Pointer(ssid))
		}
	}()
	var ptr *C.HatWMPanelNetwork
	if len(cNetworks) > 0 {
		ptr = &cNetworks[0]
	}
	C.hatwm_panel_set_networks(p.ptr, ptr, C.int(len(cNetworks)))
}

func (p *Panel) TakeNetworkMenuRequest() bool {
	return p != nil && p.ptr != nil &&
		C.hatwm_panel_take_network_menu_request(p.ptr) != 0
}

func (p *Panel) ShowNetworkMenu() bool {
	return p != nil && p.ptr != nil &&
		C.hatwm_panel_show_network_menu(p.ptr) != 0
}

func (p *Panel) TakeNetworkClick() int {
	if p == nil || p.ptr == nil {
		return -1
	}
	return int(C.hatwm_panel_take_network_click(p.ptr)) - 1
}

func (p *Panel) TakeNetworkDisconnect() bool {
	return p != nil && p.ptr != nil &&
		C.hatwm_panel_take_network_disconnect(p.ptr) != 0
}
