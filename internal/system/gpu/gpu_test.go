package gpu

import "testing"

func TestParsePercent(t *testing.T) {
	for input, want := range map[string]int{
		"42\n": 42,
		"-1":   0,
		"120":  100,
	} {
		got, err := parsePercent(input)
		if err != nil || got != want {
			t.Fatalf("parsePercent(%q) = %d, %v; want %d", input, got, err, want)
		}
	}
}

func TestParsePercentRejectsNonNumericTelemetry(t *testing.T) {
	for _, input := range []string{"", "N/A", "42.5"} {
		if _, err := parsePercent(input); err == nil {
			t.Fatalf("parsePercent(%q) returned no error", input)
		}
	}
}
