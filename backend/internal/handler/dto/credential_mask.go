package dto

import "strings"

// sensitiveCredentialKeys lists credential keys that should be masked in list responses.
var sensitiveCredentialKeys = map[string]bool{
	"access_token":  true,
	"refresh_token": true,
	"api_key":       true,
	"session_key":   true,
	"token":         true,
	"secret":        true,
	"password":      true,
	"setup_token":   true,
}

// MaskCredentials returns a copy of the credentials map with sensitive values masked.
// Non-sensitive keys (e.g., "email", "organization_id") are preserved as-is.
// The original map is not modified.
func MaskCredentials(creds map[string]any) map[string]any {
	if creds == nil {
		return nil
	}
	masked := make(map[string]any, len(creds))
	for k, v := range creds {
		if isSensitiveKey(k) {
			if s, ok := v.(string); ok && s != "" {
				masked[k] = maskString(s)
			} else {
				masked[k] = "****"
			}
		} else {
			masked[k] = v
		}
	}
	return masked
}

// MaskAccountCredentialsInPlace masks the credentials in the Account DTO.
// Used for list responses where full credential visibility is not needed.
func MaskAccountCredentialsInPlace(acc *Account) {
	if acc == nil {
		return
	}
	acc.Credentials = MaskCredentials(acc.Credentials)
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	if sensitiveCredentialKeys[lower] {
		return true
	}
	// Also match keys containing "token", "secret", "key", "password"
	for _, substr := range []string{"token", "secret", "key", "password"} {
		if strings.Contains(lower, substr) {
			return true
		}
	}
	return false
}

// maskString shows the last 4 characters for values ≥8 chars, otherwise all asterisks.
func maskString(s string) string {
	if len(s) >= 8 {
		return "****" + s[len(s)-4:]
	}
	return "****"
}
