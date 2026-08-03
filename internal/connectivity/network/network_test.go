package network

import "testing"

func TestParseNetworks(t *testing.T) {
	got := parseNetworks("*:Home\\: 5G:87:WPA2\n:Guest:54:--\n:Home\\: 5G:42:WPA2\n")
	if len(got) != 2 {
		t.Fatalf("got %d networks: %#v", len(got), got)
	}
	if got[0].SSID != "Home: 5G" || !got[0].Active || got[0].Signal != 87 {
		t.Fatalf("unexpected active network: %#v", got[0])
	}
	if got[1].SSID != "Guest" || got[1].Secured() {
		t.Fatalf("unexpected open network: %#v", got[1])
	}
}

func TestConnectedDevice(t *testing.T) {
	status := "enp4s0:ethernet:connected\nwlan0:wifi:connected\nlo:loopback:connected\n"
	if got := connectedDevice(status, KindWiFi); got != "wlan0" {
		t.Fatalf("connectedDevice(wifi) = %q, want wlan0", got)
	}
	if got := connectedDevice(status, KindEthernet); got != "enp4s0" {
		t.Fatalf("connectedDevice(ethernet) = %q, want enp4s0", got)
	}
	if got := connectedDevice("wlan0:wifi:disconnected\n", KindWiFi); got != "" {
		t.Fatalf("unexpected disconnected Wi-Fi device: %q", got)
	}
}

func TestParseEthernetConnections(t *testing.T) {
	got := parseEthernetConnections(
		"Wired connection 1:uuid-1:802-3-ethernet:enp4s0\n" +
			"Office LAN:uuid-2:802-3-ethernet:\n" +
			"Zeda:uuid-3:802-11-wireless:wlan0\n")
	if len(got) != 2 {
		t.Fatalf("got %d Ethernet profiles: %#v", len(got), got)
	}
	if got[0].SSID != "Wired connection 1" || !got[0].Active ||
		got[0].Device != "enp4s0" || !got[0].Wired() {
		t.Fatalf("unexpected active Ethernet profile: %#v", got[0])
	}
	if got[1].SSID != "Office LAN" || got[1].Active || got[1].UUID != "uuid-2" {
		t.Fatalf("unexpected inactive Ethernet profile: %#v", got[1])
	}
}

func TestSplitEscapedPreservesEscapedSeparatorsAndBackslashes(t *testing.T) {
	got := splitEscaped(`Home\: Office:path\\name:`, ':')
	want := []string{"Home: Office", `path\name`, ""}
	if len(got) != len(want) {
		t.Fatalf("splitEscaped() = %#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("splitEscaped()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseNetworksSkipsMalformedRowsAndKeepsDevice(t *testing.T) {
	got := parseNetworks(
		"malformed\n" +
			":Cafe:72:WPA2:wlan0\n" +
			":Cafe:91:WPA3:wlan1\n" +
			"*:Guest:35:--:wlan2\n" +
			":Missing Signal:not-a-number:WPA2:wlan3\n")
	if len(got) != 3 {
		t.Fatalf("parseNetworks() = %#v", got)
	}
	if got[0].SSID != "Guest" || !got[0].Active || got[0].Device != "wlan2" {
		t.Fatalf("active network = %#v", got[0])
	}
	if got[1].SSID != "Cafe" || got[1].Signal != 91 || got[1].Device != "wlan1" {
		t.Fatalf("strongest duplicate = %#v", got[1])
	}
}

func TestNetworkKindAndSecurityHelpers(t *testing.T) {
	for _, test := range []struct {
		network Network
		secured bool
		wired   bool
	}{
		{Network{Kind: KindWiFi, Security: "WPA2"}, true, false},
		{Network{Kind: KindWiFi, Security: " -- "}, false, false},
		{Network{Kind: KindWiFi}, false, false},
		{Network{Kind: KindEthernet, Security: "WPA2"}, false, true},
	} {
		if got := test.network.Secured(); got != test.secured {
			t.Fatalf("%#v.Secured() = %v", test.network, got)
		}
		if got := test.network.Wired(); got != test.wired {
			t.Fatalf("%#v.Wired() = %v", test.network, got)
		}
	}
}

func TestParseEthernetConnectionsHandlesEscapedNamesAndLegacyType(t *testing.T) {
	got := parseEthernetConnections(
		"Office\\: Wired:uuid-1:ethernet:enp5s0\n" +
			"Incomplete:uuid-2:ethernet\n")
	if len(got) != 1 || got[0].SSID != "Office: Wired" ||
		got[0].UUID != "uuid-1" || got[0].Device != "enp5s0" || !got[0].Active {
		t.Fatalf("parseEthernetConnections() = %#v", got)
	}
}
