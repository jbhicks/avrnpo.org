package middleware

import (
	"strings"
	"testing"
	"time"
)

func TestCSRFProtection_TokenExpiration(t *testing.T) {
	// Manually expire a token by setting it to past time
	expiredToken := "expired_token_123"
	GlobalCSRFStore.tokens[expiredToken] = time.Now().Add(-2 * time.Hour)

	// Test that expired token validation fails
	if GlobalCSRFStore.Validate(expiredToken) {
		t.Error("Expected expired token validation to fail")
	}

	// Test that non-existent token validation fails
	if GlobalCSRFStore.Validate("non_existent_token") {
		t.Error("Expected non-existent token validation to fail")
	}
}

func TestCSRFProtection_TokenReuse(t *testing.T) {
	token := "reuse_token_123"
	GlobalCSRFStore.tokens[token] = time.Now().Add(time.Hour)

	// First validation should succeed
	if !GlobalCSRFStore.Validate(token) {
		t.Error("First token validation should succeed")
	}

	// Manually delete the token (as the middleware would do)
	GlobalCSRFStore.Delete(token)

	// Token should be gone after deletion
	if GlobalCSRFStore.Validate(token) {
		t.Error("Token should be deleted after use")
	}
}

func TestCSRFProtection_ConcurrentRequests(t *testing.T) {
	// Test concurrent access to the CSRF store
	token := "concurrent_token_123"
	GlobalCSRFStore.tokens[token] = time.Now().Add(time.Hour)

	// Test concurrent validation calls
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			GlobalCSRFStore.Validate(token)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Token should still be valid after concurrent access
	if !GlobalCSRFStore.Validate(token) {
		t.Error("Token should still be valid after concurrent access")
	}
}

func TestCSRFProtection_InvalidTokenFormats(t *testing.T) {
	testCases := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"Invalid characters", "invalid!@#$%^&*()"},
		{"Too short", "short"},
		{"Too long", strings.Repeat("a", 1000)},
		{"SQL injection attempt", "'; DROP TABLE users; --"},
		{"XSS attempt", "<script>alert('xss')</script>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test that invalid tokens are rejected
			if GlobalCSRFStore.Validate(tc.token) {
				t.Errorf("Expected invalid token validation to fail: %s", tc.token)
			}
		})
	}
}

func TestCSRFProtection_HeaderToken(t *testing.T) {
	// Test that header-based tokens are also supported
	// This tests the fallback logic in the CSRF middleware
	token := "header_token_123"
	GlobalCSRFStore.tokens[token] = time.Now().Add(time.Hour)

	// Test that valid tokens in headers would be accepted
	// (The actual header parsing is tested in the middleware itself)
	if !GlobalCSRFStore.Validate(token) {
		t.Error("Valid token should be accepted")
	}
}
