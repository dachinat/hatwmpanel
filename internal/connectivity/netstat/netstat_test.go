package netstat

import (
	"strings"
	"testing"
	"time"
)

const sample = `Inter-|   Receive                                                |  Transmit
 face |bytes packets errs drop fifo frame compressed multicast|bytes packets errs drop fifo colls carrier compressed
    lo: 1000 1 0 0 0 0 0 0 2000 1 0 0 0 0 0 0
wlan0: 4096 2 0 0 0 0 0 0 2048 2 0 0 0 0 0 0
  eth0: 1024 1 0 0 0 0 0 0 512 1 0 0 0 0 0 0
`

func TestParseCountersExcludesLoopback(t *testing.T) {
	got, err := parseCounters(strings.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}
	if got.received != 5120 || got.transmitted != 2560 {
		t.Fatalf("unexpected counters: %#v", got)
	}
}

func TestMonitorCalculatesRates(t *testing.T) {
	monitor := Monitor{}
	start := time.Unix(100, 0)
	if _, err := monitor.sample(strings.NewReader(sample), start); err != nil {
		t.Fatal(err)
	}
	next := strings.Replace(sample, "wlan0: 4096", "wlan0: 8192", 1)
	next = strings.Replace(next, "0 2048 2", "0 4096 2", 1)
	rate, err := monitor.sample(strings.NewReader(next), start.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if rate.Download != 2048 || rate.Upload != 1024 {
		t.Fatalf("unexpected rate: %#v", rate)
	}
}

func TestRateString(t *testing.T) {
	got := (Rate{Download: 1536, Upload: 512}).String()
	if got != "↓ 1.50K/s ↑ 512B/s" {
		t.Fatalf("unexpected display: %q", got)
	}
}

func TestMonitorHandlesCounterResetAndNonPositiveElapsedTime(t *testing.T) {
	monitor := Monitor{}
	start := time.Unix(200, 0)
	high := "eth0: 9000 0 0 0 0 0 0 0 8000 0 0 0 0 0 0 0\n"
	low := "eth0: 1000 0 0 0 0 0 0 0 500 0 0 0 0 0 0 0\n"
	if _, err := monitor.sample(strings.NewReader(high), start); err != nil {
		t.Fatal(err)
	}
	rate, err := monitor.sample(strings.NewReader(low), start.Add(time.Second))
	if err != nil || rate != (Rate{}) {
		t.Fatalf("counter-reset rate = %#v, %v", rate, err)
	}
	rate, err = monitor.sample(strings.NewReader(high), start.Add(time.Second))
	if err != nil || rate != (Rate{}) {
		t.Fatalf("zero-elapsed rate = %#v, %v", rate, err)
	}
}

func TestParseCountersRequiresUsableInterface(t *testing.T) {
	for _, input := range []string{
		"",
		"lo: 100 0 0 0 0 0 0 0 200 0 0 0 0 0 0 0\n",
		"wlan0: malformed\n",
	} {
		if _, err := parseCounters(strings.NewReader(input)); err == nil {
			t.Fatalf("parseCounters(%q) returned no error", input)
		}
	}
}

func TestFormatRateUnitBoundaries(t *testing.T) {
	for bytes, want := range map[uint64]string{
		0:                  "0B/s",
		1024:               "1.00K/s",
		10 * 1024:          "10.0K/s",
		100 * 1024:         "100K/s",
		1024 * 1024:        "1.00M/s",
		1024 * 1024 * 1024: "1.00G/s",
	} {
		if got := formatRate(bytes); got != want {
			t.Fatalf("formatRate(%d) = %q, want %q", bytes, got, want)
		}
	}
}
