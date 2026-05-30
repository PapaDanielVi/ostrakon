package commands

import (
	"testing"

	"github.com/PapaDanielVi/ostrakon/pkg/crypto"
)

func TestValidateMasterPassword(t *testing.T) {
	password := "testpassword123"
	hash := crypto.HashPassword(password)

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "valid password matches hash",
			password: password,
			hash:     hash,
			expected: true,
		},
		{
			name:     "invalid password mismatches hash",
			password: "wrongpassword",
			hash:     hash,
			expected: false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validated := crypto.ValidatePassword(tt.password, tt.hash)

			if validated != tt.expected {
				t.Errorf("ValidatePassword() = %v, want %v", validated, tt.expected)
			}
		})
	}
}