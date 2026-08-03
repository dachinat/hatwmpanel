package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include "panel.h"
*/
import "C"

const missingModuleOrder = 1 << 20

func (p *Panel) SetGroupOrders(cfg Config) {
	if p == nil || p.ptr == nil {
		return
	}
	for _, group := range []struct {
		modules []Module
		value   C.int
	}{
		{cfg.Left, C.HATWM_PANEL_GROUP_LEFT},
		{cfg.Center, C.HATWM_PANEL_GROUP_CENTER},
		{cfg.Right, C.HATWM_PANEL_GROUP_RIGHT},
	} {
		launcherOrder := missingModuleOrder
		workspaceOrder, trayOrder := missingModuleOrder, missingModuleOrder
		clockOrder, buttonOrder := missingModuleOrder, missingModuleOrder
		volumeOrder := missingModuleOrder
		brightnessOrder := missingModuleOrder
		microphoneOrder := missingModuleOrder
		networkOrder := missingModuleOrder
		bluetoothOrder := missingModuleOrder
		batteryOrder := missingModuleOrder
		keyboardLayoutOrder := missingModuleOrder
		storageOrder := missingModuleOrder
		ramOrder := missingModuleOrder
		cpuOrder := missingModuleOrder
		gpuOrder := missingModuleOrder
		netstatOrder := missingModuleOrder
		textOrder := missingModuleOrder
		for index, module := range group.modules {
			switch module.Kind {
			case "launcher":
				if launcherOrder == missingModuleOrder {
					launcherOrder = index
				}
			case "workspaces":
				if workspaceOrder == missingModuleOrder {
					workspaceOrder = index
				}
			case "tray":
				if trayOrder == missingModuleOrder {
					trayOrder = index
				}
			case "clock":
				if clockOrder == missingModuleOrder {
					clockOrder = index
				}
			case "button":
				if buttonOrder == missingModuleOrder {
					buttonOrder = index
				}
			case "volume":
				if volumeOrder == missingModuleOrder {
					volumeOrder = index
				}
			case "brightness":
				if brightnessOrder == missingModuleOrder {
					brightnessOrder = index
				}
			case "microphone":
				if microphoneOrder == missingModuleOrder {
					microphoneOrder = index
				}
			case "network":
				if networkOrder == missingModuleOrder {
					networkOrder = index
				}
			case "bluetooth":
				if bluetoothOrder == missingModuleOrder {
					bluetoothOrder = index
				}
			case "battery":
				if batteryOrder == missingModuleOrder {
					batteryOrder = index
				}
			case "keyboard_layout":
				if keyboardLayoutOrder == missingModuleOrder {
					keyboardLayoutOrder = index
				}
			case "storage":
				if storageOrder == missingModuleOrder {
					storageOrder = index
				}
			case "ram":
				if ramOrder == missingModuleOrder {
					ramOrder = index
				}
			case "cpu":
				if cpuOrder == missingModuleOrder {
					cpuOrder = index
				}
			case "gpu":
				if gpuOrder == missingModuleOrder {
					gpuOrder = index
				}
			case "netstat":
				if netstatOrder == missingModuleOrder {
					netstatOrder = index
				}
			case "text", "exec":
				if textOrder == missingModuleOrder {
					textOrder = index
				}
			}
		}
		C.hatwm_panel_set_group_order(
			p.ptr, group.value,
			C.int(launcherOrder), C.int(workspaceOrder),
			C.int(trayOrder), C.int(clockOrder),
			C.int(volumeOrder), C.int(brightnessOrder),
			C.int(microphoneOrder), C.int(networkOrder),
			C.int(bluetoothOrder), C.int(batteryOrder),
			C.int(keyboardLayoutOrder), C.int(storageOrder),
			C.int(ramOrder), C.int(cpuOrder), C.int(gpuOrder),
			C.int(netstatOrder),
			C.int(buttonOrder), C.int(textOrder),
		)
	}
}
