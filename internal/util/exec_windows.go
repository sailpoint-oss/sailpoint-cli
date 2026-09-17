// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

//go:build windows
// +build windows

package util

import (
	"os/exec"
	"syscall"
)

// ExecCommand starts a command on windows environments with the
// CREATE_NEW_PROCESS_GROUP flag, equivalent to Setpgid in linux like
// environments, and returns the started command so the caller can stop it later.
func ExecCommand(name string, arg ...string) (*exec.Cmd, error) {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}

// StopCommand terminates a command started by ExecCommand.
func StopCommand(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	if err := cmd.Process.Kill(); err != nil {
		return err
	}

	_, err := cmd.Process.Wait()

	return err
}
