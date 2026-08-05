package native

/*
#cgo pkg-config: wayland-client cairo pangocairo librsvg-2.0 gtk+-3.0 gtk-layer-shell-0 dbusmenu-gtk3-0.4
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include <stdlib.h>
#include "panel.h"
*/
import "C"

import (
	"fmt"
	configpkg "hatwmpanel/config"
	traymodule "hatwmpanel/internal/desktop/tray"
	"strings"
	"time"
	"unsafe"
)

type Config = configpkg.Config
type Module = configpkg.Module

type Panel struct {
	ptr       *C.HatWMPanel
	trayItems []traymodule.TrayItem
	buttons   []Module
}

func iconTintValue(module Module, icon string) int {
	if module.HasIconColor ||
		strings.HasSuffix(strings.ToLower(icon), "-symbolic") {
		return 1
	}
	return 0
}

func launcherModule(cfg Config) (int, Module, bool) {
	return moduleByKind(cfg, "launcher")
}

func workspaceModuleGroup(cfg Config) int {
	for _, module := range cfg.Left {
		if module.Kind == "workspaces" {
			return int(C.HATWM_PANEL_GROUP_LEFT)
		}
	}
	for _, module := range cfg.Center {
		if module.Kind == "workspaces" {
			return int(C.HATWM_PANEL_GROUP_CENTER)
		}
	}
	for _, module := range cfg.Right {
		if module.Kind == "workspaces" {
			return int(C.HATWM_PANEL_GROUP_RIGHT)
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE)
}

func trayModuleGroup(cfg Config) int {
	groups := []struct {
		modules []Module
		value   C.int
	}{
		{cfg.Left, C.HATWM_PANEL_GROUP_LEFT},
		{cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT},
	}
	for _, group := range groups {
		for _, module := range group.modules {
			if module.Kind == "tray" {
				return int(group.value)
			}
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE)
}

func clockModule(cfg Config) (int, Module, bool) {
	groups := []struct {
		modules []Module
		value   C.int
	}{{cfg.Left, C.HATWM_PANEL_GROUP_LEFT}, {cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT}}
	for _, group := range groups {
		for _, module := range group.modules {
			if module.Kind == "clock" {
				return int(group.value), module, true
			}
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE), Module{}, false
}

func volumeModule(cfg Config) (int, Module, bool) {
	groups := []struct {
		modules []Module
		value   C.int
	}{{cfg.Left, C.HATWM_PANEL_GROUP_LEFT}, {cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT}}
	for _, group := range groups {
		for _, module := range group.modules {
			if module.Kind == "volume" {
				return int(group.value), module, true
			}
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE), Module{}, false
}

func brightnessModule(cfg Config) (int, Module, bool) {
	groups := []struct {
		modules []Module
		value   C.int
	}{{cfg.Left, C.HATWM_PANEL_GROUP_LEFT}, {cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT}}
	for _, group := range groups {
		for _, module := range group.modules {
			if module.Kind == "brightness" {
				return int(group.value), module, true
			}
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE), Module{}, false
}

func microphoneModule(cfg Config) (int, Module, bool) {
	groups := []struct {
		modules []Module
		value   C.int
	}{{cfg.Left, C.HATWM_PANEL_GROUP_LEFT}, {cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT}}
	for _, group := range groups {
		for _, module := range group.modules {
			if module.Kind == "microphone" {
				return int(group.value), module, true
			}
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE), Module{}, false
}

func networkModule(cfg Config) (int, Module, bool) {
	groups := []struct {
		modules []Module
		value   C.int
	}{{cfg.Left, C.HATWM_PANEL_GROUP_LEFT}, {cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT}}
	for _, group := range groups {
		for _, module := range group.modules {
			if module.Kind == "network" {
				return int(group.value), module, true
			}
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE), Module{}, false
}

func bluetoothModule(cfg Config) (int, Module, bool) {
	groups := []struct {
		modules []Module
		value   C.int
	}{{cfg.Left, C.HATWM_PANEL_GROUP_LEFT}, {cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT}}
	for _, group := range groups {
		for _, module := range group.modules {
			if module.Kind == "bluetooth" {
				return int(group.value), module, true
			}
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE), Module{}, false
}

func batteryModule(cfg Config) (int, Module, bool) {
	groups := []struct {
		modules []Module
		value   C.int
	}{{cfg.Left, C.HATWM_PANEL_GROUP_LEFT}, {cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT}}
	for _, group := range groups {
		for _, module := range group.modules {
			if module.Kind == "battery" {
				return int(group.value), module, true
			}
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE), Module{}, false
}

func keyboardLayoutModule(cfg Config) (int, Module, bool) {
	groups := []struct {
		modules []Module
		value   C.int
	}{{cfg.Left, C.HATWM_PANEL_GROUP_LEFT}, {cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT}}
	for _, group := range groups {
		for _, module := range group.modules {
			if module.Kind == "keyboard_layout" {
				return int(group.value), module, true
			}
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE), Module{}, false
}

func storageModule(cfg Config) (int, Module, bool) {
	groups := []struct {
		modules []Module
		value   C.int
	}{{cfg.Left, C.HATWM_PANEL_GROUP_LEFT}, {cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT}}
	for _, group := range groups {
		for _, module := range group.modules {
			if module.Kind == "storage" {
				return int(group.value), module, true
			}
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE), Module{}, false
}

func ramModule(cfg Config) (int, Module, bool) {
	groups := []struct {
		modules []Module
		value   C.int
	}{{cfg.Left, C.HATWM_PANEL_GROUP_LEFT}, {cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT}}
	for _, group := range groups {
		for _, module := range group.modules {
			if module.Kind == "ram" {
				return int(group.value), module, true
			}
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE), Module{}, false
}

func cpuModule(cfg Config) (int, Module, bool) {
	return moduleByKind(cfg, "cpu")
}

func gpuModule(cfg Config) (int, Module, bool) {
	return moduleByKind(cfg, "gpu")
}

func netstatModule(cfg Config) (int, Module, bool) {
	return moduleByKind(cfg, "netstat")
}

func moduleByKind(cfg Config, kind string) (int, Module, bool) {
	groups := []struct {
		modules []Module
		value   C.int
	}{{cfg.Left, C.HATWM_PANEL_GROUP_LEFT}, {cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT}}
	for _, group := range groups {
		for _, module := range group.modules {
			if module.Kind == kind {
				return int(group.value), module, true
			}
		}
	}
	return int(C.HATWM_PANEL_GROUP_NONE), Module{}, false
}

func NewPanel(cfg Config) (*Panel, error) {
	position := C.CString(cfg.Position)
	layer := C.CString(cfg.Layer)
	font := C.CString(cfg.Font)
	separator := C.CString(cfg.Separator)
	menuFont := C.CString(cfg.MenuFont)
	launcherGroup, launcherCfg, _ := launcherModule(cfg)
	launcherIcon := C.CString(launcherCfg.Icon)
	launcherIconColor := cfg.ModuleIconColor
	if launcherCfg.HasIconColor {
		launcherIconColor = launcherCfg.IconColor
	}
	launcherIconTint := iconTintValue(launcherCfg, launcherCfg.Icon)
	_, clockCfg, _ := clockModule(cfg)
	clockIcon := C.CString(clockCfg.Icon)
	clockIconColor := cfg.ModuleIconColor
	if clockCfg.HasIconColor {
		clockIconColor = clockCfg.IconColor
	}
	volumeGroup, volumeCfg, _ := volumeModule(cfg)
	volumeIcon := C.CString(volumeCfg.Icon)
	volumeIconColor := cfg.ModuleIconColor
	if volumeCfg.HasIconColor {
		volumeIconColor = volumeCfg.IconColor
	}
	brightnessGroup, brightnessCfg, _ := brightnessModule(cfg)
	brightnessIcon := C.CString(brightnessCfg.Icon)
	brightnessIconColor := cfg.ModuleIconColor
	if brightnessCfg.HasIconColor {
		brightnessIconColor = brightnessCfg.IconColor
	}
	microphoneGroup, microphoneCfg, _ := microphoneModule(cfg)
	microphoneIcon := C.CString(microphoneCfg.Icon)
	microphoneIconColor := cfg.ModuleIconColor
	if microphoneCfg.HasIconColor {
		microphoneIconColor = microphoneCfg.IconColor
	}
	networkGroup, networkCfg, _ := networkModule(cfg)
	networkIcon := C.CString(networkCfg.Icon)
	networkWiredIcon := C.CString(networkCfg.WiredIcon)
	networkIconColor := cfg.ModuleIconColor
	if networkCfg.HasIconColor {
		networkIconColor = networkCfg.IconColor
	}
	networkShowText := 1
	if networkCfg.Value == "none" || networkCfg.Value == "off" {
		networkShowText = 0
	}
	bluetoothGroup, bluetoothCfg, _ := bluetoothModule(cfg)
	bluetoothIcon := C.CString(bluetoothCfg.Icon)
	bluetoothIconColor := cfg.ModuleIconColor
	if bluetoothCfg.HasIconColor {
		bluetoothIconColor = bluetoothCfg.IconColor
	}
	bluetoothShowText := 1
	if bluetoothCfg.Value == "none" || bluetoothCfg.Value == "off" {
		bluetoothShowText = 0
	}
	batteryGroup, batteryCfg, _ := batteryModule(cfg)
	batteryDynamicIcon := 0
	batteryIconName := batteryCfg.Icon
	if batteryIconName == "auto" {
		batteryDynamicIcon = 1
		batteryIconName = "battery-missing-symbolic"
	}
	batteryIcon := C.CString(batteryIconName)
	batteryIconColor := cfg.ModuleIconColor
	if batteryCfg.HasIconColor {
		batteryIconColor = batteryCfg.IconColor
	}
	batteryShowText := 1
	if batteryCfg.Value == "none" || batteryCfg.Value == "off" {
		batteryShowText = 0
	}
	keyboardLayoutGroup, keyboardLayoutCfg, _ := keyboardLayoutModule(cfg)
	keyboardLayoutIcon := C.CString(keyboardLayoutCfg.Icon)
	keyboardLayoutIconColor := cfg.ModuleIconColor
	if keyboardLayoutCfg.HasIconColor {
		keyboardLayoutIconColor = keyboardLayoutCfg.IconColor
	}
	storageGroup, storageCfg, _ := storageModule(cfg)
	storageIcon := C.CString(storageCfg.Icon)
	storageIconColor := cfg.ModuleIconColor
	if storageCfg.HasIconColor {
		storageIconColor = storageCfg.IconColor
	}
	ramGroup, ramCfg, _ := ramModule(cfg)
	ramIcon := C.CString(ramCfg.Icon)
	ramIconColor := cfg.ModuleIconColor
	if ramCfg.HasIconColor {
		ramIconColor = ramCfg.IconColor
	}
	cpuGroup, cpuCfg, _ := cpuModule(cfg)
	cpuIcon := C.CString(cpuCfg.Icon)
	cpuIconColor := cfg.ModuleIconColor
	if cpuCfg.HasIconColor {
		cpuIconColor = cpuCfg.IconColor
	}
	gpuGroup, gpuCfg, _ := gpuModule(cfg)
	gpuIcon := C.CString(gpuCfg.Icon)
	gpuIconColor := cfg.ModuleIconColor
	if gpuCfg.HasIconColor {
		gpuIconColor = gpuCfg.IconColor
	}
	netstatGroup, netstatCfg, _ := netstatModule(cfg)
	netstatIcon := C.CString(netstatCfg.Icon)
	netstatIconColor := cfg.ModuleIconColor
	if netstatCfg.HasIconColor {
		netstatIconColor = netstatCfg.IconColor
	}
	defer C.free(unsafe.Pointer(position))
	defer C.free(unsafe.Pointer(layer))
	defer C.free(unsafe.Pointer(font))
	defer C.free(unsafe.Pointer(separator))
	defer C.free(unsafe.Pointer(menuFont))
	defer C.free(unsafe.Pointer(launcherIcon))
	defer C.free(unsafe.Pointer(clockIcon))
	defer C.free(unsafe.Pointer(volumeIcon))
	defer C.free(unsafe.Pointer(brightnessIcon))
	defer C.free(unsafe.Pointer(microphoneIcon))
	defer C.free(unsafe.Pointer(networkIcon))
	defer C.free(unsafe.Pointer(networkWiredIcon))
	defer C.free(unsafe.Pointer(bluetoothIcon))
	defer C.free(unsafe.Pointer(batteryIcon))
	defer C.free(unsafe.Pointer(keyboardLayoutIcon))
	defer C.free(unsafe.Pointer(storageIcon))
	defer C.free(unsafe.Pointer(ramIcon))
	defer C.free(unsafe.Pointer(cpuIcon))
	defer C.free(unsafe.Pointer(gpuIcon))
	defer C.free(unsafe.Pointer(netstatIcon))

	exclusive := 0
	if cfg.ExclusiveZone {
		exclusive = 1
	}
	ccfg := C.HatWMPanelConfig{
		position:                       position,
		layer:                          layer,
		height:                         C.int(cfg.Height),
		exclusive_zone:                 C.int(exclusive),
		background_color:               C.uint32_t(cfg.BackgroundColor),
		foreground_color:               C.uint32_t(cfg.ForegroundColor),
		panel_opacity:                  C.double(cfg.PanelOpacity),
		border_width:                   C.int(cfg.BorderWidth),
		border_color:                   C.uint32_t(cfg.BorderColor),
		shadow_size:                    C.int(cfg.ShadowSize),
		shadow_color:                   C.uint32_t(cfg.ShadowColor),
		font:                           font,
		padding:                        C.int(cfg.Padding),
		margin_top:                     C.int(cfg.MarginTop),
		margin_right:                   C.int(cfg.MarginRight),
		margin_bottom:                  C.int(cfg.MarginBottom),
		margin_left:                    C.int(cfg.MarginLeft),
		separator:                      separator,
		launcher_group:                 C.int(launcherGroup),
		launcher_icon:                  launcherIcon,
		launcher_icon_color:            C.uint32_t(launcherIconColor),
		launcher_icon_tint:             C.int(launcherIconTint),
		workspace_group:                C.int(workspaceModuleGroup(cfg)),
		tray_group:                     C.int(trayModuleGroup(cfg)),
		clock_group:                    C.int(func() int { group, _, _ := clockModule(cfg); return group }()),
		clock_icon:                     clockIcon,
		clock_icon_color:               C.uint32_t(clockIconColor),
		clock_icon_tint:                C.int(iconTintValue(clockCfg, clockCfg.Icon)),
		volume_group:                   C.int(volumeGroup),
		volume_icon:                    volumeIcon,
		volume_icon_color:              C.uint32_t(volumeIconColor),
		volume_icon_tint:               C.int(iconTintValue(volumeCfg, volumeCfg.Icon)),
		volume_slider_width:            C.int(cfg.VolumeSliderWidth),
		volume_step:                    C.int(volumeCfg.Step),
		brightness_group:               C.int(brightnessGroup),
		brightness_icon:                brightnessIcon,
		brightness_icon_color:          C.uint32_t(brightnessIconColor),
		brightness_icon_tint:           C.int(iconTintValue(brightnessCfg, brightnessCfg.Icon)),
		brightness_slider_width:        C.int(cfg.BrightnessSliderWidth),
		brightness_step:                C.int(brightnessCfg.Step),
		microphone_group:               C.int(microphoneGroup),
		microphone_icon:                microphoneIcon,
		microphone_icon_color:          C.uint32_t(microphoneIconColor),
		microphone_icon_tint:           C.int(iconTintValue(microphoneCfg, microphoneCfg.Icon)),
		microphone_slider_width:        C.int(cfg.MicrophoneSliderWidth),
		microphone_step:                C.int(microphoneCfg.Step),
		network_group:                  C.int(networkGroup),
		network_icon:                   networkIcon,
		network_wired_icon:             networkWiredIcon,
		network_icon_color:             C.uint32_t(networkIconColor),
		network_icon_tint:              C.int(iconTintValue(networkCfg, networkCfg.Icon)),
		network_wired_icon_tint:        C.int(iconTintValue(networkCfg, networkCfg.WiredIcon)),
		network_show_text:              C.int(networkShowText),
		bluetooth_group:                C.int(bluetoothGroup),
		bluetooth_icon:                 bluetoothIcon,
		bluetooth_icon_color:           C.uint32_t(bluetoothIconColor),
		bluetooth_icon_tint:            C.int(iconTintValue(bluetoothCfg, bluetoothCfg.Icon)),
		bluetooth_show_text:            C.int(bluetoothShowText),
		battery_group:                  C.int(batteryGroup),
		battery_icon:                   batteryIcon,
		battery_icon_color:             C.uint32_t(batteryIconColor),
		battery_icon_tint:              C.int(iconTintValue(batteryCfg, batteryIconName)),
		battery_dynamic_icon:           C.int(batteryDynamicIcon),
		battery_show_text:              C.int(batteryShowText),
		keyboard_layout_group:          C.int(keyboardLayoutGroup),
		keyboard_layout_icon:           keyboardLayoutIcon,
		keyboard_layout_icon_color:     C.uint32_t(keyboardLayoutIconColor),
		keyboard_layout_icon_tint:      C.int(iconTintValue(keyboardLayoutCfg, keyboardLayoutCfg.Icon)),
		storage_group:                  C.int(storageGroup),
		storage_icon:                   storageIcon,
		storage_icon_color:             C.uint32_t(storageIconColor),
		storage_icon_tint:              C.int(iconTintValue(storageCfg, storageCfg.Icon)),
		ram_group:                      C.int(ramGroup),
		ram_icon:                       ramIcon,
		ram_icon_color:                 C.uint32_t(ramIconColor),
		ram_icon_tint:                  C.int(iconTintValue(ramCfg, ramCfg.Icon)),
		cpu_group:                      C.int(cpuGroup),
		cpu_icon:                       cpuIcon,
		cpu_icon_color:                 C.uint32_t(cpuIconColor),
		cpu_icon_tint:                  C.int(iconTintValue(cpuCfg, cpuCfg.Icon)),
		gpu_group:                      C.int(gpuGroup),
		gpu_icon:                       gpuIcon,
		gpu_icon_color:                 C.uint32_t(gpuIconColor),
		gpu_icon_tint:                  C.int(iconTintValue(gpuCfg, gpuCfg.Icon)),
		netstat_group:                  C.int(netstatGroup),
		netstat_icon:                   netstatIcon,
		netstat_icon_color:             C.uint32_t(netstatIconColor),
		netstat_icon_tint:              C.int(iconTintValue(netstatCfg, netstatCfg.Icon)),
		module_background_color:        C.uint32_t(cfg.ModuleBackgroundColor),
		module_icon_color:              C.uint32_t(cfg.ModuleIconColor),
		module_padding:                 C.int(cfg.ModulePadding),
		module_radius:                  C.int(cfg.ModuleRadius),
		module_icon_size:               C.int(cfg.ModuleIconSize),
		workspace_width:                C.int(cfg.WorkspaceWidth),
		workspace_height:               C.int(cfg.WorkspaceHeight),
		workspace_gap:                  C.int(cfg.WorkspaceGap),
		workspace_inset:                C.int(cfg.WorkspaceInset),
		workspace_color:                C.uint32_t(cfg.WorkspaceColor),
		workspace_active_color:         C.uint32_t(cfg.WorkspaceActiveColor),
		workspace_urgent_color:         C.uint32_t(cfg.WorkspaceUrgentColor),
		workspace_border_color:         C.uint32_t(cfg.WorkspaceBorderColor),
		workspace_window_color:         C.uint32_t(cfg.WorkspaceWindowColor),
		workspace_focused_window_color: C.uint32_t(cfg.WorkspaceFocusedWindowColor),
		menu_background_color:          C.uint32_t(cfg.MenuBackgroundColor),
		menu_foreground_color:          C.uint32_t(cfg.MenuForegroundColor),
		menu_selected_background_color: C.uint32_t(cfg.MenuSelectedBackgroundColor),
		menu_selected_foreground_color: C.uint32_t(cfg.MenuSelectedForegroundColor),
		menu_border_color:              C.uint32_t(cfg.MenuBorderColor),
		menu_font:                      menuFont,
		menu_padding:                   C.int(cfg.MenuPadding),
		menu_border_width:              C.int(cfg.MenuBorderWidth),
		menu_radius:                    C.int(cfg.MenuRadius),
	}

	ptr := C.hatwm_panel_create(&ccfg)
	if ptr == nil {
		return nil, fmt.Errorf("failed to create Wayland panel: %s", C.GoString(C.hatwm_panel_last_error()))
	}
	panel := &Panel{ptr: ptr}
	panel.SetGroupOrders(cfg)
	if launcherCfg.Icon == "" {
		panel.SetLauncherLogo()
	}
	panel.SetSeparators(cfg)
	panel.SetButtons(cfg, true)
	return panel, nil
}

func (p *Panel) SetClock(cfg Config, now time.Time) {
	if p == nil || p.ptr == nil {
		return
	}
	_, module, ok := clockModule(cfg)
	text := ""
	if ok {
		text = formatClock(now, module.Value)
	}
	value := C.CString(text)
	defer C.free(unsafe.Pointer(value))
	C.hatwm_panel_set_clock(p.ptr, value)
}

func (p *Panel) TakeClockClick() bool {
	return p != nil && p.ptr != nil && C.hatwm_panel_take_clock_click(p.ptr) != 0
}

func (p *Panel) ShowCalendar() bool {
	return p != nil && p.ptr != nil && C.hatwm_panel_show_calendar(p.ptr) != 0
}

func (p *Panel) Close() {
	if p == nil || p.ptr == nil {
		return
	}
	C.hatwm_panel_destroy(p.ptr)
	p.ptr = nil
}

func (p *Panel) SetText(left, center, right string) {
	if p == nil || p.ptr == nil {
		return
	}
	l := C.CString(left)
	c := C.CString(center)
	r := C.CString(right)
	defer C.free(unsafe.Pointer(l))
	defer C.free(unsafe.Pointer(c))
	defer C.free(unsafe.Pointer(r))
	C.hatwm_panel_set_text(p.ptr, l, c, r)
}

func (p *Panel) Dispatch(timeoutMS int) error {
	if p == nil || p.ptr == nil {
		return fmt.Errorf("panel is not initialized")
	}
	if rc := C.hatwm_panel_dispatch(p.ptr, C.int(timeoutMS)); rc < 0 {
		return fmt.Errorf("Wayland dispatch failed: %s", C.GoString(C.hatwm_panel_last_error()))
	}
	return nil
}

func (p *Panel) Closed() bool {
	return p == nil || p.ptr == nil || C.hatwm_panel_closed(p.ptr) != 0
}
