// Package gpu reads GPU utilization from common Linux telemetry interfaces.
package gpu

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func Read() (int, error) {
	files, _ := filepath.Glob("/sys/class/drm/card*/device/gpu_busy_percent")
	for _, path := range files {
		value, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if percent, err := parsePercent(string(value)); err == nil {
			return percent, nil
		}
	}

	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		output, err := exec.CommandContext(
			ctx, "nvidia-smi",
			"--query-gpu=utilization.gpu",
			"--format=csv,noheader,nounits",
		).Output()
		if err == nil {
			for _, line := range strings.Split(string(output), "\n") {
				if percent, parseErr := parsePercent(line); parseErr == nil {
					return percent, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("GPU utilization telemetry unavailable")
}

func parsePercent(value string) (int, error) {
	percent, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, err
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return percent, nil
}
