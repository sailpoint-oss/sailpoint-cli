// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

//go:build windows
// +build windows

package cmd

import (
	"os/exec"
	"syscall"
)

// ExecCommand runs commands on windows environment with CREATE_NEW_PROCESS_GROUP flag,
// equivalent to Setpgid in linux like environment
func ExecCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}

	return cmd.Start()
}
