package storage

import "testing"

func TestUsageString(t *testing.T) {
	usage := Usage{
		Used:      161 * 1024 * 1024 * 1024,
		Available: 61 * 1024 * 1024 * 1024,
	}
	if got := usage.String(); got != "161G / 61G" {
		t.Fatalf("Usage.String() = %q", got)
	}
}

func TestFormatBytes(t *testing.T) {
	for bytes, want := range map[uint64]string{
		0: "0B", 512: "512B", 1536: "1.5K",
		10 * 1024 * 1024:              "10M",
		(60 * 1024 * 1024 * 1024) + 1: "61G",
	} {
		if got := formatBytes(bytes); got != want {
			t.Fatalf("formatBytes(%d) = %q, want %q", bytes, got, want)
		}
	}
}
