package config

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Position                    string
	Layer                       string
	Height                      int
	ExclusiveZone               bool
	BackgroundColor             uint32
	ForegroundColor             uint32
	PanelOpacity                float64
	BorderWidth                 int
	BorderColor                 uint32
	ShadowSize                  int
	ShadowColor                 uint32
	Font                        string
	Padding                     int
	MarginTop                   int
	MarginRight                 int
	MarginBottom                int
	MarginLeft                  int
	Separator                   string
	WorkspaceWidth              int
	WorkspaceHeight             int
	WorkspaceGap                int
	WorkspaceInset              int
	WorkspaceColor              uint32
	WorkspaceActiveColor        uint32
	WorkspaceUrgentColor        uint32
	WorkspaceBorderColor        uint32
	WorkspaceWindowColor        uint32
	WorkspaceFocusedWindowColor uint32
	MenuBackgroundColor         uint32
	MenuForegroundColor         uint32
	MenuSelectedBackgroundColor uint32
	MenuSelectedForegroundColor uint32
	MenuBorderColor             uint32
	MenuFont                    string
	MenuPadding                 int
	MenuBorderWidth             int
	MenuRadius                  int
	ModuleBackgroundColor       uint32
	ModuleIconColor             uint32
	ModulePadding               int
	ModuleRadius                int
	ModuleIconSize              int
	VolumeSliderWidth           int
	BrightnessSliderWidth       int
	MicrophoneSliderWidth       int
	KeyboardLayoutMappings      map[string]string
	Left                        []Module
	Center                      []Module
	Right                       []Module
}

const defaultConfigFile = `# HatWM Panel configuration — HatWM Midnight
# The syntax intentionally follows ~/.config/hatwm/config.

[settings]
position = top
layer = top
height = 32
exclusive_zone = true
background_color = 0x10151dff
foreground_color = 0xe6edf6ff
panel_opacity = 1.0
border_width = 0
border_color = 0x34465fff
shadow_size = 0
shadow_color = 0x070a0f99
font = Iosevka Nerd Font SemiBold 10
padding = 10
separator = "  |  "
workspace_width = 24
workspace_height = 22
workspace_gap = 5
workspace_inset = 2
workspace_color = 0x202d3fff
workspace_active_color = 0x315b86ff
workspace_urgent_color = 0xf07178ff
workspace_border_color = 0x34465fff
workspace_window_color = 0x8b9bb0ff
workspace_focused_window_color = 0x67d4c0ff
menu_background_color = 0x10151dff
menu_foreground_color = 0xe6edf6ff
menu_selected_background_color = 0x315b86ff
menu_selected_foreground_color = 0xe6edf6ff
menu_border_color = 0x34465fff
menu_font = Iosevka Nerd Font 10
menu_padding = 6
menu_border_width = 1
menu_radius = 6
module_background_color = 0x182231ff
module_icon_color = 0x6aa9e9ff
module_padding = 7
module_radius = 6
module_icon_size = 16
volume_slider_width = 80
brightness_slider_width = 80
microphone_slider_width = 80
margin_top = 0
margin_right = 0
margin_bottom = 0
margin_left = 0

# Module syntax:
#   name = launcher [icon=auto|system-icon-name] [icon_color=RGBA]
#   name = separator width=PIXELS
#   name = text any text
#   name = clock [icon=name|none] [icon_color=RGBA] strftime-like-format
#   name = button [icon=name|none] [icon_color=RGBA] [text=label|none] action=command
#   name = layout_mode [tiling_icon=name] [floating_icon=name] [icon_color=RGBA]
#   name = volume [icon=name|none] [icon_color=RGBA] [step=1..25]
#   name = brightness [icon=name|none] [icon_color=RGBA] [step=1..25]
#   name = microphone [icon=name|none] [icon_color=RGBA] [step=1..25]
#   name = network [icon=auto|name|none] [wireless_icon=name|none] [wired_icon=name|none] [icon_color=RGBA] [text=on|none]
#   name = bluetooth [icon=name|none] [icon_color=RGBA] [text=on|none]
#   name = battery [icon=auto|name|none] [icon_color=RGBA] [text=on|none]
#   name = keyboard_layout [icon=name|none] [icon_color=RGBA]
#   name = storage [path=/] [icon=name|none] [icon_color=RGBA]
#   name = ram [icon=name|none] [icon_color=RGBA]
#   name = cpu [icon=name|none] [icon_color=RGBA]
#   name = gpu [icon=name|none] [icon_color=RGBA]
#   name = netstat [icon=name|none] [icon_color=RGBA]
#   name = exec [refresh_seconds] shell command
#   name = workspaces
#   name = tray
# Supported clock tokens include %a %A %b %B %d %m %Y %H %I %M %S %p %z.

# Optional display labels for XKB layout identifiers. Unmapped layouts use
# their uppercase two-character identifier (for example, us becomes US).
[keyboard_layout_mappings]
# us = EN
# ge = GE

[left]
launcher = launcher
workspaces = workspaces
layout = layout_mode
storage = storage path=/ icon=drive-harddisk-symbolic
ram = ram icon=computer-symbolic
cpu = cpu icon=applications-system-symbolic
gpu = gpu icon=video-display-symbolic
netstat = netstat icon=network-transmit-receive-symbolic
wm = text HatWM

[center]
host = exec 30 hostname

[right]
tray = tray
clock = clock icon=preferences-system-time-symbolic %a %d %b  %H:%M
power = button icon=system-shutdown-symbolic text=none action=wlogout
volume = volume icon=audio-volume-high-symbolic step=5
brightness = brightness icon=display-brightness-symbolic step=5
microphone = microphone icon=audio-input-microphone-symbolic step=5
network = network icon=auto text=on
bluetooth = bluetooth icon=bluetooth-active-symbolic text=on
battery = battery icon=auto icon_color=0x7bd88fff text=on
keyboard = keyboard_layout icon=input-keyboard-symbolic
`

func defaultConfig() Config {
	bg, _ := parseColor("0x10151dff")
	fg, _ := parseColor("0xe6edf6ff")
	workspace, _ := parseColor("0x202d3fff")
	workspaceActive, _ := parseColor("0x315b86ff")
	workspaceUrgent, _ := parseColor("0xf07178ff")
	workspaceBorder, _ := parseColor("0x34465fff")
	workspaceWindow, _ := parseColor("0x8b9bb0ff")
	workspaceFocusedWindow, _ := parseColor("0x67d4c0ff")
	moduleBackground, _ := parseColor("0x182231ff")
	moduleIcon, _ := parseColor("0x6aa9e9ff")
	return Config{
		Position:                    "top",
		Layer:                       "top",
		Height:                      32,
		ExclusiveZone:               true,
		BackgroundColor:             bg,
		ForegroundColor:             fg,
		PanelOpacity:                1,
		BorderColor:                 workspaceBorder,
		ShadowColor:                 0x070a0f99,
		Font:                        "Iosevka Nerd Font SemiBold 10",
		Padding:                     10,
		Separator:                   "  |  ",
		WorkspaceWidth:              24,
		WorkspaceHeight:             22,
		WorkspaceGap:                5,
		WorkspaceInset:              2,
		WorkspaceColor:              workspace,
		WorkspaceActiveColor:        workspaceActive,
		WorkspaceUrgentColor:        workspaceUrgent,
		WorkspaceBorderColor:        workspaceBorder,
		WorkspaceWindowColor:        workspaceWindow,
		WorkspaceFocusedWindowColor: workspaceFocusedWindow,
		MenuBackgroundColor:         bg,
		MenuForegroundColor:         fg,
		MenuSelectedBackgroundColor: workspaceActive,
		MenuSelectedForegroundColor: fg,
		MenuBorderColor:             workspaceBorder,
		MenuFont:                    "Iosevka Nerd Font 10",
		MenuPadding:                 6,
		MenuBorderWidth:             1,
		MenuRadius:                  6,
		ModuleBackgroundColor:       moduleBackground,
		ModuleIconColor:             moduleIcon,
		ModulePadding:               7,
		ModuleRadius:                6,
		ModuleIconSize:              16,
		VolumeSliderWidth:           80,
		BrightnessSliderWidth:       80,
		MicrophoneSliderWidth:       80,
		KeyboardLayoutMappings:      make(map[string]string),
	}
}

func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "hatwmpanel", "config")
}

func LoadConfig() (Config, error) {
	cfg := defaultConfig()
	path := GetConfigPath()
	if path == "" {
		return cfg, fmt.Errorf("could not determine user home directory")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return cfg, err
		}
		if err := os.WriteFile(path, []byte(defaultConfigFile), 0o644); err != nil {
			return cfg, err
		}
		slog.Info("created default config", "path", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	section := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}

		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch section {
		case "settings":
			parseSetting(&cfg, key, value)
		case "keyboard_layout_mappings":
			key = strings.ToLower(strings.TrimSpace(key))
			value = strings.TrimSpace(unquoteValue(value))
			if key != "" && value != "" {
				cfg.KeyboardLayoutMappings[key] = value
			}
		case "left":
			if m, ok := parseModule(key, value); ok {
				cfg.Left = append(cfg.Left, m)
			}
		case "center":
			if m, ok := parseModule(key, value); ok {
				cfg.Center = append(cfg.Center, m)
			}
		case "right":
			if m, ok := parseModule(key, value); ok {
				cfg.Right = append(cfg.Right, m)
			}
		}
	}
	return cfg, scanner.Err()
}

func parseSetting(cfg *Config, key, value string) {
	switch strings.ToLower(key) {
	case "position":
		v := strings.ToLower(value)
		if v == "top" || v == "bottom" {
			cfg.Position = v
		}
	case "layer":
		v := strings.ToLower(value)
		switch v {
		case "background", "bottom", "top", "overlay":
			cfg.Layer = v
		}
	case "height":
		if n, err := strconv.Atoi(value); err == nil && n >= 16 && n <= 256 {
			cfg.Height = n
		}
	case "exclusive_zone":
		if b, err := strconv.ParseBool(value); err == nil {
			cfg.ExclusiveZone = b
		}
	case "background_color":
		if c, err := parseColor(value); err == nil {
			cfg.BackgroundColor = c
		}
	case "foreground_color":
		if c, err := parseColor(value); err == nil {
			cfg.ForegroundColor = c
		}
	case "panel_opacity":
		cfg.PanelOpacity = parseRangeFloat(
			value, 0, 1, cfg.PanelOpacity)
	case "border_width":
		cfg.BorderWidth = parseRangeInt(
			value, 0, 16, cfg.BorderWidth)
	case "border_color":
		if c, err := parseColor(value); err == nil {
			cfg.BorderColor = c
		}
	case "shadow_size":
		cfg.ShadowSize = parseRangeInt(
			value, 0, 32, cfg.ShadowSize)
	case "shadow_color":
		if c, err := parseColor(value); err == nil {
			cfg.ShadowColor = c
		}
	case "font":
		if value != "" {
			cfg.Font = value
		}
	case "padding":
		if n, err := strconv.Atoi(value); err == nil && n >= 0 && n <= 128 {
			cfg.Padding = n
		}
	case "separator":
		cfg.Separator = unquoteValue(value)
	case "workspace_width":
		cfg.WorkspaceWidth = parseRangeInt(value, 12, 128, cfg.WorkspaceWidth)
	case "workspace_height":
		cfg.WorkspaceHeight = parseRangeInt(value, 12, 128, cfg.WorkspaceHeight)
	case "workspace_gap":
		cfg.WorkspaceGap = parseRangeInt(value, 0, 64, cfg.WorkspaceGap)
	case "workspace_inset":
		cfg.WorkspaceInset = parseRangeInt(value, 1, 16, cfg.WorkspaceInset)
	case "workspace_color":
		if c, err := parseColor(value); err == nil {
			cfg.WorkspaceColor = c
		}
	case "workspace_active_color":
		if c, err := parseColor(value); err == nil {
			cfg.WorkspaceActiveColor = c
		}
	case "workspace_urgent_color":
		if c, err := parseColor(value); err == nil {
			cfg.WorkspaceUrgentColor = c
		}
	case "workspace_border_color":
		if c, err := parseColor(value); err == nil {
			cfg.WorkspaceBorderColor = c
		}
	case "workspace_window_color":
		if c, err := parseColor(value); err == nil {
			cfg.WorkspaceWindowColor = c
		}
	case "workspace_focused_window_color":
		if c, err := parseColor(value); err == nil {
			cfg.WorkspaceFocusedWindowColor = c
		}
	case "menu_background_color":
		if c, err := parseColor(value); err == nil {
			cfg.MenuBackgroundColor = c
		}
	case "menu_foreground_color":
		if c, err := parseColor(value); err == nil {
			cfg.MenuForegroundColor = c
		}
	case "menu_selected_background_color":
		if c, err := parseColor(value); err == nil {
			cfg.MenuSelectedBackgroundColor = c
		}
	case "menu_selected_foreground_color":
		if c, err := parseColor(value); err == nil {
			cfg.MenuSelectedForegroundColor = c
		}
	case "menu_border_color":
		if c, err := parseColor(value); err == nil {
			cfg.MenuBorderColor = c
		}
	case "menu_font":
		if value != "" {
			cfg.MenuFont = unquoteValue(value)
		}
	case "menu_padding":
		cfg.MenuPadding = parseRangeInt(value, 0, 64, cfg.MenuPadding)
	case "menu_border_width":
		cfg.MenuBorderWidth = parseRangeInt(value, 0, 16, cfg.MenuBorderWidth)
	case "menu_radius":
		cfg.MenuRadius = parseRangeInt(value, 0, 64, cfg.MenuRadius)
	case "module_background_color":
		if c, err := parseColor(value); err == nil {
			cfg.ModuleBackgroundColor = c
		}
	case "module_icon_color":
		if c, err := parseColor(value); err == nil {
			cfg.ModuleIconColor = c
		}
	case "module_padding":
		cfg.ModulePadding = parseRangeInt(value, 0, 64, cfg.ModulePadding)
	case "module_radius":
		cfg.ModuleRadius = parseRangeInt(value, 0, 64, cfg.ModuleRadius)
	case "module_icon_size":
		cfg.ModuleIconSize = parseRangeInt(value, 8, 64, cfg.ModuleIconSize)
	case "volume_slider_width":
		cfg.VolumeSliderWidth = parseRangeInt(value, 40, 300, cfg.VolumeSliderWidth)
	case "brightness_slider_width":
		cfg.BrightnessSliderWidth = parseRangeInt(value, 40, 300, cfg.BrightnessSliderWidth)
	case "microphone_slider_width":
		cfg.MicrophoneSliderWidth = parseRangeInt(value, 40, 300, cfg.MicrophoneSliderWidth)
	case "margin_top":
		cfg.MarginTop = parseRangeInt(value, 0, 256, cfg.MarginTop)
	case "margin_right":
		cfg.MarginRight = parseRangeInt(value, 0, 256, cfg.MarginRight)
	case "margin_bottom":
		cfg.MarginBottom = parseRangeInt(value, 0, 256, cfg.MarginBottom)
	case "margin_left":
		cfg.MarginLeft = parseRangeInt(value, 0, 256, cfg.MarginLeft)
	}
}

func unquoteValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func parseRangeInt(value string, min, max, fallback int) int {
	n, err := strconv.Atoi(value)
	if err != nil || n < min || n > max {
		return fallback
	}
	return n
}

func parseRangeFloat(value string, min, max, fallback float64) float64 {
	n, err := strconv.ParseFloat(value, 64)
	if err != nil || n < min || n > max {
		return fallback
	}
	return n
}

func parseModule(name, spec string) (Module, bool) {
	fields := strings.Fields(spec)
	if len(fields) == 0 {
		return Module{}, false
	}

	kind := strings.ToLower(fields[0])
	rest := strings.TrimSpace(strings.TrimPrefix(spec, fields[0]))
	m := Module{Name: name, Kind: kind, Interval: 5}

	switch kind {
	case "separator":
		fields := strings.Fields(rest)
		if len(fields) != 1 {
			return Module{}, false
		}
		width, err := parseSeparatorWidth(fields[0])
		if err != nil {
			return Module{}, false
		}
		m.Width = width
		return m, true
	case "text":
		m.Value = rest
		return m, true
	case "clock":
		m.Icon = "preferences-system-time-symbolic"
		for {
			options := strings.Fields(rest)
			if len(options) == 0 {
				break
			}
			option := options[0]
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				break
			}
			switch strings.ToLower(keyValue[0]) {
			case "icon":
				m.Icon = strings.TrimSpace(keyValue[1])
				if strings.EqualFold(m.Icon, "none") || strings.EqualFold(m.Icon, "off") {
					m.Icon = ""
				}
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			default:
				m.Value = rest
				return m, true
			}
			rest = strings.TrimSpace(strings.TrimPrefix(rest, option))
		}
		m.Value = rest
		return m, true
	case "button":
		actionAt := strings.Index(strings.ToLower(rest), "action=")
		if actionAt < 0 {
			return Module{}, false
		}
		options := strings.Fields(strings.TrimSpace(rest[:actionAt]))
		m.Action = strings.TrimSpace(rest[actionAt+len("action="):])
		for _, option := range options {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			switch strings.ToLower(keyValue[0]) {
			case "icon":
				m.Icon = unquoteValue(keyValue[1])
				if strings.EqualFold(m.Icon, "none") || strings.EqualFold(m.Icon, "off") {
					m.Icon = ""
				}
			case "text":
				m.Value = unquoteValue(keyValue[1])
				if strings.EqualFold(m.Value, "none") || strings.EqualFold(m.Value, "off") {
					m.Value = ""
				}
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			}
		}
		return m, m.Action != "" && (m.Icon != "" || m.Value != "")
	case "layout_mode":
		m.TilingIcon = "view-grid-symbolic"
		m.FloatingIcon = "window-restore-symbolic"
		for _, option := range strings.Fields(rest) {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				return Module{}, false
			}
			switch strings.ToLower(keyValue[0]) {
			case "tiling_icon":
				m.TilingIcon = unquoteValue(keyValue[1])
			case "floating_icon":
				m.FloatingIcon = unquoteValue(keyValue[1])
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			default:
				return Module{}, false
			}
		}
		m.Icon = m.TilingIcon
		return m, m.TilingIcon != "" && m.FloatingIcon != ""
	case "volume":
		m.Icon = "audio-volume-high-symbolic"
		m.Step = 5
		for _, option := range strings.Fields(rest) {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			switch strings.ToLower(keyValue[0]) {
			case "icon":
				m.Icon = unquoteValue(keyValue[1])
				if strings.EqualFold(m.Icon, "none") || strings.EqualFold(m.Icon, "off") {
					m.Icon = ""
				}
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			case "step":
				step, err := strconv.Atoi(keyValue[1])
				if err != nil || step < 1 || step > 25 {
					return Module{}, false
				}
				m.Step = step
			}
		}
		return m, true
	case "brightness":
		m.Icon = "display-brightness-symbolic"
		m.Step = 5
		for _, option := range strings.Fields(rest) {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			switch strings.ToLower(keyValue[0]) {
			case "icon":
				m.Icon = unquoteValue(keyValue[1])
				if strings.EqualFold(m.Icon, "none") || strings.EqualFold(m.Icon, "off") {
					m.Icon = ""
				}
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			case "step":
				step, err := strconv.Atoi(keyValue[1])
				if err != nil || step < 1 || step > 25 {
					return Module{}, false
				}
				m.Step = step
			}
		}
		return m, true
	case "microphone":
		m.Icon = "audio-input-microphone-symbolic"
		m.Step = 5
		for _, option := range strings.Fields(rest) {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			switch strings.ToLower(keyValue[0]) {
			case "icon":
				m.Icon = unquoteValue(keyValue[1])
				if strings.EqualFold(m.Icon, "none") || strings.EqualFold(m.Icon, "off") {
					m.Icon = ""
				}
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			case "step":
				step, err := strconv.Atoi(keyValue[1])
				if err != nil || step < 1 || step > 25 {
					return Module{}, false
				}
				m.Step = step
			}
		}
		return m, true
	case "network":
		m.Icon = "network-wireless-symbolic"
		m.WiredIcon = "network-wired-symbolic"
		m.Value = "on"
		for _, option := range strings.Fields(rest) {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			switch strings.ToLower(keyValue[0]) {
			case "icon":
				icon := unquoteValue(keyValue[1])
				if strings.EqualFold(icon, "auto") {
					m.Icon = "network-wireless-symbolic"
					m.WiredIcon = "network-wired-symbolic"
				} else if strings.EqualFold(icon, "none") || strings.EqualFold(icon, "off") {
					m.Icon = ""
					m.WiredIcon = ""
				} else {
					m.Icon = icon
				}
			case "wireless_icon":
				m.Icon = networkIconOption(
					keyValue[1], "network-wireless-symbolic")
			case "wired_icon":
				m.WiredIcon = networkIconOption(
					keyValue[1], "network-wired-symbolic")
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			case "text":
				m.Value = strings.ToLower(unquoteValue(keyValue[1]))
			}
		}
		return m, m.Icon != "" || m.WiredIcon != "" ||
			(m.Value != "none" && m.Value != "off")
	case "bluetooth":
		m.Icon = "bluetooth-active-symbolic"
		m.Value = "on"
		for _, option := range strings.Fields(rest) {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			switch strings.ToLower(keyValue[0]) {
			case "icon":
				m.Icon = unquoteValue(keyValue[1])
				if strings.EqualFold(m.Icon, "none") || strings.EqualFold(m.Icon, "off") {
					m.Icon = ""
				}
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			case "text":
				m.Value = strings.ToLower(unquoteValue(keyValue[1]))
			}
		}
		return m, m.Icon != "" || (m.Value != "none" && m.Value != "off")
	case "battery":
		m.Icon = "auto"
		m.Value = "on"
		for _, option := range strings.Fields(rest) {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			switch strings.ToLower(keyValue[0]) {
			case "icon":
				m.Icon = unquoteValue(keyValue[1])
				if strings.EqualFold(m.Icon, "none") || strings.EqualFold(m.Icon, "off") {
					m.Icon = ""
				}
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			case "text":
				m.Value = strings.ToLower(unquoteValue(keyValue[1]))
			}
		}
		return m, m.Icon != "" || (m.Value != "none" && m.Value != "off")
	case "keyboard_layout":
		m.Icon = "input-keyboard-symbolic"
		for _, option := range strings.Fields(rest) {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			switch strings.ToLower(keyValue[0]) {
			case "icon":
				m.Icon = unquoteValue(keyValue[1])
				if strings.EqualFold(m.Icon, "none") || strings.EqualFold(m.Icon, "off") {
					m.Icon = ""
				}
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			}
		}
		return m, true
	case "storage":
		m.Icon = "drive-harddisk-symbolic"
		m.Value = "/"
		for _, option := range strings.Fields(rest) {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			switch strings.ToLower(keyValue[0]) {
			case "path":
				m.Value = unquoteValue(keyValue[1])
				if m.Value == "" {
					return Module{}, false
				}
			case "icon":
				m.Icon = unquoteValue(keyValue[1])
				if strings.EqualFold(m.Icon, "none") || strings.EqualFold(m.Icon, "off") {
					m.Icon = ""
				}
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			}
		}
		return m, true
	case "ram":
		m.Icon = "computer-symbolic"
		for _, option := range strings.Fields(rest) {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			switch strings.ToLower(keyValue[0]) {
			case "icon":
				m.Icon = unquoteValue(keyValue[1])
				if strings.EqualFold(m.Icon, "none") || strings.EqualFold(m.Icon, "off") {
					m.Icon = ""
				}
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			}
		}
		return m, true
	case "cpu", "gpu", "netstat":
		if kind == "cpu" {
			m.Icon = "applications-system-symbolic"
		} else if kind == "gpu" {
			m.Icon = "video-display-symbolic"
		} else {
			m.Icon = "network-transmit-receive-symbolic"
		}
		for _, option := range strings.Fields(rest) {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			switch strings.ToLower(keyValue[0]) {
			case "icon":
				m.Icon = unquoteValue(keyValue[1])
				if strings.EqualFold(m.Icon, "none") || strings.EqualFold(m.Icon, "off") {
					m.Icon = ""
				}
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			}
		}
		return m, true
	case "launcher":
		for _, option := range strings.Fields(rest) {
			keyValue := strings.SplitN(option, "=", 2)
			if len(keyValue) != 2 {
				return Module{}, false
			}
			switch strings.ToLower(keyValue[0]) {
			case "icon":
				icon := unquoteValue(keyValue[1])
				if icon == "" {
					return Module{}, false
				}
				if strings.EqualFold(icon, "auto") {
					m.Icon = ""
				} else {
					m.Icon = icon
				}
			case "icon_color":
				color, err := parseColor(keyValue[1])
				if err != nil {
					return Module{}, false
				}
				m.IconColor = color
				m.HasIconColor = true
			default:
				return Module{}, false
			}
		}
		return m, true
	case "workspaces", "tray":
		if strings.TrimSpace(rest) != "" {
			return Module{}, false
		}
		return m, true
	case "exec":
		if rest == "" {
			return Module{}, false
		}
		execFields := strings.Fields(rest)
		if len(execFields) > 1 {
			if n, err := strconv.Atoi(execFields[0]); err == nil && n >= 1 && n <= 86400 {
				m.Interval = n
				m.Value = strings.TrimSpace(strings.TrimPrefix(rest, execFields[0]))
				return m, m.Value != ""
			}
		}
		m.Value = rest
		return m, true
	default:
		slog.Warn("ignoring unknown panel module", "name", name, "kind", kind)
		return Module{}, false
	}
}

func networkIconOption(value, automatic string) string {
	icon := unquoteValue(value)
	if strings.EqualFold(icon, "auto") {
		return automatic
	}
	if strings.EqualFold(icon, "none") || strings.EqualFold(icon, "off") {
		return ""
	}
	return icon
}

func parseColor(value string) (uint32, error) {
	s := strings.TrimSpace(value)
	s = strings.TrimPrefix(s, "#")
	s = strings.TrimPrefix(strings.ToLower(s), "0x")
	if len(s) == 6 {
		s += "ff"
	}
	if len(s) != 8 {
		return 0, fmt.Errorf("color must be RRGGBB or RRGGBBAA")
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}
