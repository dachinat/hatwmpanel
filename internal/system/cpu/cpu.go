// Package cpu samples aggregate CPU utilization from procfs.
package cpu

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type sample struct {
	total uint64
	idle  uint64
}

type Monitor struct {
	previous sample
	haveLast bool
}

func (m *Monitor) Read() (int, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, err
	}
	defer file.Close()
	current, err := parseSample(bufio.NewScanner(file))
	if err != nil {
		return 0, err
	}
	if !m.haveLast {
		m.previous = current
		m.haveLast = true
		return 0, nil
	}
	percent := utilization(m.previous, current)
	m.previous = current
	return percent, nil
}

func parseSample(scanner *bufio.Scanner) (sample, error) {
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return sample{}, err
		}
		return sample{}, fmt.Errorf("missing aggregate CPU line")
	}
	fields := strings.Fields(scanner.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return sample{}, fmt.Errorf("invalid aggregate CPU line")
	}
	var values []uint64
	for _, field := range fields[1:] {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return sample{}, err
		}
		values = append(values, value)
	}
	var total uint64
	for _, value := range values {
		total += value
	}
	idle := values[3]
	if len(values) > 4 {
		idle += values[4]
	}
	return sample{total: total, idle: idle}, nil
}

func utilization(previous, current sample) int {
	if current.total <= previous.total {
		return 0
	}
	totalDelta := current.total - previous.total
	idleDelta := uint64(0)
	if current.idle > previous.idle {
		idleDelta = current.idle - previous.idle
	}
	if idleDelta > totalDelta {
		idleDelta = totalDelta
	}
	return int(((totalDelta-idleDelta)*100 + totalDelta/2) / totalDelta)
}
