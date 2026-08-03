// Package memory reads system RAM usage from procfs.
package memory

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Usage struct {
	Used      uint64
	Available uint64
}

func Read() (Usage, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return Usage{}, err
	}
	defer file.Close()
	return parse(file)
}

func parse(reader io.Reader) (Usage, error) {
	values := make(map[string]uint64)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err == nil {
			values[key] = value * 1024
		}
	}
	if err := scanner.Err(); err != nil {
		return Usage{}, err
	}
	total := values["MemTotal"]
	available := values["MemAvailable"]
	if available == 0 {
		available = values["MemFree"] + values["Buffers"] + values["Cached"]
	}
	if total == 0 || available > total {
		return Usage{}, fmt.Errorf("invalid memory information")
	}
	return Usage{Used: total - available, Available: available}, nil
}

func (u Usage) String() string {
	return fmt.Sprintf("%s / %s", formatBytes(u.Used), formatBytes(u.Available))
}

func formatBytes(bytes uint64) string {
	const (
		mebibyte = uint64(1024 * 1024)
		gibibyte = uint64(1024 * 1024 * 1024)
	)
	if bytes < gibibyte {
		return fmt.Sprintf("%.0fM", float64(bytes)/float64(mebibyte))
	}
	return fmt.Sprintf("%.1fG", float64(bytes)/float64(gibibyte))
}
