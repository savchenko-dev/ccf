//go:build windows

package main

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
)

func execClaude(args []string) error {
	path, err := exec.LookPath("claude")
	if err != nil {
		return err
	}
	cmd := exec.Command(path, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	signal.Ignore(os.Interrupt)
	err = cmd.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}
	if err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
