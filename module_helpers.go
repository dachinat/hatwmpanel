package main

func moduleByKind(cfg Config, kind string) (int, Module, bool) {
	for groupIndex, group := range [][]Module{cfg.Left, cfg.Center, cfg.Right} {
		for _, module := range group {
			if module.Kind == kind {
				return groupIndex + 1, module, true
			}
		}
	}
	return 0, Module{}, false
}

func volumeModule(cfg Config) (int, Module, bool) {
	return moduleByKind(cfg, "volume")
}

func brightnessModule(cfg Config) (int, Module, bool) {
	return moduleByKind(cfg, "brightness")
}

func microphoneModule(cfg Config) (int, Module, bool) {
	return moduleByKind(cfg, "microphone")
}

func networkModule(cfg Config) (int, Module, bool) {
	return moduleByKind(cfg, "network")
}

func bluetoothModule(cfg Config) (int, Module, bool) {
	return moduleByKind(cfg, "bluetooth")
}

func batteryModule(cfg Config) (int, Module, bool) {
	return moduleByKind(cfg, "battery")
}

func storageModule(cfg Config) (int, Module, bool) {
	return moduleByKind(cfg, "storage")
}

func ramModule(cfg Config) (int, Module, bool) {
	return moduleByKind(cfg, "ram")
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
