// Package brightness controls the default display backlight.
package brightness

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Controller struct {
	requests chan int
}

func NewController() *Controller {
	controller := &Controller{requests: make(chan int, 1)}
	go func() {
		for percent := range controller.requests {
			_ = Set(percent)
		}
	}()
	return controller
}

func (c *Controller) SetPercent(percent int) {
	if c == nil {
		return
	}
	percent = clampPercent(percent)
	select {
	case c.requests <- percent:
	default:
		select {
		case <-c.requests:
		default:
		}
		c.requests <- percent
	}
}

func (c *Controller) Close() {
	if c != nil {
		close(c.requests)
	}
}

func Read() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()
	output, err := exec.CommandContext(ctx, "brightnessctl", "-m").Output()
	if err != nil {
		return 0, err
	}
	return parseBrightnessctl(string(output))
}

func Set(percent int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()
	value := strconv.Itoa(clampPercent(percent)) + "%"
	return exec.CommandContext(ctx, "brightnessctl", "set", value).Run()
}

func parseBrightnessctl(output string) (int, error) {
	line := strings.TrimSpace(strings.Split(output, "\n")[0])
	fields := strings.Split(line, ",")
	if len(fields) < 4 {
		return 0, fmt.Errorf("unexpected brightnessctl output: %q", output)
	}
	value := strings.TrimSpace(strings.TrimSuffix(fields[3], "%"))
	percent, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return clampPercent(percent), nil
}

func clampPercent(percent int) int {
	if percent < 1 {
		return 1
	}
	if percent > 100 {
		return 100
	}
	return percent
}
