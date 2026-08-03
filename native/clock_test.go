package native

import (
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	at := time.Date(2026, time.July, 28, 13, 5, 9, 0, time.UTC)
	if got, want := formatClock(at, "%a %d %b  %H:%M"), "Tue 28 Jul  13:05"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestFormatDefault(t *testing.T) {
	at := time.Date(2026, time.July, 28, 13, 5, 0, 0, time.UTC)
	if got, want := formatClock(at, ""), "Tue 28 Jul  13:05"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestFormatClockEscapesPercentAndPreservesUnknownTokens(t *testing.T) {
	location := time.FixedZone("test", 4*60*60)
	at := time.Date(2026, time.December, 3, 0, 7, 9, 0, location)
	got := formatClock(at, "%% %I:%M:%S %p %z %Q")
	if want := "% 12:07:09 AM +0400 %Q"; got != want {
		t.Fatalf("formatClock() = %q, want %q", got, want)
	}
}
