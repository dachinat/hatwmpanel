package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include <stdlib.h>
#include "panel.h"
*/
import "C"

import (
	"log/slog"
	"unsafe"
)

func configuredButtons(cfg Config, tiling bool) ([]Module, []int, []int) {
	var buttons []Module
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
			if module.Kind == "button" || module.Kind == "layout_mode" {
				if module.Kind == "layout_mode" {
					if tiling {
						module.Icon = module.TilingIcon
					} else {
						module.Icon = module.FloatingIcon
					}
				}
				buttons = append(buttons, module)
				groups = append(groups, group.value)
				orders = append(orders, index)
			}
		}
	}
	return buttons, groups, orders
}

func (p *Panel) SetButtons(cfg Config, tiling bool) {
	if p == nil || p.ptr == nil {
		return
	}
	buttons, groups, orders := configuredButtons(cfg, tiling)
	p.buttons = append(p.buttons[:0], buttons...)

	cButtons := make([]C.HatWMPanelButton, len(buttons))
	texts := make([]*C.char, len(buttons))
	icons := make([]*C.char, len(buttons))
	for i, button := range buttons {
		texts[i] = C.CString(button.Value)
		icons[i] = C.CString(button.Icon)
		iconColor := cfg.ModuleIconColor
		if button.HasIconColor {
			iconColor = button.IconColor
		}
		iconTint := iconTintValue(button, button.Icon)
		cButtons[i] = C.HatWMPanelButton{
			group:      C.int(groups[i]),
			text:       texts[i],
			icon:       icons[i],
			icon_color: C.uint32_t(iconColor),
			icon_tint:  C.int(iconTint),
			order:      C.int(orders[i]),
		}
	}
	defer func() {
		for i := range texts {
			C.free(unsafe.Pointer(texts[i]))
			C.free(unsafe.Pointer(icons[i]))
		}
	}()

	var buttonPtr *C.HatWMPanelButton
	if len(cButtons) > 0 {
		buttonPtr = &cButtons[0]
	}
	C.hatwm_panel_set_buttons(p.ptr, buttonPtr, C.int(len(cButtons)))
}

func (p *Panel) ButtonKind(index int) string {
	if p == nil || index < 0 || index >= len(p.buttons) {
		return ""
	}
	return p.buttons[index].Kind
}

func (p *Panel) TakeButtonClick() int {
	if p == nil || p.ptr == nil {
		return -1
	}
	return int(C.hatwm_panel_take_button_click(p.ptr)) - 1
}

func (p *Panel) ActivateButton(index int) {
	if p == nil || index < 0 || index >= len(p.buttons) {
		return
	}
	button := p.buttons[index]
	wait, err := activateButton(button)
	if err != nil {
		slog.Error("button action failed", "button", button.Name, "error", err)
		return
	}
	go func() {
		if err := wait(); err != nil {
			slog.Warn("button action exited with an error",
				"button", button.Name, "error", err)
		}
	}()
}
