// Package storage reads filesystem space usage.
package storage

import (
	"fmt"
	"math"
	"syscall"
)

type Usage struct {
	Used      uint64
	Available uint64
}

func Read(path string) (Usage, error) {
	if path == "" {
		path = "/"
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return Usage{}, err
	}
	blockSize := uint64(stat.Bsize)
	return Usage{
		Used:      (stat.Blocks - stat.Bfree) * blockSize,
		Available: stat.Bavail * blockSize,
	}, nil
}

func (u Usage) String() string {
	return fmt.Sprintf("%s / %s", formatBytes(u.Used), formatBytes(u.Available))
}

func formatBytes(bytes uint64) string {
	const unit = uint64(1024)
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	value := float64(bytes)
	suffixes := []string{"K", "M", "G", "T", "P"}
	suffix := suffixes[0]
	for _, candidate := range suffixes {
		suffix = candidate
		value /= 1024
		if value < 1024 {
			break
		}
	}
	if value >= 10 {
		return fmt.Sprintf("%.0f%s", math.Ceil(value), suffix)
	}
	return fmt.Sprintf("%.1f%s", value, suffix)
}
