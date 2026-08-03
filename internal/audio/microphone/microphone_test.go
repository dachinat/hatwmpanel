package microphone

import "testing"

func TestParseWPCTL(t *testing.T) {
	state, err := parseWPCTL("Volume: 0.42 [MUTED]\n")
	if err != nil || state.Percent != 42 || !state.Muted {
		t.Fatalf("unexpected state: %#v, err=%v", state, err)
	}
}

func TestParsePactl(t *testing.T) {
	state, err := parsePactl("Volume: front-left: 32768 / 50% / -18.06 dB")
	if err != nil || state.Percent != 50 {
		t.Fatalf("unexpected state: %#v, err=%v", state, err)
	}
}

func TestMicrophoneParsersHandleUnmutedAndMalformedOutput(t *testing.T) {
	state, err := parseWPCTL("Volume: 0.505")
	if err != nil || state.Percent != 51 || state.Muted {
		t.Fatalf("rounded parseWPCTL() = %#v, %v", state, err)
	}
	if _, err := parseWPCTL("Volume: invalid"); err == nil {
		t.Fatal("invalid wpctl level returned no error")
	}
	if _, err := parsePactl("Volume unavailable"); err == nil {
		t.Fatal("pactl output without a percentage returned no error")
	}
}
