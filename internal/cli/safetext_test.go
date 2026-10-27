package cli

import "testing"

func TestSafeText(t *testing.T) {
	tests := []struct{ in, want string }{
		{"plain text é", "plain text é"},
		{"a\x1b[31mb", "a\\u001b[31mb"},
		{"a\r\n\tb", "a\\u000d\\u000a\\u0009b"},
	}
	for _, tt := range tests {
		if got := safeText(tt.in); got != tt.want {
			t.Errorf("safeText(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
