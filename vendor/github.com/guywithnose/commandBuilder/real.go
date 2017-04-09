package commandBuilder

import "os/exec"

// Builder builds a command that can then be run
type Builder interface {
	CreateCommand(path string, command ...string) Command
}

// Command is used to run commands.  It is a sub-interface of os.exec.Cmd
type Command interface {
	Output() ([]byte, error)
	CombinedOutput() ([]byte, error)
}

// Real is the standard CommandBuilder that will return instances of os.exec.Cmd
type Real struct{}

// CreateCommand creates an instance os.exec.Cmd
func (w Real) CreateCommand(path string, command ...string) Command {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = path
	return cmd
}
