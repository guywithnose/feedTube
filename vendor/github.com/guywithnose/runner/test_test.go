package runner_test

import (
	"errors"
	"os/exec"
	"syscall"
	"testing"

	"github.com/guywithnose/runner"
	"github.com/stretchr/testify/assert"
)

func TestTestRunner(t *testing.T) {
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{runner.NewExpectedCommand("", "ls", "error", 12)}}
	_, err := cb.New("", "ls").Output()
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

func TestTestRunnerUnexpectedCommandError(t *testing.T) {
	cb := &runner.Test{}
	command := cb.New("", "ls")
	_, _ = command.Output()
	_, _ = command.CombinedOutput()
	assert.Equal(t, []error{errors.New("More commands were run than expected.  Extra command: ls")}, cb.Errors)
}

func TestTestRunnerWrongPathError(t *testing.T) {
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{runner.NewExpectedCommand("/foo", "ls", "", 0)}}
	command := cb.New("/bar", "ls")
	_, _ = command.Output()
	_, _ = command.CombinedOutput()
	assert.Equal(t, []error{errors.New("Path /bar did not match expected path /foo")}, cb.Errors)
}

func TestTestRunnerWrongCommandError(t *testing.T) {
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{runner.NewExpectedCommand("", "ls", "", 0)}}
	command := cb.New("", "cat")
	_, _ = command.Output()
	_, _ = command.CombinedOutput()
	assert.Equal(t, []error{errors.New("Command 'cat' did not match expected command 'ls'")}, cb.Errors)
}

func TestTestRunnerRanCommandsAreRemoved(t *testing.T) {
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{runner.NewExpectedCommand("", "ls", "error", 12)}}
	_, _ = cb.New("", "ls").Output()
	assert.Equal(t, []*runner.ExpectedCommand{}, cb.ExpectedCommands)
}

func TestNegativeOneError(t *testing.T) {
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{runner.NewExpectedCommand("", "ls", "error", -1)}}
	_, err := cb.New("", "ls").CombinedOutput()
	assert.EqualError(t, err, "Error running command")
}

func TestClosure(t *testing.T) {
	expectedCommand := runner.NewExpectedCommand("", "ls", "", 0)
	closureRan := false
	expectedCommand.Closure = func(string) {
		closureRan = true
	}
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{expectedCommand}}
	_, _ = cb.New("", "ls").CombinedOutput()
	assert.True(t, closureRan)
}

func TestHelperProcess(*testing.T) {
	runner.ErrorCodeHelper()
}
