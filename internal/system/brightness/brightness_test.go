package brightness

import "testing"

func TestParseBrightnessctl(t *testing.T) {
	percent, err := parseBrightnessctl("intel_backlight,backlight,12000,50%,24000\n")
	if err != nil || percent != 50 {
		t.Fatalf("percent=%d, err=%v", percent, err)
	}
}

func TestParseBrightnessctlClampsAndRejectsMalformedOutput(t *testing.T) {
	for input, want := range map[string]int{
		"display,backlight,0,0%,100\n":     1,
		"display,backlight,150,150%,100\n": 100,
	} {
		got, err := parseBrightnessctl(input)
		if err != nil || got != want {
			t.Fatalf("parseBrightnessctl(%q) = %d, %v; want %d", input, got, err, want)
		}
	}
	for _, input := range []string{"", "display,backlight", "display,backlight,10,nope%"} {
		if _, err := parseBrightnessctl(input); err == nil {
			t.Fatalf("parseBrightnessctl(%q) returned no error", input)
		}
	}
}
