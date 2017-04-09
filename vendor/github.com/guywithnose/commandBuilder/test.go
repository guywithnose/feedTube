package commandBuilder

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Test is used for testing code that runs system commands without actually running the commands
type Test struct {
	ExpectedCommands []*ExpectedCommand
	Errors           []error
}

// TestCommand emulates an os.exec.Cmd, but returns mock data
type TestCommand struct {
	cmdBuilder      *Test
	Dir             string
	expectedCommand *ExpectedCommand
}

// ExpectedCommand represents a command that will be handled by a TestCommand
type ExpectedCommand struct {
	command string
	output  []byte
	path    string
	error   error
}

// CreateCommand returns a TestCommand
func (testBuilder *Test) CreateCommand(path string, command ...string) Command {
	var expectedCommand *ExpectedCommand
	commandString := strings.Join(command, " ")
	if len(testBuilder.ExpectedCommands) == 0 {
		testBuilder.Errors = append(testBuilder.Errors, fmt.Errorf("More commands were run than expected.  Extra command: %s", commandString))
	} else {
		expectedCommand = testBuilder.ExpectedCommands[0]
		if expectedCommand.path != path {
			testBuilder.Errors = append(testBuilder.Errors, fmt.Errorf("Path %s did not match expected path %s", path, expectedCommand.path))
		} else if expectedCommand.command != commandString {
			testBuilder.Errors = append(testBuilder.Errors, fmt.Errorf("Command '%s' did not match expected command '%s'", commandString, expectedCommand.command))
		} else {
			testBuilder.ExpectedCommands = testBuilder.ExpectedCommands[1:]
		}
	}

	return TestCommand{cmdBuilder: testBuilder, Dir: path, expectedCommand: expectedCommand}
}

// Output returns the expected mock data
func (cmd TestCommand) Output() ([]byte, error) {
	if cmd.expectedCommand == nil {
		return nil, nil
	}

	return cmd.expectedCommand.output, cmd.expectedCommand.error
}

// CombinedOutput returns the expected mock data
func (cmd TestCommand) CombinedOutput() ([]byte, error) {
	if cmd.expectedCommand == nil {
		return nil, nil
	}

	return cmd.expectedCommand.output, cmd.expectedCommand.error
}

// NewExpectedCommand returns a new ExpectedCommand
func NewExpectedCommand(path, command, output string, exitCode int) *ExpectedCommand {
	var err error
	if exitCode == -1 {
		err = errors.New("Error running command")
	} else if exitCode != 0 {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", strconv.Itoa(exitCode))
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		err = cmd.Run()
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitErr.Stderr = []byte(output)
			err = exitErr
		}
	}

	return &ExpectedCommand{
		command: command,
		output:  []byte(output),
		path:    path,
		error:   err,
	}
}

// ErrorCodeHelper exits with a specified error code
// This is used in tests that require a command to return an error code other than 0
// To use this the test must include a test like this:
//
// func TestHelperProcess(*testing.T) {
//     commandBuilder.ErrorCodeHelper()
// }
func ErrorCodeHelper() {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	code, err := strconv.Atoi(os.Args[3])
	if err != nil {
		code = 1
	}

	defer os.Exit(code)
}
