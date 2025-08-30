# Running Email Integration Tests (Dangerous)

This document explains how to run the optional SMTP/email integration tests. These tests send real email and are disabled by default — do not run them in CI and do not commit credentials.

Important safety notes

- These tests WILL SEND REAL EMAIL when enabled. Use a disposable or test recipient address.
- Never enable integration tests in CI.
- Do not commit your SMTP credentials or a .env file containing them.

How the tests are gated

- Integration tests that perform real network/email sends are skipped unless the following environment variable is set:
  - EMAIL_INTEGRATION_TESTS=true
- The integration test also requires a recipient address via:
  - TEST_EMAIL_RECIPIENT=you@example.com

Run the integration test locally

1. Create a local .env file (do NOT commit it) or export the variables in your shell.

Example .env (local only):

EMAIL_INTEGRATION_TESTS=true
TEST_EMAIL_RECIPIENT=you@example.com
EMAIL_ENABLED=true
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your_username
SMTP_PASSWORD=your_password

2. Export variables in your shell instead (example):

export EMAIL_INTEGRATION_TESTS=true
export TEST_EMAIL_RECIPIENT=you@example.com
export EMAIL_ENABLED=true
export SMTP_HOST=smtp.example.com
export SMTP_PORT=587
export SMTP_USERNAME=your_username
export SMTP_PASSWORD=your_password

3. Run the specific test (recommended to limit scope):

```bash
# Run only the single integration test in the services package
go test ./services -run TestEmailService_SendDonationReceipt_Gmail -v
```

What to do after running

- Remove or unset the environment variables and local .env when finished.
- If you accidentally enabled the flag in CI, immediately remove it and re-run CI.

Recommended long-term improvements

- Keep integration tests behind a build tag (eg. //go:build integration) so they never run with normal `go test`.
- Enforce a CI policy that EMAIL_INTEGRATION_TESTS is never set in CI.
- Use a dedicated test SMTP account (or Mailtrap-like service) for integration tests.

Contact

If you want, I can also:
- Add a build tag to move integration tests to the `integration` tag.
- Add a GitHub Actions workflow that prevents accidental runs in CI.

