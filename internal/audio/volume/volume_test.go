package volume

import "testing"

func TestParseWPCTL(t *testing.T) {
	state, err := parseWPCTL("Volume: 0.57 [MUTED]\n")
	if err != nil || state.Percent != 57 || !state.Muted {
		t.Fatalf("unexpected state: %#v, err=%v", state, err)
	}
}

func TestParsePactl(t *testing.T) {
	state, err := parsePactl("Volume: front-left: 42598 /  65% / -11.22 dB")
	if err != nil || state.Percent != 65 {
		t.Fatalf("unexpected state: %#v, err=%v", state, err)
	}
}

func TestVolumeParsersClampAndRejectMalformedOutput(t *testing.T) {
	for input, want := range map[string]int{
		"Volume: -0.25": 0,
		"Volume: 1.75":  100,
	} {
		state, err := parseWPCTL(input)
		if err != nil || state.Percent != want {
			t.Fatalf("parseWPCTL(%q) = %#v, %v; want %d%%", input, state, err, want)
		}
	}
	state, err := parsePactl("Volume: front-left: 98304 / 150% / 10.0 dB")
	if err != nil || state.Percent != 100 {
		t.Fatalf("clamped parsePactl() = %#v, %v", state, err)
	}
	for _, parse := range []func(string) (State, error){parseWPCTL, parsePactl} {
		if _, err := parse("unavailable"); err == nil {
			t.Fatal("malformed audio output returned no error")
		}
	}
}
