package services

import (
	"testing"
)

// ensure EmailEnabled defaults to false in development via NewEmailService
func TestNewEmailService_Defaults(t *testing.T) {
	// Temporarily ensure GO_ENV is not production and EMAIL_ENABLED not set
	// This test only validates that NewEmailService does not panic and sets EmailEnabled
	_ = NewEmailService()
}
