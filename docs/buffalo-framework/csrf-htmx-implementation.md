# CSRF + HTMX Implementation Guide

## Overview

This document outlines the correct implementation pattern for CSRF protection with HTMX in Buffalo applications. The previous implementation suffered from architectural flaws and security issues that have been resolved.

## Core Principles

### 1. Trust Buffalo's Built-in CSRF Middleware

**DO NOT** create custom CSRF middleware or token management. Buffalo's `csrf.New` middleware handles everything correctly:

```go
// ✅ CORRECT: Use Buffalo's built-in CSRF middleware
app.Use(csrf.New)
```

### 2. HTMX Automatically Includes Form Fields

HTMX automatically includes all form fields, including hidden CSRF tokens, in AJAX requests. No custom JavaScript token handling is required.

```javascript
// ❌ WRONG: Custom token handling
document.addEventListener('htmx:configRequest', function(event) {
    // Complex token management code...
});

// ✅ CORRECT: Let HTMX handle it automatically
// No custom CSRF code needed!
```

### 3. Combined Handlers for HTMX Compatibility

Use single handlers that support both GET and POST methods for HTMX compatibility:

```go
// ✅ CORRECT: Combined handler
func ContactHandler(c buffalo.Context) error {
    if c.Request().Method == "GET" {
        return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
    }
    // POST logic...
}
```

## Implementation Pattern

### 1. Form Structure

```html
<form method="post" action="/contact"
      hx-post="/contact" hx-target="body" hx-swap="outerHTML">
  <!-- CSRF token automatically included by Buffalo -->
  <input type="hidden" name="authenticity_token" value="<%= authenticity_token %>" />

  <label for="email">
    Email *
    <input type="email" id="email" name="email" required />
  </label>

  <button type="submit">Send Message</button>
</form>
```

### 2. Server-side Validation

```go
func ContactHandler(c buffalo.Context) error {
    if c.Request().Method == "GET" {
        return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
    }

    // Sanitize and validate input
    email := SanitizeInput(c.Param("email"))

    if err := ValidateEmail(email); err != nil {
        c.Flash().Add("error", err.Error())
        return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
    }

    // Process form...
}
```

### 3. Validation Utilities

```go
// Secure email validation
func ValidateEmail(email string) error {
    if len(email) == 0 {
        return fmt.Errorf("email is required")
    }

    if len(email) > 254 { // RFC 5321 limit
        return fmt.Errorf("email address is too long")
    }

    if !emailRegex.MatchString(email) {
        return fmt.Errorf("please enter a valid email address")
    }

    return nil
}

// Input sanitization
func SanitizeInput(input string) string {
    // Remove dangerous characters
    input = strings.Map(func(r rune) rune {
        if r < 32 && r != 9 && r != 10 && r != 13 {
            return -1
        }
        return r
    }, input)

    return strings.TrimSpace(input)
}
```

## Security Considerations

### 1. Input Validation

- **Always validate** user input on the server
- **Never trust** client-side validation alone
- **Sanitize** input to prevent injection attacks
- **Enforce** reasonable length limits

### 2. CSRF Protection

- **Use Buffalo's built-in** CSRF middleware
- **Include CSRF tokens** in all forms
- **Validate tokens** on all state-changing requests
- **Skip CSRF only** for legitimate API endpoints

### 3. Session Management

- **Use Buffalo's session** management
- **Don't create custom** session handling
- **Maintain session consistency** across requests

## Common Mistakes to Avoid

### 1. Custom CSRF Middleware

```go
// ❌ WRONG: Don't do this
func CSRFTokenMiddleware() buffalo.MiddlewareFunc {
    // Custom token management...
}
```

### 2. Complex JavaScript Token Handling

```javascript
// ❌ WRONG: Don't do this
window.CSRFUtils = {
    getToken: function() { /* complex logic */ }
};
```

### 3. Poor Email Validation

```go
// ❌ WRONG: Don't do this
if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
    // Invalid validation
}
```

### 4. Incorrect HTML Structure

```html
<!-- ❌ WRONG: Don't do this -->
<label for="email">Email *
  <input type="email" id="email" name="email" required />
</label>

<!-- ✅ CORRECT: Do this -->
<label for="email">
  Email *
  <input type="email" id="email" name="email" required />
</label>
```

## Testing Strategy

### 1. Unit Tests

```go
func TestContactHandler(t *testing.T) {
    // Test validation functions
    err := ValidateEmail("test@example.com")
    require.NoError(t, err)

    err = ValidateEmail("invalid-email")
    require.Error(t, err)
}
```

### 2. Integration Tests

```go
func TestCSRFProtection(t *testing.T) {
    app := buffalo.New(buffalo.Options{Env: "test"})

    // Test without CSRF token
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/contact", strings.NewReader("email=test"))
    app.ServeHTTP(w, req)
    require.Equal(t, 403, w.Code) // Should be blocked
}
```

### 3. HTMX Integration Tests

```go
func TestHTMXContactForm(t *testing.T) {
    // Test with proper session and CSRF token
    // Verify HTMX requests work correctly
}
```

## Performance Considerations

### 1. Middleware Order

```go
// Optimal middleware order
app.Use(popmw.Transaction(models.DB))  // Database first
app.Use(translations())                // Translations
app.Use(csrf.New)                     // CSRF protection
// Route handlers
```

### 2. Validation Efficiency

- **Validate early** to fail fast
- **Cache compiled regex** patterns
- **Use efficient string operations**

## Maintenance Guidelines

### 1. Code Reviews

- **Require** security review for form handlers
- **Verify** CSRF token inclusion in new forms
- **Check** input validation completeness

### 2. Documentation

- **Document** security decisions
- **Maintain** this implementation guide
- **Update** when Buffalo or HTMX versions change

### 3. Monitoring

- **Log** CSRF validation failures
- **Monitor** for unusual patterns
- **Alert** on security events

## Migration from Previous Implementation

### 1. Remove Custom Code

```bash
# Remove these files/functions:
# - CSRFTokenMiddleware()
# - CSRFDebugMiddleware()
# - Custom token generation functions
# - Complex JavaScript CSRF handling
```

### 2. Update Forms

```html
<!-- Update form structure -->
<label for="email">
  Email *
  <input type="email" id="email" name="email" required />
</label>
```

### 3. Simplify JavaScript

```javascript
// Replace complex CSRF code with simple HTMX setup
document.addEventListener('DOMContentLoaded', function() {
    // Basic HTMX enhancements only
});
```

## Conclusion

The correct CSRF + HTMX implementation follows these principles:

1. **Trust Buffalo's built-in security**
2. **Let HTMX handle form submissions automatically**
3. **Validate input securely on the server**
4. **Keep client-side code simple**
5. **Follow established patterns**

This approach provides robust security while maintaining simplicity and maintainability.