package battery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFrom(t *testing.T) {
	root := t.TempDir()
	batteryPath := filepath.Join(root, "BAT0")
	if err := os.Mkdir(batteryPath, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{
		"type": "Battery\n", "capacity": "73\n", "status": "Charging\n",
	} {
		if err := os.WriteFile(filepath.Join(batteryPath, name), []byte(value), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	state, err := readFrom(root)
	if err != nil {
		t.Fatal(err)
	}
	if state.Percent != 73 || !state.Charging || state.Full {
		t.Fatalf("unexpected battery state: %#v", state)
	}
}

func TestReadFromWithoutBattery(t *testing.T) {
	if _, err := readFrom(t.TempDir()); err == nil {
		t.Fatal("expected no-battery error")
	}
}

func TestReadFromClampsCapacityAndRecognizesFullState(t *testing.T) {
	for name, capacity := range map[string]string{"low": "-12", "high": "145"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "BAT0")
			if err := os.Mkdir(path, 0o755); err != nil {
				t.Fatal(err)
			}
			for file, value := range map[string]string{
				"type": "battery", "capacity": capacity, "status": "Full",
			} {
				if err := os.WriteFile(filepath.Join(path, file), []byte(value), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			state, err := readFrom(root)
			if err != nil {
				t.Fatal(err)
			}
			want := 0
			if name == "high" {
				want = 100
			}
			if state.Percent != want || state.Charging || !state.Full {
				t.Fatalf("state = %#v, want percent=%d and full", state, want)
			}
		})
	}
}
