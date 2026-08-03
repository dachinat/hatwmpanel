package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include <stdlib.h>
#include "panel.h"
*/
import "C"

import "unsafe"

func (p *Panel) SetKeyboardLayout(layout string) {
	if p == nil || p.ptr == nil {
		return
	}
	value := C.CString(layout)
	defer C.free(unsafe.Pointer(value))
	C.hatwm_panel_set_keyboard_layout(p.ptr, value)
}

func (p *Panel) TakeKeyboardLayoutClick() bool {
	return p != nil && p.ptr != nil &&
		C.hatwm_panel_take_keyboard_layout_click(p.ptr) != 0
}
