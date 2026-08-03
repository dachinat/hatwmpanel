package main

import (
	"testing"
	"time"
)

func TestRunnerRendersOnlyTextualModules(t *testing.T) {
	runner := NewRunner()
	modules := []Module{
		{Name: "first", Kind: "text", Value: "HatWM"},
		{Name: "workspaces", Kind: "workspaces"},
		{Name: "empty", Kind: "text"},
		{Name: "second", Kind: "text", Value: "Panel"},
		{Name: "unknown", Kind: "unsupported", Value: "ignored"},
	}

	if got, want := runner.Render(modules, time.Unix(0, 0), " | "), "HatWM | Panel"; got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRunnerCachesExecOutputUntilIntervalExpires(t *testing.T) {
	runner := NewRunner()
	started := time.Unix(100, 0)
	module := Module{
		Name: "command", Kind: "exec", Value: "printf 'first'", Interval: 10,
	}

	if got := runner.Render([]Module{module}, started, ""); got != "first" {
		t.Fatalf("initial Render() = %q, want first", got)
	}

	module.Value = "printf 'second'"
	if got := runner.Render([]Module{module}, started.Add(9*time.Second), ""); got != "first" {
		t.Fatalf("cached Render() = %q, want first", got)
	}
	if got := runner.Render([]Module{module}, started.Add(10*time.Second), ""); got != "second" {
		t.Fatalf("refreshed Render() = %q, want second", got)
	}
}

func TestRunnerKeepsIndependentExecCaches(t *testing.T) {
	runner := NewRunner()
	now := time.Unix(200, 0)
	modules := []Module{
		{Name: "first", Kind: "exec", Value: "printf A", Interval: 60},
		{Name: "second", Kind: "exec", Value: "printf B", Interval: 60},
	}
	if got := runner.Render(modules, now, ":"); got != "A:B" {
		t.Fatalf("initial Render() = %q, want A:B", got)
	}

	modules[0].Value = "printf changed"
	modules[1].Name = "third"
	modules[1].Value = "printf C"
	if got := runner.Render(modules, now.Add(time.Second), ":"); got != "A:C" {
		t.Fatalf("independent cache Render() = %q, want A:C", got)
	}
}

func TestRunModuleCommandTrimsOutputAndHidesFailures(t *testing.T) {
	if got := runModuleCommand("printf '  ready  \\n'"); got != "ready" {
		t.Fatalf("runModuleCommand() = %q, want ready", got)
	}
	if got := runModuleCommand("exit 7"); got != "" {
		t.Fatalf("failed command returned %q, want empty output", got)
	}
}
