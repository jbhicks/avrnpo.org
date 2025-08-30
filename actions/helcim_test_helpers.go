package actions

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

// AttachHelcimSignature generates a Helcim HMAC-SHA256 signature for the given body
// and returns the header value to set (format: sha256=<hex>)
func AttachHelcimSignature(body []byte) string {
	secret := os.Getenv("HELCIM_WEBHOOK_VERIFIER_TOKEN")
	if secret == "" {
		// Fallback to test secret for local tests
		secret = "test_verifier_token"
	}
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	sig := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("sha256=%s", sig)
}
