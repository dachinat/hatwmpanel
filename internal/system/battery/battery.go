// Package battery reads battery charge state from Linux power_supply sysfs.
package battery

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type State struct {
	Percent  int
	Charging bool
	Full     bool
}

func Read() (State, error) {
	return readFrom("/sys/class/power_supply")
}

func readFrom(root string) (State, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return State{}, err
	}
	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		powerType, err := readValue(filepath.Join(path, "type"))
		if err != nil || !strings.EqualFold(powerType, "Battery") {
			continue
		}
		capacityText, err := readValue(filepath.Join(path, "capacity"))
		if err != nil {
			continue
		}
		percent, err := strconv.Atoi(capacityText)
		if err != nil {
			continue
		}
		if percent < 0 {
			percent = 0
		}
		if percent > 100 {
			percent = 100
		}
		status, _ := readValue(filepath.Join(path, "status"))
		return State{
			Percent:  percent,
			Charging: strings.EqualFold(status, "Charging"),
			Full:     strings.EqualFold(status, "Full"),
		}, nil
	}
	return State{}, fmt.Errorf("no battery found")
}

func readValue(path string) (string, error) {
	value, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(value)), nil
}
