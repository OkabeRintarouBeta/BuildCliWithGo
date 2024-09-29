package main

import (
	"fmt"
	"os/exec"
)

type step struct {
	name    string   // step name
	exe     string   // executable name of the external tool to execute
	args    []string // arguments for the executable
	message string   // output message in case of success
	proj    string   // target project on which to execute the task
}

func NewStep(name, exe, message, proj string, args []string) step {
	return step{
		name:    name,
		exe:     exe,
		args:    args,
		message: message,
		proj:    proj,
	}
}

func (s step) Execute() (string, error) {
	cmd := exec.Command(s.exe, s.args...)
	cmd.Dir = s.proj
	if err := cmd.Run(); err != nil {
		return "", &stepError{step: s.name, msg: fmt.Sprintf("failed to execute %s", s.name), cause: err}
	}
	return s.message, nil
}
