package config

import "testing"

func TestConfirmationCodeFromID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "uuid uses last eight characters",
			id:   "12345678-90ab-cdef-1234-567890abcdef",
			want: "90AB-CDEF",
		},
		{
			name: "short id is uppercased",
			id:   "abc123",
			want: "ABC123",
		},
		{
			name: "whitespace is ignored",
			id:   "  12345678-90ab-cdef-1234-567890abcdef  ",
			want: "90AB-CDEF",
		},
		{
			name: "empty id stays empty",
			id:   "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := confirmationCodeFromID(tt.id); got != tt.want {
				t.Fatalf("confirmationCodeFromID(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
