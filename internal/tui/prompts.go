package tui

import (
	"github.com/charmbracelet/huh"
)

// Confirm displays an interactive yes/no prompt and returns the user's choice.
func Confirm(message string) (bool, error) {
	var confirmed bool
	err := huh.NewConfirm().
		Title(message).
		Affirmative("Yes").
		Negative("No").
		Value(&confirmed).
		Run()
	if err != nil {
		return false, err
	}
	return confirmed, nil
}

// Input displays an interactive text input prompt and returns the entered value.
// If the user leaves the input empty, defaultValue is returned.
func Input(label string, defaultValue string) (string, error) {
	var value string
	input := huh.NewInput().
		Title(label).
		Value(&value)

	if defaultValue != "" {
		input = input.Placeholder(defaultValue)
	}

	err := input.Run()
	if err != nil {
		return "", err
	}

	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

// Password displays an interactive password input (masked) and returns the entered value.
func Password(label string) (string, error) {
	var value string
	err := huh.NewInput().
		Title(label).
		EchoMode(huh.EchoModePassword).
		Value(&value).
		Run()
	if err != nil {
		return "", err
	}
	return value, nil
}
