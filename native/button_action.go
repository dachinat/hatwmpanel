package native

import (
	"fmt"
	"os/exec"
)

// activateButton starts a configured button action and returns a wait function for
// asynchronous exit reporting.
func activateButton(module Module) (func() error, error) {
	command := exec.Command("sh", "-c", module.Action)
	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("start %q: %w", module.Name, err)
	}
	return command.Wait, nil
}
