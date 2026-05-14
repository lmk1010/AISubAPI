package dto

import "testing"

func TestMaskCredentials(t *testing.T) {
	creds := map[string]any{
		"access_token":    "sk-1234567890abcdef",
		"refresh_token":   "rt-abcdefghij",
		"email":           "user@example.com",
		"organization_id": "org-123",
		"api_key":         "key-secret-value",
		"password":        "short",
	}

	masked := MaskCredentials(creds)

	// Sensitive keys should be masked
	if masked["access_token"] != "****cdef" {
		t.Errorf("access_token = %v, want ****cdef", masked["access_token"])
	}
	if masked["refresh_token"] != "****ghij" {
		t.Errorf("refresh_token = %v, want ****ghij", masked["refresh_token"])
	}
	if masked["api_key"] != "****alue" {
		t.Errorf("api_key = %v, want ****alue", masked["api_key"])
	}
	if masked["password"] != "****" {
		t.Errorf("password = %v, want ****", masked["password"])
	}

	// Non-sensitive keys should be preserved
	if masked["email"] != "user@example.com" {
		t.Errorf("email = %v, want user@example.com", masked["email"])
	}
	if masked["organization_id"] != "org-123" {
		t.Errorf("organization_id = %v, want org-123", masked["organization_id"])
	}

	// Original map should not be modified
	if creds["access_token"] != "sk-1234567890abcdef" {
		t.Error("original map was modified")
	}
}

func TestMaskCredentials_Nil(t *testing.T) {
	if MaskCredentials(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		key       string
		sensitive bool
	}{
		{"access_token", true},
		{"refresh_token", true},
		{"api_key", true},
		{"session_key", true},
		{"password", true},
		{"secret", true},
		{"token", true},
		{"my_custom_token", true},
		{"client_secret_id", true},
		{"email", false},
		{"organization_id", false},
		{"name", false},
		{"type", false},
	}
	for _, tt := range tests {
		if got := isSensitiveKey(tt.key); got != tt.sensitive {
			t.Errorf("isSensitiveKey(%q) = %v, want %v", tt.key, got, tt.sensitive)
		}
	}
}
