// Package microphone controls the default PipeWire/PulseAudio input source.
package microphone

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type State struct {
	Percent int
	Muted   bool
}

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

func Read() (State, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()
	if output, err := exec.CommandContext(
		ctx, "wpctl", "get-volume", "@DEFAULT_AUDIO_SOURCE@").Output(); err == nil {
		return parseWPCTL(string(output))
	}
	output, err := exec.CommandContext(
		ctx, "pactl", "get-source-volume", "@DEFAULT_SOURCE@").Output()
	if err != nil {
		return State{}, err
	}
	state, err := parsePactl(string(output))
	if err != nil {
		return State{}, err
	}
	muteOutput, _ := exec.CommandContext(
		ctx, "pactl", "get-source-mute", "@DEFAULT_SOURCE@").Output()
	state.Muted = strings.Contains(strings.ToLower(string(muteOutput)), "yes")
	return state, nil
}

func Set(percent int) error {
	percent = clampPercent(percent)
	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()
	value := strconv.Itoa(percent) + "%"
	if err := exec.CommandContext(
		ctx, "wpctl", "set-volume", "@DEFAULT_AUDIO_SOURCE@", value).Run(); err == nil {
		_ = exec.CommandContext(
			ctx, "wpctl", "set-mute", "@DEFAULT_AUDIO_SOURCE@", "0").Run()
		return nil
	}
	if err := exec.CommandContext(
		ctx, "pactl", "set-source-volume", "@DEFAULT_SOURCE@", value).Run(); err != nil {
		return err
	}
	_ = exec.CommandContext(
		ctx, "pactl", "set-source-mute", "@DEFAULT_SOURCE@", "0").Run()
	return nil
}

func parseWPCTL(output string) (State, error) {
	fields := strings.Fields(output)
	if len(fields) < 2 {
		return State{}, fmt.Errorf("unexpected wpctl output: %q", output)
	}
	level, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return State{}, err
	}
	return State{
		Percent: clampPercent(int(level*100 + 0.5)),
		Muted:   strings.Contains(strings.ToUpper(output), "[MUTED]"),
	}, nil
}

func parsePactl(output string) (State, error) {
	for _, field := range strings.Fields(output) {
		if !strings.HasSuffix(field, "%") {
			continue
		}
		percent, err := strconv.Atoi(strings.TrimSuffix(field, "%"))
		if err == nil {
			return State{Percent: clampPercent(percent)}, nil
		}
	}
	return State{}, fmt.Errorf("unexpected pactl output: %q", output)
}

func clampPercent(percent int) int {
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}
