package workspaces

import (
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestIPCClientReceivesGeometryAndSwitchesWorkspace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ipc.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	t.Setenv("HATWM_SOCKET", path)

	commands := make(chan ipcRequest, 2)
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		decoder := json.NewDecoder(conn)
		encoder := json.NewEncoder(conn)
		for {
			var req ipcRequest
			if err := decoder.Decode(&req); err != nil {
				return
			}
			ok := true
			switch req.Type {
			case "hello":
				_ = encoder.Encode(map[string]any{
					"type":             "hello",
					"id":               req.ID,
					"success":          ok,
					"protocol_version": panelIPCProtocolVersion,
				})
			case "get_state":
				_ = encoder.Encode(map[string]any{
					"type": "response", "id": req.ID, "success": ok,
					"result": IPCState{
						Workspace: 1, WorkspaceCount: 2,
						KeyboardLayout: "us",
						Tiling:         true,
						Workspaces:     []IPCWorkspace{{Number: 1, Active: true, Windows: 1}, {Number: 2, Urgent: true}},
						Output:         IPCOutput{Width: 1920, Height: 1080, UsableY: 32, UsableWidth: 1920, UsableHeight: 1048},
					},
				})
			case "get_windows":
				_ = encoder.Encode(map[string]any{
					"type": "response", "id": req.ID, "success": ok,
					"result": []IPCWindow{{ID: 7, Workspace: 1, Mapped: true, Focused: true, X: 20, Y: 52, Width: 930, Height: 1008}},
				})
			case "subscribe":
				_ = encoder.Encode(map[string]any{"type": "response", "id": req.ID, "success": ok, "result": map[string]any{"events": req.Events}})
			case "command":
				commands <- req
				_ = encoder.Encode(map[string]any{"type": "response", "id": req.ID, "success": ok, "result": IPCState{Workspace: req.Workspace, WorkspaceCount: 2}})
			}
		}
	}()

	client := NewIPCClient()
	defer client.Close()

	var snapshot WorkspaceSnapshot
	select {
	case snapshot = <-client.Updates():
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for workspace snapshot")
	}
	for !snapshot.Connected || len(snapshot.Windows) == 0 {
		select {
		case snapshot = <-client.Updates():
		case <-time.After(3 * time.Second):
			t.Fatal("timed out waiting for complete workspace snapshot")
		}
	}
	if snapshot.Count != 2 || snapshot.Output.Width != 1920 || !snapshot.Tiling ||
		snapshot.KeyboardLayout != "us" {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	if !snapshot.Workspaces[1].Urgent {
		t.Fatalf("urgent workspace state was not preserved: %#v", snapshot.Workspaces)
	}
	window := snapshot.Windows[0]
	if window.X != 20 || window.Y != 52 || window.Width != 930 || window.Height != 1008 {
		t.Fatalf("unexpected window geometry: %#v", window)
	}

	client.SwitchWorkspace(2)
	select {
	case command := <-commands:
		if command.Command != "workspace" || command.Workspace != 2 {
			t.Fatalf("unexpected workspace command: %#v", command)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for workspace command")
	}

	client.ToggleKeyboardLayout()
	select {
	case command := <-commands:
		if command.Command != "toggle_keyboard_layout" {
			t.Fatalf("unexpected keyboard command: %#v", command)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for keyboard-layout command")
	}

	client.ToggleTiling()
	select {
	case command := <-commands:
		if command.Command != "toggle_tiling" {
			t.Fatalf("unexpected tiling command: %#v", command)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for tiling command")
	}
}
