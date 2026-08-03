package workspaces

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const panelIPCProtocolVersion = 1

type IPCWorkspace struct {
	Number  int  `json:"number"`
	Active  bool `json:"active"`
	Focused bool `json:"focused"`
	Urgent  bool `json:"urgent"`
	Windows int  `json:"windows"`
}

type IPCOutput struct {
	X            int `json:"x"`
	Y            int `json:"y"`
	Width        int `json:"width"`
	Height       int `json:"height"`
	UsableX      int `json:"usable_x"`
	UsableY      int `json:"usable_y"`
	UsableWidth  int `json:"usable_width"`
	UsableHeight int `json:"usable_height"`
}

type IPCWindow struct {
	ID         uint64 `json:"id"`
	Workspace  int    `json:"workspace"`
	Mapped     bool   `json:"mapped"`
	Focused    bool   `json:"focused"`
	Urgent     bool   `json:"urgent"`
	Fullscreen bool   `json:"fullscreen"`
	XWayland   bool   `json:"xwayland"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
}

type IPCState struct {
	Workspace      int            `json:"workspace"`
	WorkspaceCount int            `json:"workspace_count"`
	KeyboardLayout string         `json:"keyboard_layout"`
	Tiling         bool           `json:"tiling"`
	Workspaces     []IPCWorkspace `json:"workspaces"`
	Output         IPCOutput      `json:"output"`
}

type WorkspaceSnapshot struct {
	Connected      bool
	Current        int
	Count          int
	Workspaces     []IPCWorkspace
	Windows        []IPCWindow
	Output         IPCOutput
	KeyboardLayout string
	Tiling         bool
}

type ipcCommand struct {
	name      string
	workspace int
}

type ipcRequest struct {
	Type            string   `json:"type"`
	ID              int64    `json:"id,omitempty"`
	ProtocolVersion int      `json:"protocol_version,omitempty"`
	Client          string   `json:"client,omitempty"`
	ClientVersion   string   `json:"client_version,omitempty"`
	Events          []string `json:"events,omitempty"`
	Command         string   `json:"command,omitempty"`
	Workspace       int      `json:"workspace,omitempty"`
}

type ipcMessage struct {
	Type            string          `json:"type"`
	ID              json.RawMessage `json:"id,omitempty"`
	Success         *bool           `json:"success,omitempty"`
	Error           string          `json:"error,omitempty"`
	ProtocolVersion int             `json:"protocol_version,omitempty"`
	Event           string          `json:"event,omitempty"`
	Result          json.RawMessage `json:"result,omitempty"`
}

type IPCClient struct {
	updates  chan WorkspaceSnapshot
	commands chan ipcCommand
	stop     chan struct{}
	done     chan struct{}
	close    sync.Once
}

func NewIPCClient() *IPCClient {
	client := &IPCClient{
		updates:  make(chan WorkspaceSnapshot, 1),
		commands: make(chan ipcCommand, 8),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	go client.run()
	return client
}

func (c *IPCClient) Updates() <-chan WorkspaceSnapshot { return c.updates }

func (c *IPCClient) SwitchWorkspace(number int) {
	if c == nil || number < 1 {
		return
	}
	select {
	case c.commands <- ipcCommand{name: "workspace", workspace: number}:
	default:
	}
}

func (c *IPCClient) ToggleKeyboardLayout() {
	if c == nil {
		return
	}
	select {
	case c.commands <- ipcCommand{name: "toggle_keyboard_layout"}:
	default:
	}
}

func (c *IPCClient) ToggleTiling() {
	if c == nil {
		return
	}
	select {
	case c.commands <- ipcCommand{name: "toggle_tiling"}:
	default:
	}
}

func (c *IPCClient) Close() {
	if c == nil {
		return
	}
	c.close.Do(func() { close(c.stop) })
	<-c.done
}

func defaultIPCSocketPath() (string, error) {
	if path := os.Getenv("HATWM_SOCKET"); path != "" {
		return path, nil
	}
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		return "", errors.New("neither HATWM_SOCKET nor XDG_RUNTIME_DIR is set")
	}
	return filepath.Join(runtimeDir, "hatwm", "ipc.sock"), nil
}

func (c *IPCClient) publish(snapshot WorkspaceSnapshot) {
	select {
	case c.updates <- snapshot:
	default:
		select {
		case <-c.updates:
		default:
		}
		select {
		case c.updates <- snapshot:
		default:
		}
	}
}

func (c *IPCClient) run() {
	defer close(c.done)
	var last WorkspaceSnapshot
	for {
		select {
		case <-c.stop:
			return
		default:
		}

		path, err := defaultIPCSocketPath()
		if err != nil {
			last.Connected = false
			c.publish(last)
			if !c.waitReconnect() {
				return
			}
			continue
		}

		conn, err := net.DialTimeout("unix", path, 500*time.Millisecond)
		if err != nil {
			last.Connected = false
			c.publish(last)
			if !c.waitReconnect() {
				return
			}
			continue
		}

		slog.Info("connected to HatWM IPC", "socket", path)
		last = c.runConnection(conn, last)
		_ = conn.Close()
		last.Connected = false
		c.publish(last)
		slog.Warn("HatWM IPC disconnected; retrying")
		if !c.waitReconnect() {
			return
		}
	}
}

func (c *IPCClient) waitReconnect() bool {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-c.stop:
		return false
	}
}

func (c *IPCClient) runConnection(conn net.Conn, previous WorkspaceSnapshot) WorkspaceSnapshot {
	rawMessages := make(chan []byte, 32)
	readErr := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(conn)
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		for scanner.Scan() {
			line := append([]byte(nil), scanner.Bytes()...)
			select {
			case rawMessages <- line:
			case <-c.stop:
				return
			}
		}
		readErr <- scanner.Err()
	}()

	encoder := json.NewEncoder(conn)
	pending := make(map[int64]string)
	nextID := int64(1)
	state := IPCState{
		Workspace:      previous.Current,
		WorkspaceCount: previous.Count,
		KeyboardLayout: previous.KeyboardLayout,
		Tiling:         previous.Tiling,
		Workspaces:     append([]IPCWorkspace(nil), previous.Workspaces...),
		Output:         previous.Output,
	}
	windows := append([]IPCWindow(nil), previous.Windows...)
	haveState := previous.Count > 0
	haveWindows := previous.Windows != nil

	send := func(req ipcRequest, kind string) bool {
		if req.ID == 0 {
			req.ID = nextID
			nextID++
		}
		if kind != "" {
			pending[req.ID] = kind
		}
		if err := encoder.Encode(req); err != nil {
			slog.Warn("HatWM IPC write failed", "error", err)
			return false
		}
		return true
	}

	requestRefresh := func() bool {
		return send(ipcRequest{Type: "get_state"}, "state") &&
			send(ipcRequest{Type: "get_windows"}, "windows")
	}

	if !send(ipcRequest{
		Type:            "hello",
		ProtocolVersion: panelIPCProtocolVersion,
		Client:          "hatwmpanel",
		ClientVersion:   "0.3.0",
	}, "hello") {
		return previous
	}
	if !requestRefresh() {
		return previous
	}
	if !send(ipcRequest{
		Type: "subscribe",
		Events: []string{
			"workspace_changed", "workspace_updated", "window_opened",
			"window_closed", "window_moved", "window_urgent", "focus_changed",
			"layout_changed", "fullscreen_changed", "keyboard_layout_changed",
			"config_reloaded", "shutdown",
		},
	}, "subscribe") {
		return previous
	}

	refreshTicker := time.NewTicker(750 * time.Millisecond)
	defer refreshTicker.Stop()
	lastEventRefresh := time.Time{}

	publish := func() {
		if !haveState || !haveWindows {
			return
		}
		snapshot := WorkspaceSnapshot{
			Connected:      true,
			Current:        state.Workspace,
			Count:          state.WorkspaceCount,
			Workspaces:     append([]IPCWorkspace(nil), state.Workspaces...),
			Windows:        append([]IPCWindow(nil), windows...),
			Output:         state.Output,
			KeyboardLayout: state.KeyboardLayout,
			Tiling:         state.Tiling,
		}
		c.publish(snapshot)
		previous = snapshot
	}

	for {
		select {
		case <-c.stop:
			return previous
		case <-readErr:
			return previous
		case command := <-c.commands:
			if !send(ipcRequest{
				Type: "command", Command: command.name,
				Workspace: command.workspace,
			}, "command") {
				return previous
			}
		case <-refreshTicker.C:
			if !requestRefresh() {
				return previous
			}
		case raw := <-rawMessages:
			var msg ipcMessage
			if err := json.Unmarshal(raw, &msg); err != nil {
				slog.Warn("invalid HatWM IPC message", "error", err)
				continue
			}
			if msg.Type == "event" {
				if time.Since(lastEventRefresh) >= 40*time.Millisecond {
					if !requestRefresh() {
						return previous
					}
					lastEventRefresh = time.Now()
				}
				continue
			}

			var id int64
			if len(msg.ID) == 0 || json.Unmarshal(msg.ID, &id) != nil {
				continue
			}
			kind := pending[id]
			delete(pending, id)
			if msg.Success != nil && !*msg.Success {
				slog.Warn("HatWM IPC request rejected", "kind", kind, "error", msg.Error)
				continue
			}
			switch kind {
			case "hello":
				if msg.ProtocolVersion != 0 && msg.ProtocolVersion != panelIPCProtocolVersion {
					slog.Error("HatWM IPC protocol mismatch", "server", msg.ProtocolVersion, "panel", panelIPCProtocolVersion)
					return previous
				}
			case "state":
				if err := json.Unmarshal(msg.Result, &state); err != nil {
					slog.Warn("could not decode HatWM state", "error", err)
					continue
				}
				haveState = true
				publish()
			case "windows":
				if err := json.Unmarshal(msg.Result, &windows); err != nil {
					slog.Warn("could not decode HatWM windows", "error", err)
					continue
				}
				haveWindows = true
				publish()
			case "command":
				if !requestRefresh() {
					return previous
				}
			}
		}
	}
}

func (s WorkspaceSnapshot) Validate() error {
	if s.Count < 0 || s.Current < 0 {
		return fmt.Errorf("invalid workspace snapshot")
	}
	return nil
}
