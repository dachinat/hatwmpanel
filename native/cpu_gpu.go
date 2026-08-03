package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include <stdlib.h>
#include "panel.h"
*/
import "C"

import "unsafe"

func (p *Panel) SetCPU(text string) {
	if p == nil || p.ptr == nil {
		return
	}
	value := C.CString(text)
	defer C.free(unsafe.Pointer(value))
	C.hatwm_panel_set_cpu(p.ptr, value)
}

func (p *Panel) SetGPU(text string) {
	if p == nil || p.ptr == nil {
		return
	}
	value := C.CString(text)
	defer C.free(unsafe.Pointer(value))
	C.hatwm_panel_set_gpu(p.ptr, value)
}

func (p *Panel) SetNetstat(text string) {
	if p == nil || p.ptr == nil {
		return
	}
	value := C.CString(text)
	defer C.free(unsafe.Pointer(value))
	C.hatwm_panel_set_netstat(p.ptr, value)
}
