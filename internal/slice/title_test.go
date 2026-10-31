package slice

import "testing"

// TestValidateTitle checks the trimming and the empty title (0002 R31).
func TestValidateTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"plain", "Login with Google", "Login with Google", false},
		{"spaces at both ends", "  Login  ", "Login", false},
		{"tabs and line breaks at the ends", "\tLogin\n", "Login", false},
		{"inner spaces stay", "Submit  a spec", "Submit  a spec", false},
		{"not English", "ภาษาไทย", "ภาษาไทย", false},
		{"empty", "", "", true},
		{"only spaces", "   ", "", true},
		{"only white space", " \t\n ", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateTitle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateTitle(%q) error = %v, want error %v", tt.input, err, tt.wantErr)
			}
			if err != nil && err.Error() != "the title is empty" {
				t.Errorf("error = %q, want %q", err, "the title is empty")
			}
			if got != tt.want {
				t.Errorf("validateTitle(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
