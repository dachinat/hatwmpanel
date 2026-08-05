package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseColor(t *testing.T) {
	for input, want := range map[string]uint32{
		"0xff33ccff": 0xff33ccff,
		"#112233":    0x112233ff,
	} {
		got, err := parseColor(input)
		if err != nil {
			t.Fatalf("parseColor(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("parseColor(%q) = %#x, want %#x", input, got, want)
		}
	}
}

func TestParseColorRejectsInvalidValues(t *testing.T) {
	for _, input := range []string{"", "#12345", "0x123456789", "not-a-color"} {
		if _, err := parseColor(input); err == nil {
			t.Fatalf("parseColor(%q) returned no error", input)
		}
	}
}

func TestParsePanelEffects(t *testing.T) {
	cfg := defaultConfig()
	parseSetting(&cfg, "panel_opacity", "0.85")
	parseSetting(&cfg, "border_width", "2")
	parseSetting(&cfg, "border_color", "0x89b4faff")
	parseSetting(&cfg, "shadow_size", "6")
	parseSetting(&cfg, "shadow_color", "0x11111b99")
	if cfg.PanelOpacity != 0.85 || cfg.BorderWidth != 2 ||
		cfg.BorderColor != 0x89b4faff || cfg.ShadowSize != 6 ||
		cfg.ShadowColor != 0x11111b99 {
		t.Fatalf("unexpected panel effects: %#v", cfg)
	}
}

func TestPanelOpacityRejectsOutOfRangeValues(t *testing.T) {
	cfg := defaultConfig()
	parseSetting(&cfg, "panel_opacity", "1.5")
	if cfg.PanelOpacity != 1 {
		t.Fatalf("out-of-range opacity was accepted: %v", cfg.PanelOpacity)
	}
}

func TestParseExecModule(t *testing.T) {
	m, ok := parseModule("kernel", "exec 60 uname -r")
	if !ok || m.Interval != 60 || m.Value != "uname -r" {
		t.Fatalf("unexpected module: %#v, ok=%v", m, ok)
	}
}

func TestLoadConfigCreatesDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Position != "top" || cfg.Height != 32 || len(cfg.Left) == 0 || len(cfg.Right) == 0 {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
	path := filepath.Join(home, ".config", "hatwmpanel", "config")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("default config not created: %v", err)
	}
}

func TestLoadConfigParsesKeyboardLayoutMappings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "hatwmpanel")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	contents := `[keyboard_layout_mappings]
US = English
ge = "ქართული"
blank =
`
	if err := os.WriteFile(filepath.Join(configDir, "config"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.KeyboardLayoutMappings["us"]; got != "English" {
		t.Fatalf("US mapping = %q, want English", got)
	}
	if got := cfg.KeyboardLayoutMappings["ge"]; got != "ქართული" {
		t.Fatalf("Georgian mapping = %q, want ქართული", got)
	}
	if _, ok := cfg.KeyboardLayoutMappings["blank"]; ok {
		t.Fatal("blank mapping should be ignored")
	}
}

func TestParseWorkspacesModule(t *testing.T) {
	m, ok := parseModule("pager", "workspaces")
	if !ok || m.Kind != "workspaces" {
		t.Fatalf("unexpected module: %#v, ok=%v", m, ok)
	}
}

func TestParseLauncherModule(t *testing.T) {
	m, ok := parseModule("menu", "launcher")
	if !ok || m.Kind != "launcher" {
		t.Fatalf("unexpected launcher module: %#v, ok=%v", m, ok)
	}
	m, ok = parseModule("menu",
		"launcher icon=start-here-symbolic icon_color=0x78a9ffff")
	if !ok || m.Icon != "start-here-symbolic" ||
		!m.HasIconColor || m.IconColor != 0x78a9ffff {
		t.Fatalf("unexpected custom launcher module: %#v, ok=%v", m, ok)
	}
	if _, ok := parseModule("menu", "launcher fill_color=0x315b86ff"); ok {
		t.Fatal("launcher should reject removed fill_color option")
	}
	if _, ok := parseModule("menu", "launcher unexpected"); ok {
		t.Fatal("launcher should reject unexpected options")
	}
}

func TestParseSeparatorModule(t *testing.T) {
	for _, spec := range []string{"separator width=12", "separator 12"} {
		m, ok := parseModule("gap", spec)
		if !ok || m.Kind != "separator" || m.Width != 12 {
			t.Fatalf("unexpected separator for %q: %#v, ok=%v", spec, m, ok)
		}
	}
	for _, spec := range []string{
		"separator", "separator width=0", "separator width=513",
		"separator height=12", "separator 12 extra",
	} {
		if _, ok := parseModule("gap", spec); ok {
			t.Fatalf("separator should reject %q", spec)
		}
	}
}

func TestParseTrayModule(t *testing.T) {
	m, ok := parseModule("systray", "tray")
	if !ok || m.Kind != "tray" {
		t.Fatalf("unexpected tray module: %#v, ok=%v", m, ok)
	}
}

func TestParseClockModule(t *testing.T) {
	m, ok := parseModule("clock", "clock %a %d %b  %H:%M")
	if !ok || m.Kind != "clock" || m.Value != "%a %d %b  %H:%M" ||
		m.Icon != "preferences-system-time-symbolic" {
		t.Fatalf("unexpected clock module: %#v, ok=%v", m, ok)
	}
}

func TestParseClockWithoutIcon(t *testing.T) {
	m, ok := parseModule("clock", "clock icon=none %H:%M")
	if !ok || m.Value != "%H:%M" || m.Icon != "" {
		t.Fatalf("unexpected icon-free clock module: %#v, ok=%v", m, ok)
	}
}

func TestParseClockCustomIcon(t *testing.T) {
	m, ok := parseModule("clock",
		"clock icon=appointment-soon-symbolic icon_color=0x78a9ffff %H:%M")
	if !ok || m.Value != "%H:%M" || m.Icon != "appointment-soon-symbolic" ||
		!m.HasIconColor || m.IconColor != 0x78a9ffff {
		t.Fatalf("unexpected custom-icon clock module: %#v, ok=%v", m, ok)
	}
}

func TestParseButtonModule(t *testing.T) {
	m, ok := parseModule("power",
		"button icon=system-shutdown-symbolic icon_color=0xee5396ff text=Power action=wlogout")
	if !ok || m.Kind != "button" || m.Icon != "system-shutdown-symbolic" ||
		m.Value != "Power" || m.Action != "wlogout" ||
		!m.HasIconColor || m.IconColor != 0xee5396ff {
		t.Fatalf("unexpected button module: %#v, ok=%v", m, ok)
	}
}

func TestParseIconOnlyButtonWithArguments(t *testing.T) {
	m, ok := parseModule("lock",
		"button icon=system-lock-screen-symbolic text=none action=loginctl lock-session")
	if !ok || m.Icon != "system-lock-screen-symbolic" || m.Value != "" ||
		m.Action != "loginctl lock-session" {
		t.Fatalf("unexpected icon-only button: %#v, ok=%v", m, ok)
	}
}

func TestParseLayoutModeModule(t *testing.T) {
	m, ok := parseModule("layout",
		"layout_mode tiling_icon=view-grid-symbolic floating_icon=window-restore-symbolic icon_color=0xc4a7e7ff")
	if !ok || m.Kind != "layout_mode" || m.TilingIcon != "view-grid-symbolic" ||
		m.FloatingIcon != "window-restore-symbolic" ||
		!m.HasIconColor || m.IconColor != 0xc4a7e7ff {
		t.Fatalf("unexpected layout-mode module: %#v, ok=%v", m, ok)
	}
	if _, ok := parseModule("layout", "layout_mode unknown=value"); ok {
		t.Fatal("layout mode should reject unknown options")
	}
}

func TestParseVolumeModule(t *testing.T) {
	m, ok := parseModule("volume",
		"volume icon=audio-volume-high-symbolic icon_color=0x78a9ffff step=3")
	if !ok || m.Kind != "volume" || m.Icon != "audio-volume-high-symbolic" ||
		m.Step != 3 || !m.HasIconColor || m.IconColor != 0x78a9ffff {
		t.Fatalf("unexpected volume module: %#v, ok=%v", m, ok)
	}
}

func TestParseBrightnessModule(t *testing.T) {
	m, ok := parseModule("brightness",
		"brightness icon=display-brightness-symbolic icon_color=0xeea846ff step=4")
	if !ok || m.Kind != "brightness" ||
		m.Icon != "display-brightness-symbolic" || m.Step != 4 ||
		!m.HasIconColor || m.IconColor != 0xeea846ff {
		t.Fatalf("unexpected brightness module: %#v, ok=%v", m, ok)
	}
}

func TestParseMicrophoneModule(t *testing.T) {
	m, ok := parseModule("microphone",
		"microphone icon=audio-input-microphone-symbolic icon_color=0xbe95ffff step=6")
	if !ok || m.Kind != "microphone" ||
		m.Icon != "audio-input-microphone-symbolic" || m.Step != 6 ||
		!m.HasIconColor || m.IconColor != 0xbe95ffff {
		t.Fatalf("unexpected microphone module: %#v, ok=%v", m, ok)
	}
}

func TestParseNetworkModule(t *testing.T) {
	m, ok := parseModule("wifi",
		"network icon=network-wireless-symbolic icon_color=0x25be6aff text=none")
	if !ok || m.Kind != "network" ||
		m.Icon != "network-wireless-symbolic" ||
		m.WiredIcon != "network-wired-symbolic" || m.Value != "none" ||
		!m.HasIconColor || m.IconColor != 0x25be6aff {
		t.Fatalf("unexpected network module: %#v, ok=%v", m, ok)
	}
}

func TestParseNetworkIconOverrides(t *testing.T) {
	m, ok := parseModule("network",
		"network icon=auto wireless_icon=network-cellular-symbolic wired_icon=network-vpn-symbolic")
	if !ok || m.Icon != "network-cellular-symbolic" ||
		m.WiredIcon != "network-vpn-symbolic" {
		t.Fatalf("unexpected network icons: %#v, ok=%v", m, ok)
	}

	m, ok = parseModule("network", "network icon=none text=on")
	if !ok || m.Icon != "" || m.WiredIcon != "" {
		t.Fatalf("icon=none did not disable both icons: %#v, ok=%v", m, ok)
	}
}

func TestParseBluetoothModule(t *testing.T) {
	m, ok := parseModule("bluetooth",
		"bluetooth icon=bluetooth-active-symbolic icon_color=0x8cb6ffff text=none")
	if !ok || m.Kind != "bluetooth" ||
		m.Icon != "bluetooth-active-symbolic" || m.Value != "none" ||
		!m.HasIconColor || m.IconColor != 0x8cb6ffff {
		t.Fatalf("unexpected Bluetooth module: %#v, ok=%v", m, ok)
	}
}

func TestParseBatteryModule(t *testing.T) {
	m, ok := parseModule("battery",
		"battery icon=auto icon_color=0x25be6aff text=none")
	if !ok || m.Kind != "battery" || m.Icon != "auto" || m.Value != "none" ||
		!m.HasIconColor || m.IconColor != 0x25be6aff {
		t.Fatalf("unexpected battery module: %#v, ok=%v", m, ok)
	}
}

func TestParseKeyboardLayoutModule(t *testing.T) {
	m, ok := parseModule("keyboard",
		"keyboard_layout icon=input-keyboard-symbolic icon_color=0x08bdbaff")
	if !ok || m.Kind != "keyboard_layout" ||
		m.Icon != "input-keyboard-symbolic" ||
		!m.HasIconColor || m.IconColor != 0x08bdbaff {
		t.Fatalf("unexpected keyboard-layout module: %#v, ok=%v", m, ok)
	}
}

func TestParseStorageModule(t *testing.T) {
	m, ok := parseModule("root",
		"storage path=/ icon=drive-harddisk-symbolic icon_color=0x78a9ffff")
	if !ok || m.Kind != "storage" || m.Value != "/" ||
		m.Icon != "drive-harddisk-symbolic" ||
		!m.HasIconColor || m.IconColor != 0x78a9ffff {
		t.Fatalf("unexpected storage module: %#v, ok=%v", m, ok)
	}
}

func TestParseRAMModule(t *testing.T) {
	m, ok := parseModule("memory",
		"ram icon=computer-symbolic icon_color=0xbe95ffff")
	if !ok || m.Kind != "ram" || m.Icon != "computer-symbolic" ||
		!m.HasIconColor || m.IconColor != 0xbe95ffff {
		t.Fatalf("unexpected RAM module: %#v, ok=%v", m, ok)
	}
}

func TestParseCPUModule(t *testing.T) {
	m, ok := parseModule("processor",
		"cpu icon=applications-system-symbolic icon_color=0x78a9ffff")
	if !ok || m.Kind != "cpu" || m.Icon != "applications-system-symbolic" ||
		!m.HasIconColor || m.IconColor != 0x78a9ffff {
		t.Fatalf("unexpected CPU module: %#v, ok=%v", m, ok)
	}
}

func TestParseGPUModule(t *testing.T) {
	m, ok := parseModule("graphics",
		"gpu icon=video-display-symbolic icon_color=0xbe95ffff")
	if !ok || m.Kind != "gpu" || m.Icon != "video-display-symbolic" ||
		!m.HasIconColor || m.IconColor != 0xbe95ffff {
		t.Fatalf("unexpected GPU module: %#v, ok=%v", m, ok)
	}
}

func TestParseNetstatModule(t *testing.T) {
	m, ok := parseModule("traffic",
		"netstat icon=network-transmit-receive-symbolic icon_color=0x33b1ffff")
	if !ok || m.Kind != "netstat" ||
		m.Icon != "network-transmit-receive-symbolic" ||
		!m.HasIconColor || m.IconColor != 0x33b1ffff {
		t.Fatalf("unexpected netstat module: %#v, ok=%v", m, ok)
	}
}

func TestWorkspaceDefaults(t *testing.T) {
	cfg := defaultConfig()
	if cfg.WorkspaceWidth != 24 || cfg.WorkspaceHeight != 22 ||
		cfg.WorkspaceGap != 5 || cfg.WorkspaceUrgentColor != 0xf07178ff ||
		cfg.BackgroundColor != 0x10151dff ||
		cfg.ModuleIconColor != 0x6aa9e9ff {
		t.Fatalf("unexpected workspace defaults: %#v", cfg)
	}
}
