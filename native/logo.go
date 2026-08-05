package native

import _ "embed"

//go:embed logo.png
var logoPNG []byte

// LogoPNG returns the bundled launcher artwork at its native 32 px size.
func LogoPNG() []byte {
	return logoPNG
}
