// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package terminal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"unicode/utf8"

	"golang.org/x/term"
)

type Term struct{}

type Terminal interface {
	PromptPassword(promptMsg string) (string, error)
}

// PromptPassword prompts user to enter password and then returns it
func (c *Term) PromptPassword(promptMsg string) (string, error) {
	fmt.Print(promptMsg)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return strings.TrimSpace(string(bytePassword)), nil
}

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

// ErrInterrupted is returned by PromptMasked when the user presses ^C.
var ErrInterrupted = errors.New("input interrupted")

// PromptMasked prompts for a secret and prints a * for each character received,
// so the user can see that typing or pasting worked. When stdin is not a
// terminal, it reads one line without echo handling.
func PromptMasked(promptMsg string) (string, error) {
	fmt.Print(promptMsg + " ")

	fd := int(syscall.Stdin)
	if !term.IsTerminal(fd) {
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return "", err
		}
		fmt.Println()
		return strings.TrimSpace(line), nil
	}

	state, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, state)

	value, err := readMasked(os.Stdin, os.Stdout)
	fmt.Print("\r\n")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

// readMasked reads raw terminal input from r until Enter and writes one * to w
// for each character. Backspace removes the last character, ^C returns
// ErrInterrupted, and escape sequences such as arrow keys are ignored.
func readMasked(r io.Reader, w io.Writer) (string, error) {
	var input []byte
	buf := make([]byte, 256)
	inEscape := false

	for {
		n, err := r.Read(buf)
		for _, b := range buf[:n] {
			if inEscape {
				// CSI sequences end with a byte in the range 0x40-0x7e.
				if b != '[' && b >= 0x40 && b <= 0x7e {
					inEscape = false
				}
				continue
			}

			switch {
			case b == '\r' || b == '\n':
				return string(input), nil
			case b == 0x03:
				return "", ErrInterrupted
			case b == 0x04 && len(input) == 0:
				return "", io.EOF
			case b == 0x7f || b == 0x08:
				if len(input) > 0 {
					_, size := utf8.DecodeLastRune(input)
					input = input[:len(input)-size]
					fmt.Fprint(w, "\b \b")
				}
			case b == 0x1b:
				inEscape = true
			case b < 0x20:
				// Ignore other control characters.
			default:
				input = append(input, b)
				// Print one * per character, not per byte of a multi-byte rune.
				if b < 0x80 || b >= 0xc0 {
					fmt.Fprint(w, "*")
				}
			}
		}

		if err != nil {
			if err == io.EOF && len(input) > 0 {
				return string(input), nil
			}
			return "", err
		}
	}
}
