package native

import (
	"bytes"
	"image/png"
	"testing"
)

func TestLogoPNGIsBundledAtPanelSize(t *testing.T) {
	logo := LogoPNG()
	if len(logo) == 0 {
		t.Fatal("launcher PNG is empty")
	}
	config, err := png.DecodeConfig(bytes.NewReader(logo))
	if err != nil {
		t.Fatalf("decode launcher PNG: %v", err)
	}
	if config.Width != 32 || config.Height != 32 {
		t.Fatalf("launcher PNG is %dx%d, want 32x32", config.Width, config.Height)
	}
}
