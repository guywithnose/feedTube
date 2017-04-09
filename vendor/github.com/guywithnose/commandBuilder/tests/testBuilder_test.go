package test

import (
	"os/exec"
	"syscall"
	"testing"

	"github.com/guywithnose/commandBuilder"
)

func Test(t *testing.T) {
	cb := &commandBuilder.Test{ExpectedCommands: []*commandBuilder.ExpectedCommand{commandBuilder.NewExpectedCommand("", "ls", "error", 12)}}
	_, err := cb.CreateCommand("", "ls").Output()
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if status.ExitStatus() == 12 {
				return
			}

			t.Fatalf("Generated code was expected to be 12, was %d", status.ExitStatus())
		}
	}

	t.Fatalf("Generated error was not an ExitError: %v", err)
}

func TestHelperProcess(*testing.T) {
	commandBuilder.ErrorCodeHelper()
}
