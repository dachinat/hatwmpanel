package cpu

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseSampleAndUtilization(t *testing.T) {
	before, err := parseSample(bufio.NewScanner(
		strings.NewReader("cpu  100 10 40 800 50 0 0 0 0 0\n")))
	if err != nil {
		t.Fatal(err)
	}
	after, err := parseSample(bufio.NewScanner(
		strings.NewReader("cpu  130 10 50 850 50 0 0 0 0 0\n")))
	if err != nil {
		t.Fatal(err)
	}
	if got := utilization(before, after); got != 44 {
		t.Fatalf("utilization() = %d, want 44", got)
	}
}

func TestParseSampleRejectsInvalidAggregateLines(t *testing.T) {
	for _, input := range []string{
		"",
		"cpu0 1 2 3 4\n",
		"cpu 1 2 3\n",
		"cpu 1 2 invalid 4\n",
	} {
		if _, err := parseSample(bufio.NewScanner(strings.NewReader(input))); err == nil {
			t.Fatalf("parseSample(%q) returned no error", input)
		}
	}
}

func TestUtilizationHandlesCounterResetAndIdleOverflow(t *testing.T) {
	if got := utilization(sample{total: 100, idle: 50}, sample{total: 90, idle: 40}); got != 0 {
		t.Fatalf("counter-reset utilization = %d, want 0", got)
	}
	if got := utilization(sample{total: 100, idle: 0}, sample{total: 200, idle: 250}); got != 0 {
		t.Fatalf("idle-overflow utilization = %d, want 0", got)
	}
	if got := utilization(sample{total: 100, idle: 50}, sample{total: 200, idle: 50}); got != 100 {
		t.Fatalf("fully busy utilization = %d, want 100", got)
	}
}
