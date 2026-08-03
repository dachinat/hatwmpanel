package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include "panel.h"
*/
import "C"

func (p *Panel) SetBrightness(percent int) {
	if p == nil || p.ptr == nil {
		return
	}
	C.hatwm_panel_set_brightness(p.ptr, C.int(percent))
}

func (p *Panel) TakeBrightnessChange() int {
	if p == nil || p.ptr == nil {
		return -1
	}
	return int(C.hatwm_panel_take_brightness_change(p.ptr))
}
