// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package terminal

import (
	"fmt"
	"strings"
	"syscall"

	"golang.org/x/term"
)

type Term struct{}

type Terminal interface {
	PromptPassword(promptMsg string) (string, error)
}

func (c *Term) PromptPassword(promptMsg string) (string, error) {
	fmt.Print(promptMsg)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return strings.TrimSpace(string(bytePassword)), nil
}
