package native

import (
	"testing"
)

func TestConfiguredLayoutModeButtonUsesCurrentIcon(t *testing.T) {
	cfg := Config{Left: []Module{{
		Name: "layout", Kind: "layout_mode",
		TilingIcon: "view-grid-symbolic", FloatingIcon: "window-restore-symbolic",
	}}}

	buttons, groups, orders := configuredButtons(cfg, true)
	if len(buttons) != 1 || buttons[0].Icon != "view-grid-symbolic" ||
		len(groups) != 1 || len(orders) != 1 || orders[0] != 0 {
		t.Fatalf("unexpected tiling button: %#v %#v %#v", buttons, groups, orders)
	}
	buttons, _, _ = configuredButtons(cfg, false)
	if len(buttons) != 1 || buttons[0].Icon != "window-restore-symbolic" {
		t.Fatalf("unexpected floating button: %#v", buttons)
	}
}
