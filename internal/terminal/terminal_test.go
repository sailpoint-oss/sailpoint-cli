package terminal

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestReadMasked(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		echo  string
	}{
		{"typed", "abc\r", "abc", "***"},
		{"pasted secret", strings.Repeat("x", 64) + "\r", strings.Repeat("x", 64), strings.Repeat("*", 64)},
		{"backspace", "abd\x7fc\r", "abc", "***\b \b*"},
		{"backspace on empty", "\x7fab\n", "ab", "**"},
		{"arrow keys ignored", "a\x1b[Db\r", "ab", "**"},
		{"multi-byte rune", "é\r", "é", "*"},
		{"eof after input", "abc", "abc", "***"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			got, err := readMasked(strings.NewReader(tt.input), &out)
			if err != nil {
				t.Fatalf("readMasked() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("readMasked() = %q, want %q", got, tt.want)
			}
			if out.String() != tt.echo {
				t.Errorf("echo = %q, want %q", out.String(), tt.echo)
			}
		})
	}
}

func TestReadMaskedInterrupt(t *testing.T) {
	_, err := readMasked(strings.NewReader("ab\x03"), io.Discard)
	if !errors.Is(err, ErrInterrupted) {
		t.Fatalf("readMasked() error = %v, want ErrInterrupted", err)
	}
}

func TestReadMaskedEmptyEOF(t *testing.T) {
	_, err := readMasked(strings.NewReader("\x04"), io.Discard)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("readMasked() error = %v, want io.EOF", err)
	}
}
