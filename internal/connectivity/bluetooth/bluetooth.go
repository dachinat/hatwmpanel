// Package bluetooth discovers and controls BlueZ devices through bluetoothctl.
package bluetooth

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type Device struct {
	Address   string
	Name      string
	Connected bool
}

type Result struct {
	Name         string
	Disconnected bool
	Err          error
}

type request struct {
	device     Device
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

func (c *Controller) Connect(device Device) {
	if c == nil || device.Address == "" {
		return
	}
	c.enqueue(request{device: device})
}

func (c *Controller) Disconnect(device Device) {
	if c == nil || device.Address == "" {
		return
	}
	c.enqueue(request{device: device, disconnect: true})
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

func (c *Controller) run() {
	for action := range c.requests {
		var err error
		if action.disconnect {
			_, err = run(15*time.Second, "disconnect", action.device.Address)
		} else {
			err = connect(action.device)
		}
		c.sendResult(Result{
			Name:         action.device.Name,
			Disconnected: action.disconnect,
			Err:          err,
		})
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

func Scan(discover bool) ([]Device, error) {
	if discover {
		// Discovery is reference-counted by BlueZ and stops when this short-lived
		// bluetoothctl client exits.
		_, _ = run(5*time.Second, "--timeout", "3", "scan", "on")
	}
	allOutput, err := run(5*time.Second, "devices")
	if err != nil {
		return nil, err
	}
	connectedOutput, connectedErr := run(5*time.Second, "devices", "Connected")
	if connectedErr != nil {
		connectedOutput = ""
	}
	return parseDevices(allOutput, connectedOutput), nil
}

func connect(device Device) error {
	if _, err := run(20*time.Second, "connect", device.Address); err == nil {
		return nil
	}
	if _, err := run(35*time.Second, "--timeout", "30", "pair", device.Address); err != nil {
		return err
	}
	_, _ = run(10*time.Second, "trust", device.Address)
	_, err := run(20*time.Second, "connect", device.Address)
	return err
}

func run(timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	output, err := exec.CommandContext(ctx, "bluetoothctl", args...).CombinedOutput()
	text := strings.TrimSpace(string(output))
	if err != nil {
		return text, fmt.Errorf("%s: %w", text, err)
	}
	lower := strings.ToLower(text)
	if strings.Contains(lower, "failed") || strings.Contains(lower, "not available") {
		return text, fmt.Errorf("%s", text)
	}
	return text, nil
}

func parseDevices(allOutput, connectedOutput string) []Device {
	connected := make(map[string]bool)
	for _, device := range parseDeviceLines(connectedOutput) {
		connected[device.Address] = true
	}
	devices := parseDeviceLines(allOutput)
	for i := range devices {
		devices[i].Connected = connected[devices[i].Address]
	}
	sort.SliceStable(devices, func(i, j int) bool {
		if devices[i].Connected != devices[j].Connected {
			return devices[i].Connected
		}
		return strings.ToLower(devices[i].Name) < strings.ToLower(devices[j].Name)
	})
	return devices
}

func parseDeviceLines(output string) []Device {
	seen := make(map[string]bool)
	var devices []Device
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 3 || fields[0] != "Device" || seen[fields[1]] {
			continue
		}
		seen[fields[1]] = true
		devices = append(devices, Device{
			Address: fields[1],
			Name:    strings.Join(fields[2:], " "),
		})
	}
	return devices
}
