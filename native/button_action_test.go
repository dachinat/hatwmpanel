package native

import "testing"

func TestActivateButtonReportsCommandResult(t *testing.T) {
	wait, err := activateButton(Module{Name: "success", Action: "exit 0"})
	if err != nil {
		t.Fatalf("activateButton(): %v", err)
	}
	if err := wait(); err != nil {
		t.Fatalf("successful button action: %v", err)
	}

	wait, err = activateButton(Module{Name: "failure", Action: "exit 7"})
	if err != nil {
		t.Fatalf("activateButton(): %v", err)
	}
	if err := wait(); err == nil {
		t.Fatal("failed button action returned no error")
	}
}
