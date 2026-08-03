package native

import (
	"testing"
)

func TestConfiguredSeparatorsPreserveGroupOrderAndWidth(t *testing.T) {
	cfg := Config{
		Left: []Module{
			{Name: "launcher", Kind: "launcher"},
			{Name: "gap", Kind: "separator", Width: 12},
			{Name: "workspaces", Kind: "workspaces"},
		},
		Right: []Module{
			{Name: "right-gap", Kind: "separator", Width: 20},
		},
	}
	separators, groups, orders := configuredSeparators(cfg)
	if len(separators) != 2 ||
		separators[0].Width != 12 || separators[1].Width != 20 ||
		groups[0] != 1 || groups[1] != 3 ||
		orders[0] != 1 || orders[1] != 0 {
		t.Fatalf("unexpected separators: %#v groups=%v orders=%v",
			separators, groups, orders)
	}
}
