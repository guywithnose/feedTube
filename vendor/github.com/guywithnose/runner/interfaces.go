package runner

// Builder builds a command that can then be run
type Builder interface {
	New(path string, command ...string) Command
}

// Command is used to run commands.  It is a sub-interface of os.exec.Cmd
type Command interface {
	Output() ([]byte, error)
	CombinedOutput() ([]byte, error)
}
