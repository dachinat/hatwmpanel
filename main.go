package main

import (
	"errors"
	"hatwmpanel/internal/audio/microphone"
	"hatwmpanel/internal/audio/volume"
	"hatwmpanel/internal/connectivity/bluetooth"
	netstatmodule "hatwmpanel/internal/connectivity/netstat"
	"hatwmpanel/internal/connectivity/network"
	traymodule "hatwmpanel/internal/desktop/tray"
	"hatwmpanel/internal/hatwm/keyboardlayout"
	"hatwmpanel/internal/hatwm/workspaces"
	"hatwmpanel/internal/system/battery"
	"hatwmpanel/internal/system/brightness"
	cpumodule "hatwmpanel/internal/system/cpu"
	gpumodule "hatwmpanel/internal/system/gpu"
	memorymodule "hatwmpanel/internal/system/memory"
	storagemodule "hatwmpanel/internal/system/storage"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		slog.Error("hatwmpanel stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	panel, err := NewNativePanel(cfg)
	if err != nil {
		return err
	}
	defer panel.Close()

	tray, err := traymodule.NewTray()
	if err != nil {
		slog.Warn("system tray unavailable", "error", err)
	}
	if tray != nil {
		defer tray.Close()
		panel.SetTray(tray.Items())
	}

	runner := NewRunner()
	volumeController := volume.NewController()
	defer volumeController.Close()
	volumeState := volume.State{}
	if _, _, ok := volumeModule(cfg); ok {
		if state, readErr := volume.Read(); readErr == nil {
			volumeState = state
		}
		panel.SetVolume(volumeState)
	}
	brightnessController := brightness.NewController()
	defer brightnessController.Close()
	brightnessPercent := 1
	if _, _, ok := brightnessModule(cfg); ok {
		if percent, readErr := brightness.Read(); readErr == nil {
			brightnessPercent = percent
		}
		panel.SetBrightness(brightnessPercent)
	}
	microphoneController := microphone.NewController()
	defer microphoneController.Close()
	microphoneState := microphone.State{}
	if _, _, ok := microphoneModule(cfg); ok {
		if state, readErr := microphone.Read(); readErr == nil {
			microphoneState = state
		}
		panel.SetMicrophone(microphoneState)
	}
	batteryState := battery.State{}
	if _, _, ok := batteryModule(cfg); ok {
		if state, readErr := battery.Read(); readErr == nil {
			batteryState = state
		}
		panel.SetBattery(batteryState)
	}
	storageText := "-- / --"
	if _, storageCfg, ok := storageModule(cfg); ok {
		if usage, readErr := storagemodule.Read(storageCfg.Value); readErr == nil {
			storageText = usage.String()
		}
		panel.SetStorage(storageText)
	}
	ramText := "-- / --"
	if _, _, ok := ramModule(cfg); ok {
		if usage, readErr := memorymodule.Read(); readErr == nil {
			ramText = usage.String()
		}
		panel.SetRAM(ramText)
	}
	cpuMonitor := &cpumodule.Monitor{}
	cpuText := "0%"
	if _, _, ok := cpuModule(cfg); ok {
		if percent, readErr := cpuMonitor.Read(); readErr == nil {
			cpuText = strconv.Itoa(percent) + "%"
		}
		panel.SetCPU(cpuText)
	}
	gpuText := "N/A"
	if _, _, ok := gpuModule(cfg); ok {
		if percent, readErr := gpumodule.Read(); readErr == nil {
			gpuText = strconv.Itoa(percent) + "%"
		}
		panel.SetGPU(gpuText)
	}
	netstatMonitor := &netstatmodule.Monitor{}
	netstatText := (netstatmodule.Rate{}).String()
	if _, _, ok := netstatModule(cfg); ok {
		if rate, readErr := netstatMonitor.Read(); readErr == nil {
			netstatText = rate.String()
		}
		panel.SetNetstat(netstatText)
	}
	networkController := network.NewController()
	defer networkController.Close()
	var networks []network.Network
	networkUpdates := make(chan []network.Network, 1)
	requestNetworkScan := func(rescan bool) {
		go func() {
			items, scanErr := network.Scan(rescan)
			if scanErr != nil {
				return
			}
			select {
			case networkUpdates <- items:
			default:
				select {
				case <-networkUpdates:
				default:
				}
				networkUpdates <- items
			}
		}()
	}
	if _, _, ok := networkModule(cfg); ok {
		requestNetworkScan(false)
	}
	bluetoothController := bluetooth.NewController()
	defer bluetoothController.Close()
	var bluetoothDevices []bluetooth.Device
	bluetoothUpdates := make(chan []bluetooth.Device, 1)
	requestBluetoothScan := func(discover bool) {
		go func() {
			devices, scanErr := bluetooth.Scan(discover)
			if scanErr != nil {
				return
			}
			select {
			case bluetoothUpdates <- devices:
			default:
				select {
				case <-bluetoothUpdates:
				default:
				}
				bluetoothUpdates <- devices
			}
		}()
	}
	if _, _, ok := bluetoothModule(cfg); ok {
		requestBluetoothScan(false)
	}
	ipc := workspaces.NewIPCClient()
	defer ipc.Close()
	var workspaceSnapshot workspaces.WorkspaceSnapshot
	tilingMode := true
	tilingKnown := false
	updatePanelText(panel, runner, cfg, time.Now())

	configPath := GetConfigPath()
	var configModTime time.Time
	if st, err := os.Stat(configPath); err == nil {
		configModTime = st.ModTime()
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(
		signals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR1)
	defer signal.Stop(signals)

	lastTextUpdate := time.Time{}
	lastConfigCheck := time.Time{}
	lastNetworkScan := time.Time{}
	lastBluetoothScan := time.Time{}
	lastGPUUpdate := time.Now()

	for !panel.Closed() {
		select {
		case sig := <-signals:
			if sig == syscall.SIGUSR1 {
				panel.ShowLauncher()
				continue
			}
			if sig == syscall.SIGHUP {
				newCfg, err := LoadConfig()
				if err != nil {
					slog.Error("config reload failed", "error", err)
					continue
				}
				newPanel, err := NewNativePanel(newCfg)
				if err != nil {
					slog.Error("could not apply reloaded config", "error", err)
					continue
				}
				panel.Close()
				panel = newPanel
				cfg = newCfg
				panel.SetButtons(cfg, tilingMode)
				runner = NewRunner()
				updatePanelText(panel, runner, cfg, time.Now())
				if tray != nil {
					panel.SetTray(tray.Items())
				}
				panel.SetWorkspaces(workspaceSnapshot)
				panel.SetKeyboardLayout(
					keyboardlayout.Code(workspaceSnapshot.KeyboardLayout))
				panel.SetVolume(volumeState)
				panel.SetBrightness(brightnessPercent)
				panel.SetMicrophone(microphoneState)
				panel.SetNetworks(networks)
				panel.SetBluetoothDevices(bluetoothDevices)
				panel.SetBattery(batteryState)
				panel.SetStorage(storageText)
				panel.SetRAM(ramText)
				panel.SetCPU(cpuText)
				panel.SetGPU(gpuText)
				panel.SetNetstat(netstatText)
				continue
			}
			return nil
		default:
		}

		select {
		case workspaceSnapshot = <-ipc.Updates():
			if !tilingKnown || tilingMode != workspaceSnapshot.Tiling {
				tilingMode = workspaceSnapshot.Tiling
				tilingKnown = true
				panel.SetButtons(cfg, tilingMode)
			}
			panel.SetWorkspaces(workspaceSnapshot)
			panel.SetKeyboardLayout(
				keyboardlayout.Code(workspaceSnapshot.KeyboardLayout))
		default:
		}

		if err := panel.Dispatch(200); err != nil {
			return err
		}
		if workspace := panel.TakeWorkspaceClick(); workspace > 0 {
			ipc.SwitchWorkspace(workspace)
		}
		if panel.TakeKeyboardLayoutClick() {
			ipc.ToggleKeyboardLayout()
		}
		if panel.TakeClockClick() {
			panel.ShowCalendar()
		}
		if panel.TakeLauncherClick() {
			panel.ShowLauncher()
		}
		if button := panel.TakeButtonClick(); button >= 0 {
			if panel.ButtonKind(button) == "layout_mode" {
				tilingMode = !tilingMode
				tilingKnown = true
				panel.SetButtons(cfg, tilingMode)
				ipc.ToggleTiling()
			} else {
				panel.ActivateButton(button)
			}
		}
		if percent := panel.TakeVolumeChange(); percent >= 0 {
			volumeState.Percent = percent
			volumeState.Muted = false
			panel.SetVolume(volumeState)
			volumeController.SetPercent(percent)
		}
		if percent := panel.TakeBrightnessChange(); percent >= 0 {
			brightnessPercent = percent
			panel.SetBrightness(brightnessPercent)
			brightnessController.SetPercent(percent)
		}
		if percent := panel.TakeMicrophoneChange(); percent >= 0 {
			microphoneState.Percent = percent
			microphoneState.Muted = false
			panel.SetMicrophone(microphoneState)
			microphoneController.SetPercent(percent)
		}
		// Read a menu selection before applying a newly scanned list so the
		// native row index and this snapshot always refer to the same SSID.
		if index := panel.TakeNetworkClick(); index >= 0 && index < len(networks) {
			slog.Info("connecting to network", "name", networks[index].SSID)
			networkController.Connect(networks[index])
		}
		if panel.TakeNetworkDisconnect() {
			var connected network.Network
			for _, candidate := range networks {
				if candidate.Active {
					connected = candidate
					if candidate.Wired() {
						break
					}
				}
			}
			slog.Info("disconnecting network", "name", connected.SSID)
			networkController.Disconnect(connected)
		}
		if index := panel.TakeBluetoothClick(); index >= 0 &&
			index < len(bluetoothDevices) {
			slog.Info("connecting to Bluetooth device",
				"name", bluetoothDevices[index].Name)
			bluetoothController.Connect(bluetoothDevices[index])
		}
		if panel.TakeBluetoothDisconnect() {
			for _, device := range bluetoothDevices {
				if device.Connected {
					slog.Info("disconnecting Bluetooth device", "name", device.Name)
					bluetoothController.Disconnect(device)
					break
				}
			}
		}
		select {
		case networks = <-networkUpdates:
			panel.SetNetworks(networks)
		default:
		}
		select {
		case result := <-networkController.Results():
			if result.Err != nil {
				if result.Disconnected {
					slog.Warn("network disconnect failed",
						"name", result.SSID, "error", result.Err)
				} else {
					slog.Warn("network connection failed",
						"name", result.SSID, "error", result.Err)
				}
			}
			requestNetworkScan(false)
		default:
		}
		select {
		case bluetoothDevices = <-bluetoothUpdates:
			panel.SetBluetoothDevices(bluetoothDevices)
		default:
		}
		select {
		case result := <-bluetoothController.Results():
			if result.Err != nil {
				action := "connection"
				if result.Disconnected {
					action = "disconnect"
				}
				slog.Warn("Bluetooth action failed",
					"action", action, "name", result.Name, "error", result.Err)
			}
			requestBluetoothScan(false)
		default:
		}
		if panel.TakeNetworkMenuRequest() {
			panel.SetNetworks(networks)
			panel.ShowNetworkMenu()
			requestNetworkScan(true)
		}
		if panel.TakeBluetoothMenuRequest() {
			panel.SetBluetoothDevices(bluetoothDevices)
			panel.ShowBluetoothMenu()
			requestBluetoothScan(true)
		}
		if tray != nil {
			panel.SetTray(tray.Items())
			if index, button := panel.TakeTrayClick(); index >= 0 {
				if service, path, ok := tray.Menu(index, button); ok &&
					panel.ShowTrayMenu(service, path) {
					continue
				}
				tray.Interact(index, button, 0, 0)
			}
		}

		now := time.Now()
		if lastTextUpdate.IsZero() || now.Sub(lastTextUpdate) >= time.Second {
			updatePanelText(panel, runner, cfg, now)
			if _, _, ok := volumeModule(cfg); ok {
				if state, readErr := volume.Read(); readErr == nil {
					volumeState = state
					panel.SetVolume(volumeState)
				}
			}
			if _, _, ok := brightnessModule(cfg); ok {
				if percent, readErr := brightness.Read(); readErr == nil {
					brightnessPercent = percent
					panel.SetBrightness(brightnessPercent)
				}
			}
			if _, _, ok := microphoneModule(cfg); ok {
				if state, readErr := microphone.Read(); readErr == nil {
					microphoneState = state
					panel.SetMicrophone(microphoneState)
				}
			}
			if _, _, ok := batteryModule(cfg); ok {
				if state, readErr := battery.Read(); readErr == nil {
					batteryState = state
					panel.SetBattery(batteryState)
				}
			}
			if _, storageCfg, ok := storageModule(cfg); ok {
				if usage, readErr := storagemodule.Read(storageCfg.Value); readErr == nil {
					storageText = usage.String()
					panel.SetStorage(storageText)
				}
			}
			if _, _, ok := ramModule(cfg); ok {
				if usage, readErr := memorymodule.Read(); readErr == nil {
					ramText = usage.String()
					panel.SetRAM(ramText)
				}
			}
			if _, _, ok := cpuModule(cfg); ok {
				if percent, readErr := cpuMonitor.Read(); readErr == nil {
					cpuText = strconv.Itoa(percent) + "%"
					panel.SetCPU(cpuText)
				}
			}
			if _, _, ok := netstatModule(cfg); ok {
				if rate, readErr := netstatMonitor.Read(); readErr == nil {
					netstatText = rate.String()
					panel.SetNetstat(netstatText)
				}
			}
			lastTextUpdate = now
		}
		if _, _, ok := gpuModule(cfg); ok &&
			now.Sub(lastGPUUpdate) >= 5*time.Second {
			lastGPUUpdate = now
			if percent, readErr := gpumodule.Read(); readErr == nil {
				gpuText = strconv.Itoa(percent) + "%"
			} else {
				gpuText = "N/A"
			}
			panel.SetGPU(gpuText)
		}

		if now.Sub(lastConfigCheck) >= time.Second {
			lastConfigCheck = now
			st, err := os.Stat(configPath)
			if err == nil && st.ModTime().After(configModTime) {
				newCfg, loadErr := LoadConfig()
				if loadErr != nil {
					slog.Error("config reload failed", "error", loadErr)
					configModTime = st.ModTime()
					continue
				}
				newPanel, createErr := NewNativePanel(newCfg)
				if createErr != nil {
					slog.Error("could not apply reloaded config", "error", createErr)
					configModTime = st.ModTime()
					continue
				}
				panel.Close()
				panel = newPanel
				cfg = newCfg
				panel.SetButtons(cfg, tilingMode)
				runner = NewRunner()
				updatePanelText(panel, runner, cfg, now)
				if tray != nil {
					panel.SetTray(tray.Items())
				}
				panel.SetWorkspaces(workspaceSnapshot)
				panel.SetKeyboardLayout(
					keyboardlayout.Code(workspaceSnapshot.KeyboardLayout))
				panel.SetVolume(volumeState)
				panel.SetBrightness(brightnessPercent)
				panel.SetMicrophone(microphoneState)
				panel.SetNetworks(networks)
				panel.SetBluetoothDevices(bluetoothDevices)
				panel.SetBattery(batteryState)
				panel.SetStorage(storageText)
				panel.SetRAM(ramText)
				panel.SetCPU(cpuText)
				panel.SetGPU(gpuText)
				panel.SetNetstat(netstatText)
				configModTime = st.ModTime()
				slog.Info("config reloaded")
			}
		}
		if _, _, ok := networkModule(cfg); ok &&
			(lastNetworkScan.IsZero() || now.Sub(lastNetworkScan) >= 10*time.Second) {
			lastNetworkScan = now
			requestNetworkScan(false)
		}
		if _, _, ok := bluetoothModule(cfg); ok &&
			(lastBluetoothScan.IsZero() ||
				now.Sub(lastBluetoothScan) >= 10*time.Second) {
			lastBluetoothScan = now
			requestBluetoothScan(false)
		}
	}

	return errors.New("compositor closed the layer surface")
}

func updatePanelText(panel *NativePanel, runner *Runner, cfg Config, now time.Time) {
	left := runner.Render(cfg.Left, now, cfg.Separator)
	center := runner.Render(cfg.Center, now, cfg.Separator)
	right := runner.Render(cfg.Right, now, cfg.Separator)
	panel.SetText(left, center, right)
	panel.SetClock(cfg, now)
}

func init() {
	// Keep log output compact and useful when started from HatWM autostart.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
}
