// Package network discovers and connects to Wi-Fi and Ethernet networks through NetworkManager.
package network

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Network struct {
	SSID     string
	Signal   int
	Security string
	Active   bool
	Kind     Kind
	UUID     string
	Device   string
}

type Kind string

const (
	KindWiFi     Kind = "wifi"
	KindEthernet Kind = "ethernet"
)

func (n Network) Secured() bool {
	if n.Kind == KindEthernet {
		return false
	}
	security := strings.TrimSpace(n.Security)
	return security != "" && security != "--"
}

func (n Network) Wired() bool {
	return n.Kind == KindEthernet
}

type Result struct {
	SSID         string
	Disconnected bool
	Err          error
}

type request struct {
	network    Network
	disconnect bool
}

type Controller struct {
	requests chan request
	results  chan Result
}

func NewController() *Controller {
	controller := &Controller{
		requests: make(chan request, 1),
		results:  make(chan Result, 1),
	}
	go controller.run()
	return controller
}

func (c *Controller) Connect(network Network) {
	if c == nil || network.SSID == "" {
		return
	}
	c.enqueue(request{network: network})
}

func (c *Controller) Disconnect(network Network) {
	if c == nil {
		return
	}
	c.enqueue(request{network: network, disconnect: true})
}

func (c *Controller) enqueue(action request) {
	select {
	case c.requests <- action:
	default:
		select {
		case <-c.requests:
		default:
		}
		c.requests <- action
	}
}

func (c *Controller) Results() <-chan Result {
	if c == nil {
		return nil
	}
	return c.results
}

func (c *Controller) Close() {
	if c != nil {
		close(c.requests)
	}
}

func (c *Controller) run() {
	for action := range c.requests {
		if action.disconnect {
			c.sendResult(Result{
				SSID:         action.network.SSID,
				Disconnected: true,
				Err:          disconnect(action.network),
			})
			continue
		}
		err := connect(action.network, "")
		if err != nil && action.network.Secured() {
			password, promptErr := promptPassword(action.network.SSID)
			if promptErr == nil && password != "" {
				err = connect(action.network, password)
			} else if promptErr != nil {
				err = promptErr
			}
		}
		c.sendResult(Result{SSID: action.network.SSID, Err: err})
	}
}

func (c *Controller) sendResult(result Result) {
	select {
	case c.results <- result:
	default:
		select {
		case <-c.results:
		default:
		}
		c.results <- result
	}
}

func Scan(rescan bool) ([]Network, error) {
	rescanValue := "no"
	if rescan {
		rescanValue = "yes"
	}
	wifiOutput, wifiErr := runNmcli(
		"-t", "--escape", "yes",
		"-f", "IN-USE,SSID,SIGNAL,SECURITY,DEVICE",
		"device", "wifi", "list", "--rescan", rescanValue,
	)
	ethernetOutput, ethernetErr := runNmcli(
		"-t", "--escape", "yes",
		"-f", "NAME,UUID,TYPE,DEVICE", "connection", "show",
	)
	if wifiErr != nil && ethernetErr != nil {
		return nil, fmt.Errorf("Wi-Fi scan: %v; Ethernet profiles: %w",
			wifiErr, ethernetErr)
	}
	networks := parseNetworks(string(wifiOutput))
	networks = append(networks, parseEthernetConnections(string(ethernetOutput))...)
	sortNetworks(networks)
	return networks, nil
}

func connect(network Network, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var args []string
	if network.Wired() {
		if network.UUID != "" {
			args = []string{"connection", "up", "uuid", network.UUID}
		} else {
			args = []string{"connection", "up", "id", network.SSID}
		}
	} else {
		args = []string{"device", "wifi", "connect", network.SSID}
		if password != "" {
			args = append(args, "password", password)
		}
	}
	output, err := exec.CommandContext(ctx, "nmcli", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

func disconnect(network Network) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	device := network.Device
	if device == "" {
		status := exec.CommandContext(
			ctx, "nmcli", "-t", "--escape", "yes",
			"-f", "DEVICE,TYPE,STATE", "device", "status",
		)
		status.Env = append(os.Environ(), "LC_ALL=C")
		output, err := status.Output()
		if err != nil {
			return err
		}
		device = connectedDevice(string(output), network.Kind)
	}
	if device == "" {
		return fmt.Errorf("no connected network device")
	}

	command := exec.CommandContext(ctx, "nmcli", "device", "disconnect", device)
	command.Env = append(os.Environ(), "LC_ALL=C")
	commandOutput, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(commandOutput)), err)
	}
	return nil
}

func promptPassword(ssid string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	output, err := exec.CommandContext(
		ctx, "zenity", "--password", "--title=Connect to "+ssid,
	).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func parseNetworks(output string) []Network {
	bySSID := make(map[string]Network)
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		fields := splitEscaped(line, ':')
		if len(fields) < 4 || fields[1] == "" {
			continue
		}
		signal, _ := strconv.Atoi(fields[2])
		candidate := Network{
			SSID:     fields[1],
			Signal:   signal,
			Security: fields[3],
			Active:   fields[0] == "*",
			Kind:     KindWiFi,
		}
		if len(fields) >= 5 {
			candidate.Device = fields[4]
		}
		existing, exists := bySSID[candidate.SSID]
		if !exists || candidate.Active || candidate.Signal > existing.Signal {
			bySSID[candidate.SSID] = candidate
		}
	}
	networks := make([]Network, 0, len(bySSID))
	for _, network := range bySSID {
		networks = append(networks, network)
	}
	sortNetworks(networks)
	return networks
}

func parseEthernetConnections(output string) []Network {
	var networks []Network
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		fields := splitEscaped(line, ':')
		if len(fields) < 4 || fields[0] == "" ||
			(fields[2] != "802-3-ethernet" && fields[2] != "ethernet") {
			continue
		}
		networks = append(networks, Network{
			SSID:   fields[0],
			Signal: 100,
			Active: fields[3] != "" && fields[3] != "--",
			Kind:   KindEthernet,
			UUID:   fields[1],
			Device: fields[3],
		})
	}
	return networks
}

func sortNetworks(networks []Network) {
	sort.SliceStable(networks, func(i, j int) bool {
		if networks[i].Active != networks[j].Active {
			return networks[i].Active
		}
		return networks[i].Signal > networks[j].Signal
	})
}

func connectedDevice(output string, kind Kind) string {
	deviceType := "wifi"
	if kind == KindEthernet {
		deviceType = "ethernet"
	}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		fields := splitEscaped(line, ':')
		if len(fields) >= 3 && fields[0] != "" &&
			fields[1] == deviceType && fields[2] == "connected" {
			return fields[0]
		}
	}
	return ""
}

func runNmcli(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "nmcli", args...)
	command.Env = append(os.Environ(), "LC_ALL=C")
	return command.Output()
}

func splitEscaped(value string, separator byte) []string {
	var fields []string
	var field strings.Builder
	escaped := false
	for i := 0; i < len(value); i++ {
		char := value[i]
		if escaped {
			field.WriteByte(char)
			escaped = false
			continue
		}
		if char == '\\' {
			escaped = true
			continue
		}
		if char == separator {
			fields = append(fields, field.String())
			field.Reset()
			continue
		}
		field.WriteByte(char)
	}
	fields = append(fields, field.String())
	return fields
}
