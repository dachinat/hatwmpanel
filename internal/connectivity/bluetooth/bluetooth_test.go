package bluetooth

import "testing"

func TestParseDevices(t *testing.T) {
	all := "Device AA:BB:CC:DD:EE:01 Headphones Pro\n" +
		"Device AA:BB:CC:DD:EE:02 Keyboard\n"
	connected := "Device AA:BB:CC:DD:EE:01 Headphones Pro\n"
	got := parseDevices(all, connected)
	if len(got) != 2 {
		t.Fatalf("got %d devices: %#v", len(got), got)
	}
	if got[0].Name != "Headphones Pro" || !got[0].Connected {
		t.Fatalf("unexpected connected device: %#v", got[0])
	}
	if got[1].Name != "Keyboard" || got[1].Connected {
		t.Fatalf("unexpected disconnected device: %#v", got[1])
	}
}

func TestParseDeviceLinesSkipsMalformedAndDuplicateEntries(t *testing.T) {
	output := "Controller AA:BB:CC:DD:EE:FF Host\n" +
		"Device AA:BB:CC:DD:EE:01 First Name\n" +
		"Device AA:BB:CC:DD:EE:01 Renamed Duplicate\n" +
		"Device missing-name\n" +
		"  Device AA:BB:CC:DD:EE:02 Second Device  \n"
	devices := parseDeviceLines(output)
	if len(devices) != 2 {
		t.Fatalf("parseDeviceLines() = %#v", devices)
	}
	if devices[0].Address != "AA:BB:CC:DD:EE:01" || devices[0].Name != "First Name" {
		t.Fatalf("first device = %#v", devices[0])
	}
	if devices[1].Name != "Second Device" {
		t.Fatalf("second device = %#v", devices[1])
	}
}

func TestParseDevicesSortsConnectedFirstThenByName(t *testing.T) {
	all := "Device 00:00:00:00:00:01 zebra\n" +
		"Device 00:00:00:00:00:02 alpha\n" +
		"Device 00:00:00:00:00:03 Beta\n"
	connected := "Device 00:00:00:00:00:03 Beta\n"
	devices := parseDevices(all, connected)
	if len(devices) != 3 || !devices[0].Connected || devices[0].Name != "Beta" ||
		devices[1].Name != "alpha" || devices[2].Name != "zebra" {
		t.Fatalf("sorted devices = %#v", devices)
	}
}
