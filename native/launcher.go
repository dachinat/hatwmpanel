package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include "panel.h"
*/
import "C"

import "unsafe"

func (p *Panel) SetLauncherLogo() {
	if p == nil || p.ptr == nil {
		return
	}
	data := LogoPNG()
	if len(data) == 0 {
		return
	}
	C.hatwm_panel_set_launcher_logo_png(
		p.ptr, (*C.uint8_t)(unsafe.Pointer(&data[0])), C.int(len(data)))
}

func (p *Panel) TakeLauncherClick() bool {
	return p != nil && p.ptr != nil &&
		C.hatwm_panel_take_launcher_click(p.ptr) != 0
}

func (p *Panel) ShowLauncher() bool {
	return p != nil && p.ptr != nil &&
		C.hatwm_panel_show_launcher(p.ptr) != 0
}
