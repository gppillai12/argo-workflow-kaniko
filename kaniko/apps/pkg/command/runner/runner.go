package runner

import (
	"io"
	"os/exec"

	errors "github.com/icelolly/go-errors"
)

// Runner is a simple struct for running commands, Runner implements CmdRunner interface
type Runner struct {
	StdOutWriter io.Writer
	StdErrWriter io.Writer
	WorkingDir   string
}

// Run runs a given command with a variable number of argmuments
func (r *Runner) Run(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)

	cmd.Dir = r.WorkingDir
	cmd.Stdout = r.StdOutWriter
	cmd.Stderr = r.StdErrWriter

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Output runs a command and returns the output to the caller
func (r *Runner) Output(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = r.WorkingDir

	return cmd.Output()
}

// SetDirectory sets the current working directory to use for command execution
func (r *Runner) SetDirectory(dir string) {
	r.WorkingDir = dir
}

// BinaryExists checks for the existence of the given command in user's system
func (r *Runner) BinaryExists(cmd string) (bool, error) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		return false, errors.Wrap(err, "binary not found").WithField("cmd", cmd)
	}
	return true, nil
}

// New creates a new Runner object
func New(stdOutWriter, stdErrWriter io.Writer, workingDir string) (*Runner, error) {
	if stdOutWriter == nil || stdErrWriter == nil {
		return nil, errors.New("expected non empty value for").WithFields("stdOutWriter", stdOutWriter, "stdErrWriter", stdErrWriter)
	}

	if workingDir == "" {
		return nil, errors.New("expected non empty string for workingDir")
	}

	return &Runner{
		StdOutWriter: stdOutWriter,
		StdErrWriter: stdErrWriter,
		WorkingDir:   workingDir,
	}, nil
}
