package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include <stdlib.h>
#include "panel.h"
*/
import "C"

import "unsafe"

func (p *Panel) SetStorage(text string) {
	if p == nil || p.ptr == nil {
		return
	}
	value := C.CString(text)
	defer C.free(unsafe.Pointer(value))
	C.hatwm_panel_set_storage(p.ptr, value)
}
