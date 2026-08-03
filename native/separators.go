package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include "panel.h"
*/
import "C"

func configuredSeparators(cfg Config) ([]Module, []int, []int) {
	var separators []Module
	var groups []int
	var orders []int
	for _, group := range []struct {
		modules []Module
		value   int
	}{
		{cfg.Left, int(C.HATWM_PANEL_GROUP_LEFT)},
		{cfg.Center, int(C.HATWM_PANEL_GROUP_CENTER)},
		{cfg.Right, int(C.HATWM_PANEL_GROUP_RIGHT)},
	} {
		for index, module := range group.modules {
			if module.Kind != "separator" {
				continue
			}
			separators = append(separators, module)
			groups = append(groups, group.value)
			orders = append(orders, index)
		}
	}
	return separators, groups, orders
}

func (p *Panel) SetSeparators(cfg Config) {
	if p == nil || p.ptr == nil {
		return
	}
	separators, groups, orders := configuredSeparators(cfg)
	native := make([]C.HatWMPanelSeparator, len(separators))
	for i, separator := range separators {
		native[i] = C.HatWMPanelSeparator{
			group: C.int(groups[i]),
			width: C.int(separator.Width),
			order: C.int(orders[i]),
		}
	}
	var pointer *C.HatWMPanelSeparator
	if len(native) > 0 {
		pointer = &native[0]
	}
	C.hatwm_panel_set_separators(p.ptr, pointer, C.int(len(native)))
}
