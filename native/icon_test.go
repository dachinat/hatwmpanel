package native

import (
	"testing"
)

func TestIconTintSelection(t *testing.T) {
	if got := iconTintValue(Module{}, "archlinux"); got != 0 {
		t.Fatalf("regular icon tint = %d, want 0", got)
	}
	if got := iconTintValue(
		Module{}, "network-wireless-symbolic"); got != 1 {
		t.Fatalf("symbolic icon tint = %d, want 1", got)
	}
	if got := iconTintValue(
		Module{HasIconColor: true}, "archlinux"); got != 1 {
		t.Fatalf("explicitly colored regular icon tint = %d, want 1", got)
	}
}
