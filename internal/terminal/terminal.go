// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package terminal

import (
	"bufio"
	"fmt"
	"os"
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

// InputPrompt receives a string value using the label
func InputPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}
