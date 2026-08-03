package native

import (
	"bytes"
	"testing"
)

func TestLogoSVGIsBundled(t *testing.T) {
	logo := LogoSVG()
	if len(logo) == 0 {
		t.Fatal("launcher SVG is empty")
	}
	if !bytes.Contains(logo, []byte("<svg")) {
		t.Fatal("launcher asset is not SVG markup")
	}
}
