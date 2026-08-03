package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include "panel.h"
*/
import "C"

import "hatwmpanel/internal/system/battery"

func (p *Panel) SetBattery(state battery.State) {
	if p == nil || p.ptr == nil {
		return
	}
	charging, full := 0, 0
	if state.Charging {
		charging = 1
	}
	if state.Full {
		full = 1
	}
	C.hatwm_panel_set_battery(
		p.ptr, C.int(state.Percent), C.int(charging), C.int(full))
}
