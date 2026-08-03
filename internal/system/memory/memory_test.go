package memory

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	usage, err := parse(strings.NewReader(
		"MemTotal:       16777216 kB\n" +
			"MemFree:        1024000 kB\n" +
			"MemAvailable:  14680064 kB\n"))
	if err != nil {
		t.Fatal(err)
	}
	if usage.Used != 2*1024*1024*1024 ||
		usage.Available != 14*1024*1024*1024 {
		t.Fatalf("unexpected memory usage: %#v", usage)
	}
	if got := usage.String(); got != "2.0G / 14.0G" {
		t.Fatalf("Usage.String() = %q", got)
	}
}

func TestParseFallbackAvailable(t *testing.T) {
	usage, err := parse(strings.NewReader(
		"MemTotal: 4096000 kB\nMemFree: 1024000 kB\n" +
			"Buffers: 512000 kB\nCached: 512000 kB\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got := usage.String(); got != "2.0G / 2.0G" {
		t.Fatalf("fallback Usage.String() = %q", got)
	}
}

func TestParseRejectsInvalidMemoryInformation(t *testing.T) {
	for _, input := range []string{
		"",
		"MemAvailable: 1024 kB\n",
		"MemTotal: 1024 kB\nMemAvailable: 2048 kB\n",
		"MemTotal: invalid kB\nMemAvailable: 512 kB\n",
	} {
		if _, err := parse(strings.NewReader(input)); err == nil {
			t.Fatalf("parse(%q) returned no error", input)
		}
	}
}
