package main

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

type state struct {
	text    string
	updated time.Time
}

// Runner renders module groups and retains refresh state for exec modules.
type Runner struct {
	states map[string]state
}

func NewRunner() *Runner {
	return &Runner{states: make(map[string]state)}
}

func (r *Runner) Render(group []Module, now time.Time, separator string) string {
	parts := make([]string, 0, len(group))
	for _, module := range group {
		if text := r.render(module, now); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, separator)
}

func (r *Runner) render(module Module, now time.Time) string {
	switch module.Kind {
	case "text":
		return module.Value
	case "exec":
		return r.renderExec(module, now)
	case "launcher", "separator", "tray", "workspaces", "clock", "button", "layout_mode", "volume", "brightness", "microphone", "network", "bluetooth", "battery", "keyboard_layout", "storage", "ram", "cpu", "gpu", "netstat":
		return ""
	default:
		return ""
	}
}

func (r *Runner) renderExec(module Module, now time.Time) string {
	state := r.states[module.Name]
	if state.updated.IsZero() ||
		now.Sub(state.updated) >= time.Duration(module.Interval)*time.Second {
		state.text = runModuleCommand(module.Value)
		state.updated = now
		r.states[module.Name] = state
	}
	return state.text
}

func runModuleCommand(command string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "sh", "-c", command).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
