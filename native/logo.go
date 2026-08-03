package native

import _ "embed"

//go:embed logo.svg
var logoSVG []byte

// LogoSVG returns the bundled launcher artwork. Its color is applied by the
// native renderer so it can follow the launcher's icon_color setting.
func LogoSVG() []byte {
	return logoSVG
}
