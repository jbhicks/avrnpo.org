# Dev Email Testing

This document covers two email testing scenarios:
1. Testing the contact form with real email delivery (DEV_EMAIL_OVERRIDE)
2. Running SMTP integration tests (EMAIL_INTEGRATION_TESTS)

## Testing Contact Form with Real Email (Recommended)

Use `DEV_EMAIL_OVERRIDE` to redirect all emails to your address for testing without affecting production recipients.

### Quick Setup

1. Edit `.env`:
```bash
EMAIL_ENABLED=true
DEV_EMAIL_OVERRIDE=your-email@example.com
```

2. Ensure valid SMTP credentials in `.env`:
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=noreply@avrnpo.org
FROM_NAME=American Veterans Rebuilding
```

3. Start the server:
```bash
make dev
```

4. Test the contact form at `http://localhost:8090/contact`

### What Happens

- All emails (contact notifications, donation receipts) redirect to `DEV_EMAIL_OVERRIDE`
- Original recipient shown in server logs: `[EMAIL_DEV] Original recipient: info@avrnpo.org`
- BCC recipients cleared in dev mode
- Server logs confirm redirect: `[EMAIL_DEV] Redirecting to: your-email@example.com`

### When to Use

- Testing contact form submissions
- Verifying email templates
- Debugging email delivery issues
- Testing without affecting production email addresses

## Running SMTP Integration Tests (Advanced)

These tests send real email and are disabled by default — do not run them in CI and do not commit credentials.

### Important Safety Notes

- These tests WILL SEND REAL EMAIL when enabled. Use a disposable or test recipient address.
- Never enable integration tests in CI.
- Do not commit your SMTP credentials or a .env file containing them.

### How the Tests Are Gated

- Integration tests that perform real network/email sends are skipped unless the following environment variable is set:
  - `EMAIL_INTEGRATION_TESTS=true`
- The integration test also requires a recipient address via:
  - `TEST_EMAIL_RECIPIENT=you@example.com`

### Run the Integration Test Locally

1. Create a local .env file (do NOT commit it) or export the variables in your shell.

Example .env (local only):

```bash
EMAIL_INTEGRATION_TESTS=true
TEST_EMAIL_RECIPIENT=you@example.com
EMAIL_ENABLED=true
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your_username
SMTP_PASSWORD=your_password
```

2. Export variables in your shell instead (example):

```bash
export EMAIL_INTEGRATION_TESTS=true
export TEST_EMAIL_RECIPIENT=you@example.com
export EMAIL_ENABLED=true
export SMTP_HOST=smtp.example.com
export SMTP_PORT=587
export SMTP_USERNAME=your_username
export SMTP_PASSWORD=your_password
```

3. Run the specific test (recommended to limit scope):

```bash
# Run only the single integration test in the services package
go test ./services -run TestEmailService_SendDonationReceipt_Gmail -v
```

### What to Do After Running

- Remove or unset the environment variables and local .env when finished.
- If you accidentally enabled the flag in CI, immediately remove it and re-run CI.

### Recommended Long-term Improvements

- Keep integration tests behind a build tag (eg. `//go:build integration`) so they never run with normal `go test`.
- Enforce a CI policy that `EMAIL_INTEGRATION_TESTS` is never set in CI.
- Use a dedicated test SMTP account (or Mailtrap-like service) for integration tests.

