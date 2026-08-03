package native

import "testing"

func TestModuleByKindFindsGroupAndConfiguration(t *testing.T) {
	cfg := Config{
		Left:   []Module{{Name: "launcher", Kind: "launcher"}},
		Center: []Module{{Name: "clock", Kind: "clock", Value: "%H:%M"}},
		Right:  []Module{{Name: "cpu", Kind: "cpu"}},
	}

	group, module, ok := moduleByKind(cfg, "clock")
	if !ok || group != 2 || module.Name != "clock" || module.Value != "%H:%M" {
		t.Fatalf("moduleByKind(clock) = %d, %#v, %v", group, module, ok)
	}
	group, module, ok = moduleByKind(cfg, "missing")
	if ok || group != 0 || module != (Module{}) {
		t.Fatalf("moduleByKind(missing) = %d, %#v, %v", group, module, ok)
	}
}

func TestWorkspaceAndTrayModuleGroups(t *testing.T) {
	cfg := Config{
		Center: []Module{{Name: "workspaces", Kind: "workspaces"}},
		Right:  []Module{{Name: "tray", Kind: "tray"}},
	}
	if got := workspaceModuleGroup(cfg); got != 2 {
		t.Fatalf("workspaceModuleGroup() = %d, want 2", got)
	}
	if got := trayModuleGroup(cfg); got != 3 {
		t.Fatalf("trayModuleGroup() = %d, want 3", got)
	}
	if got := workspaceModuleGroup(Config{}); got != 0 {
		t.Fatalf("empty workspaceModuleGroup() = %d, want 0", got)
	}
}
