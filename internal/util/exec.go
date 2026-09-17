// Copyright (c) 2022, SailPoint Technologies, Inc. All rights reserved.

//go:build linux || darwin || dragonfly || freebsd || netbsd || openbsd
// +build linux darwin dragonfly freebsd netbsd openbsd

package util

import (
	"os/exec"
	"syscall"
)

// ExecCommand starts a command on non windows environments with Setpgid flag set
// to true and returns the started command so the caller can stop it later.
func ExecCommand(name string, arg ...string) (*exec.Cmd, error) {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}

// StopCommand terminates a command started by ExecCommand together with every
// process in its group, so children spawned by the command do not survive it.
func StopCommand(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	// The negative pid targets the process group ExecCommand created.
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil && err != syscall.ESRCH {
		return err
	}

	// The command exits with a signal, which Wait reports as an error.
	_, err := cmd.Process.Wait()

	return err
}
