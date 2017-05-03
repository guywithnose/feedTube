package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
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
	actualCommand   string
}

// ExpectedCommand represents a command that will be handled by a TestCommand
type ExpectedCommand struct {
	commandRegex *regexp.Regexp
	output       []byte
	path         string
	error        error
	Closure      func(string)
}

// New returns a TestCommand
func (testBuilder *Test) New(path string, command ...string) Command {
	var expectedCommand *ExpectedCommand
	commandString := strings.Join(command, " ")
	if len(testBuilder.ExpectedCommands) == 0 {
		testBuilder.Errors = append(testBuilder.Errors, fmt.Errorf("More commands were run than expected.  Extra command: %s", commandString))
	} else {
		expectedCommand = testBuilder.ExpectedCommands[0]
		if expectedCommand.path != path {
			testBuilder.Errors = append(testBuilder.Errors, fmt.Errorf("Path %s did not match expected path %s", path, expectedCommand.path))
		} else if !expectedCommand.commandRegex.MatchString(commandString) {
			testBuilder.Errors = append(testBuilder.Errors, fmt.Errorf("Command '%s' did not match expected command '%s'", commandString, expectedCommand.commandRegex))
		} else {
			testBuilder.ExpectedCommands = testBuilder.ExpectedCommands[1:]
		}
	}

	return TestCommand{cmdBuilder: testBuilder, Dir: path, expectedCommand: expectedCommand, actualCommand: commandString}
}

// Output returns the expected mock data
func (cmd TestCommand) Output() ([]byte, error) {
	return cmd.run()
}

// CombinedOutput returns the expected mock data
func (cmd TestCommand) CombinedOutput() ([]byte, error) {
	return cmd.run()
}

func (cmd TestCommand) run() ([]byte, error) {
	if cmd.expectedCommand == nil {
		return nil, nil
	}

	if cmd.expectedCommand.Closure != nil {
		cmd.expectedCommand.Closure(cmd.actualCommand)
	}

	return cmd.expectedCommand.output, cmd.expectedCommand.error
}

// NewExpectedCommand returns a new ExpectedCommand
func NewExpectedCommand(path, command, output string, exitCode int) *ExpectedCommand {
	commandRegex, err := regexp.Compile(fmt.Sprintf("^%s$", command))
	if err != nil {
		panic(err)
	}

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
		commandRegex: commandRegex,
		output:       []byte(output),
		path:         path,
		error:        err,
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
