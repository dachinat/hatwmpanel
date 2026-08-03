package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include "panel.h"
*/
import "C"

import "hatwmpanel/internal/audio/microphone"

func (p *Panel) SetMicrophone(state microphone.State) {
	if p == nil || p.ptr == nil {
		return
	}
	muted := 0
	if state.Muted {
		muted = 1
	}
	C.hatwm_panel_set_microphone(p.ptr, C.int(state.Percent), C.int(muted))
}

func (p *Panel) TakeMicrophoneChange() int {
	if p == nil || p.ptr == nil {
		return -1
	}
	return int(C.hatwm_panel_take_microphone_change(p.ptr))
}
