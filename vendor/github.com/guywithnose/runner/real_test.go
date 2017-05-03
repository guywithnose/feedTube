package runner_test

import (
	"os/exec"
	"testing"

	"github.com/guywithnose/runner"
	"github.com/stretchr/testify/assert"
)

func TestRealRunner(t *testing.T) {
	cb := &runner.Real{}
	command := cb.New("", "ls")
	assert.IsType(t, &exec.Cmd{}, command)
}
