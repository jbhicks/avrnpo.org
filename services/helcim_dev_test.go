package services

import (
    "os"
    "testing"
)

func TestNewHelcimClient_Development_NoAPIKey(t *testing.T) {
    // Ensure environment is development and API key is unset
    os.Setenv("GO_ENV", "development")
    os.Unsetenv("HELCIM_PRIVATE_API_KEY")

    // Call NewHelcimClient; it should not panic and should return a non-nil client
    client := NewHelcimClient()
    if client == nil {
        t.Fatalf("expected non-nil HelcimClient in development without API key")
    }
}
