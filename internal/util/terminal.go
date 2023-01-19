// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package util

import (
	"fmt"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// PromptPassword prompts user to enter password and then returns it
func PromptPassword(promptMsg string) (string, error) {
	fmt.Print(promptMsg)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return strings.TrimSpace(string(bytePassword)), nil
}
