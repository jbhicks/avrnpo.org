# Testing guide — DB-backed tests and webhook tests

Overview
This document explains how to run the full test suite that depends on a database and how to run webhook-related tests (Helcim). It also documents new test helpers to make those tests deterministic.

Prerequisites
- Docker or a local Postgres instance configured per project `database.yml`.
- soda CLI (for migrations) or your standard migration workflow.
- Make sure you are not running long-lived dev servers while running tests (see [agents guide](./development/agents-guide.md) rules).

Environment variables commonly used in tests
- GO_ENV=test
- DATABASE_URL (or the project-specific env var used by pop)
- HELCIM_SECRET (used only for signature generation in tests; use a test value)
- HELCIM_TEST_BYPASS=true (optional; only if team prefers bypassing verification in test env)

Recommended local test sequence
1. Prepare DB
   - Run migrations for test DB:
     - soda migrate up --env test
     - Alternatively: make test will run migrations if configured in the Makefile
2. Run tests
   - Full suite (DB-backed): GO_ENV=test HELCIM_SECRET=test-secret make test
   - Fast suite (no DB): make test-fast
   - Single package: GO_ENV=test go test ./services -run TestHelcimSignature
3. Troubleshooting
   - If tests skip DB steps, check that migrations ran against the test DB. Run `soda migrate up --env test` manually.
   - If webhook tests fail due to signatures, either set HELCIM_SECRET and use the AttachHelcimSignature helper or enable test bypass (see below).

Test helpers (new)
- CreatePostForTest(t, tx, opts...) -> *models.Post
  - Creates and persists a post with sensible defaults:
    - PublishedAt = time.Now()
    - Title = "Test Post"
    - Slug = "test-post" (overridable)
  - Use in blog tests that previously assumed a post row existed.

- AttachHelcimSignature(req, payload, secret)
  - Computes the Helcim-compatible signature using the same algorithm as production code and attaches the header(s) to req.
  - Use in webhook tests to supply a valid signature matching HELCIM_SECRET.

Test-mode signature bypass (optional)
- For higher-level webhook handler tests (not unit tests of verification), you may choose to bypass signature verification when GO_ENV=test:
- Guard in production code: only bypass when GO_ENV == "test" or when a dedicated env var is set.
- Prefer using a signer helper (AttachHelcimSignature) when you want to test verification logic itself.

Idempotency key testing
- New comprehensive test suite in `services/helcim_test.go`:
  - `TestIdempotencyKey_UUIDFormat`: Verifies UUID v4 format (36 chars, hyphens)
  - `TestIdempotencyKey_Uniqueness`: Ensures 1000 generated UUIDs are unique
  - `TestProcessPayment_IdempotencyKeyGeneration`: Tests header presence and body exclusion
  - `TestCreatePaymentPlan_IdempotencyKeyGeneration`: Tests payment plan idempotency
  - `TestCreateSubscription_IdempotencyKeyGeneration`: Tests subscription idempotency
  - `TestPaymentAPIRequest_NoIdempotencyKeyField`: Ensures struct doesn't have body field
- All tests verify Helcim API compliance: header-based UUID v4 keys, no body inclusion

Unit tests to add
- Signature verification: valid signature, invalid signature, missing header, test-bypass case.
- Idempotency key implementation: UUID generation, header presence, body exclusion, uniqueness.
- Template prewarm: fails if partials missing.

Best practices
- Add fixtures in test setup rather than relying on shared DB state.
- Keep bypass logic strictly tied to test environment to avoid accidental production behavior.
- Centralize fixture creation to reduce duplication and make teardown easier.
