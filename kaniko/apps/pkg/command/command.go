package command

import (
	"os/exec"

	errors "github.com/icelolly/go-errors"
)

// CmdRunner is an interface for running commands
type CmdRunner interface {
	Run(name string, arg ...string) error
	Output(name string, arg ...string) ([]byte, error)
	SetDirectory(dir string)
	BinaryExists(cmd string) (bool, error)
}

// BinaryExists checks for the existence of the given command in user's system
func BinaryExists(cmd string) (bool, error) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		return false, errors.Wrap(err, "binary not found").WithField("cmd", cmd)
	}
	return true, nil
}
