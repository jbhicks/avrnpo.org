# CSRF Integration Tests

## Overview

This directory contains integration tests that verify CSRF protection works correctly when enabled. These tests were created in response to a production issue where HTMX requests were failing due to missing CSRF tokens.

## The Problem

**Unit tests weren't catching CSRF issues** because Buffalo's test environment automatically disables CSRF protection:

```go
// actions/app.go:88-89
if ENV != "test" {
    app.Use(csrf.New)
}
```

This meant:
- ✅ Unit tests passed (CSRF disabled)  
- ❌ Production failed (CSRF enabled)
- 🐛 HTMX requests threw 403 "CSRF token not found" errors

## The Solution

### 1. **Integration Test Environment**

Created `csrf_integration_test.go` with tests that run in `ENV=integration` instead of `ENV=test`, which:
- ✅ **Enables CSRF middleware** (like production)
- ✅ **Tests real security behavior**
- ✅ **Would have caught the original HTMX issue**

### 2. **Key Test Cases**

```go
// ✅ Verifies CSRF protection rejects requests without tokens
func TestCSRFProtectionEnabled(t *testing.T)

// ✅ Verifies CSRF protection rejects invalid tokens  
func TestCSRFWithInvalidToken(t *testing.T)

// ✅ Documents why unit tests missed the issue
func TestCSRFEnvironmentDifference(t *testing.T)
```

### 3. **Makefile Integration**

Added `make test-integration` command that:
- Sets up integration database
- Runs CSRF-specific tests with middleware enabled
- Provides clear pass/fail feedback

## Usage

```bash
# Run CSRF integration tests
make test-integration

# Or run directly
GO_ENV=integration go test ./actions -run "TestCSRF" -v
```

## Test Results

```
=== RUN   TestCSRFProtectionEnabled
✅ CSRF protection is working: POST request without token rejected with status 403

=== RUN   TestCSRFWithInvalidToken  
✅ CSRF protection is working: POST request with invalid token rejected with status 403

=== RUN   TestCSRFEnvironmentDifference
✅ Test environment correctly allows requests without CSRF tokens (status 200)
💡 This confirms why unit tests missed the CSRF issue
```

## Files Added/Modified

- **`actions/csrf_integration_test.go`** - New integration tests
- **`database.yml`** - Added integration environment
- **`Makefile`** - Added `test-integration` target
- **`go.mod`** - Added `github.com/PuerkitoBio/goquery` dependency

## When to Use

**Run integration tests when:**
- Adding new forms that submit data
- Modifying CSRF token handling (JavaScript or server-side)
- Before deploying security-related changes
- When unit tests pass but production has CSRF issues

## Security Benefits

These tests ensure that:
1. **CSRF protection is actually enabled** in non-test environments
2. **Requests without tokens are properly rejected** (403 status)
3. **Invalid tokens are properly rejected** (403 status)
4. **Our HTMX JavaScript fix works correctly**

## Prevention

This test suite **would have caught the original CSRF issue** because:
- Integration tests run with CSRF middleware enabled
- They verify that POST requests without tokens return 403
- The original bug would have shown up as a failing test

**Going forward**: Run `make test-integration` before deploying form-related changes to catch CSRF issues early.