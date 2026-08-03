// Package netstat samples aggregate network throughput from procfs.
package netstat

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type counters struct {
	received    uint64
	transmitted uint64
}

type Rate struct {
	Download uint64
	Upload   uint64
}

type Monitor struct {
	previous counters
	updated  time.Time
	haveLast bool
}

func (m *Monitor) Read() (Rate, error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return Rate{}, err
	}
	defer file.Close()
	return m.sample(file, time.Now())
}

func (m *Monitor) sample(reader io.Reader, now time.Time) (Rate, error) {
	current, err := parseCounters(reader)
	if err != nil {
		return Rate{}, err
	}
	if !m.haveLast {
		m.previous = current
		m.updated = now
		m.haveLast = true
		return Rate{}, nil
	}
	elapsed := now.Sub(m.updated)
	previous := m.previous
	m.previous = current
	m.updated = now
	if elapsed <= 0 {
		return Rate{}, nil
	}
	var received, transmitted uint64
	if current.received >= previous.received {
		received = current.received - previous.received
	}
	if current.transmitted >= previous.transmitted {
		transmitted = current.transmitted - previous.transmitted
	}
	seconds := elapsed.Seconds()
	return Rate{
		Download: uint64(math.Round(float64(received) / seconds)),
		Upload:   uint64(math.Round(float64(transmitted) / seconds)),
	}, nil
}

func parseCounters(reader io.Reader) (counters, error) {
	var result counters
	found := false
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		colon := strings.IndexByte(line, ':')
		if colon < 0 {
			continue
		}
		name := strings.TrimSpace(line[:colon])
		if name == "" || name == "lo" {
			continue
		}
		fields := strings.Fields(line[colon+1:])
		if len(fields) < 16 {
			continue
		}
		received, receiveErr := strconv.ParseUint(fields[0], 10, 64)
		transmitted, transmitErr := strconv.ParseUint(fields[8], 10, 64)
		if receiveErr != nil || transmitErr != nil {
			continue
		}
		result.received += received
		result.transmitted += transmitted
		found = true
	}
	if err := scanner.Err(); err != nil {
		return counters{}, err
	}
	if !found {
		return counters{}, fmt.Errorf("no network interfaces found")
	}
	return result, nil
}

func (r Rate) String() string {
	return fmt.Sprintf("↓ %s ↑ %s", formatRate(r.Download), formatRate(r.Upload))
}

func formatRate(bytes uint64) string {
	const unit = uint64(1024)
	if bytes < unit {
		return fmt.Sprintf("%dB/s", bytes)
	}
	value := float64(bytes)
	suffix := "K"
	for _, candidate := range []string{"K", "M", "G", "T"} {
		suffix = candidate
		value /= 1024
		if value < 1024 {
			break
		}
	}
	if value >= 100 {
		return fmt.Sprintf("%.0f%s/s", value, suffix)
	}
	if value >= 10 {
		return fmt.Sprintf("%.1f%s/s", value, suffix)
	}
	return fmt.Sprintf("%.2f%s/s", value, suffix)
}
