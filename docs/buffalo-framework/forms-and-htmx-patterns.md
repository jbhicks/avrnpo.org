# Buffalo + HTMX Form Handling Patterns

## 🚨 CRITICAL: Form Submission URL Issues

**NEVER submit forms to API endpoints for user-facing pages!**

### ❌ The Problem: API Endpoint Form Submission

When you submit a regular HTML form to an API endpoint like `/api/donations/initialize`, you encounter the "URL stuck on POST route" issue:

1. **Form submits to POST `/api/donations/initialize`**
2. **Handler redirects to `/donate/payment`**
3. **Browser URL shows `/api/donations/initialize`** (not the redirect destination)
4. **Refreshing the page results in 404** (trying to GET a POST-only route)

### ✅ The Solution: Same Route for GET and POST with HTMX

Use the same route for both displaying and processing the form:

```html
<!-- ✅ CORRECT: Same route for GET and POST -->
<form method="post" action="/donate"
      hx-post="/donate" 
      hx-target="body" 
      hx-swap="outerHTML" 
      hx-push-url="true">
  <!-- Form fields -->
</form>
```

**Why this works:**
- **Same URL**: `/donate` handles both GET (show form) and POST (process form)
- **No redirect needed**: HTMX gets the result page directly
- **URL stays correct**: Browser URL always shows `/donate` or destination page
- **Progressive enhancement**: Works for all users regardless of JavaScript support

## 🎯 Required Form Handler Pattern

All form handlers MUST handle both GET and POST methods in the same handler:

```go
func MyFormHandler(c buffalo.Context) error {
    // Handle GET request - show the form
    if c.Request().Method == "GET" {
        // Set up form defaults
        c.Set("errors", nil)
        // ... other form setup
        return c.Render(http.StatusOK, r.HTML("pages/myform.plush.html"))
    }
    
    // Handle POST request - process form data
    // ... validation logic
    if errors.HasAny() {
        // Set error context for template
        c.Set("errors", errors)
        // Return form with errors (same for both HTMX and regular requests)
        return c.Render(http.StatusOK, r.HTML("pages/myform.plush.html"))
    }
    
    // Success case
    c.Flash().Add("success", "Success message")
    
    // For HTMX requests, return destination page directly
    if c.Request().Header.Get("HX-Request") == "true" {
        return c.Render(http.StatusOK, r.HTML("pages/success.plush.html"))
    }
    // For regular requests, redirect to destination
    return c.Redirect(http.StatusSeeOther, "/success")
}
```

## 📋 Form Implementation Checklist

### ✅ HTML Form Requirements

- [ ] **Action attribute**: Points to same route as the form page (not a separate submit route)
- [ ] **HTMX attributes**: Include `hx-post`, `hx-target="body"`, `hx-swap="outerHTML"`, `hx-push-url="true"`
- [ ] **Progressive enhancement**: Form works without JavaScript
- [ ] **Proper method**: Use `method="post"` for form submission

### ✅ Handler Requirements

- [ ] **Method support**: Handle both GET (show form) and POST (process form) in same handler
- [ ] **Error handling**: Return form page with errors for both request types
- [ ] **Success handling**: Return success page for HTMX, redirect for regular
- [ ] **Flash messages**: Use Buffalo flash messages for user feedback
- [ ] **URL management**: Same route for both form display and processing

### ✅ Route Configuration

- [ ] **Single route**: Same route handles both GET and POST (e.g., `/donate`)
- [ ] **Authorization**: Skip authorization for public forms (contact, donate)
- [ ] **Route naming**: Use descriptive names like `/contact`, `/donate`

## 🛠️ Implementation Examples

### Example 1: Contact Form

```html
<!-- Template: templates/pages/contact.plush.html -->
<form method="post" action="/contact"
      hx-post="/contact" 
      hx-target="body" 
      hx-swap="outerHTML" 
      hx-push-url="true">
  <label for="name">Name *
    <input type="text" id="name" name="name" required />
  </label>
  <label for="email">Email *
    <input type="email" id="email" name="email" required />
  </label>
  <label for="message">Message *
    <textarea id="message" name="message" required></textarea>
  </label>
  <button type="submit">Send Message</button>
</form>
```

```go
// Handler: actions/pages.go
func ContactHandler(c buffalo.Context) error {
    // Handle GET request - show the contact form
    if c.Request().Method == "GET" {
        // Set up form defaults
        return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
    }
    
    // Handle POST request - process the form
    name := c.Param("name")
    email := c.Param("email")
    message := c.Param("message")
    
    // Validation
    if name == "" || email == "" || message == "" {
        c.Flash().Add("error", "Please fill in all required fields.")
        return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
    }
    
    // Process form (send email, save to database, etc.)
    
    // Success
    c.Flash().Add("success", "Thank you for your message!")
    if c.Request().Header.Get("HX-Request") == "true" {
        return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
    }
    return c.Redirect(http.StatusSeeOther, "/contact")
}
```

```go
// Routes: actions/app.go
app.GET("/contact", ContactHandler)
app.POST("/contact", ContactHandler)
```

### Example 2: Donation Form

```html
<!-- Template: templates/pages/_donate_form.plush.html -->
<form id="donation-form" method="post" action="/donate"
      hx-post="/donate" 
      hx-target="body" 
      hx-swap="outerHTML" 
      hx-push-url="true">
  <!-- Donation form fields -->
  <button type="submit">Donate Now</button>
</form>
```

```go
// Handler: actions/pages.go
func DonateHandler(c buffalo.Context) error {
    // Handle GET request - show the donation form
    if c.Request().Method == "GET" {
        // Set up form defaults
        return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
    }
    
    // Handle POST request - process the donation
    // Validation and processing logic
    
    if errors.HasAny() {
        // Return form with errors
        return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
    }
    
    // Success - go to payment page
    if c.Request().Header.Get("HX-Request") == "true" {
        return c.Render(http.StatusOK, r.HTML("pages/donate_payment.plush.html"))
    }
    return c.Redirect(http.StatusSeeOther, "/donate/payment")
}
```

```go
// Routes: actions/app.go
app.GET("/donate", DonateHandler)
app.POST("/donate", DonateHandler)
```

## 🚨 Anti-Patterns to Avoid

### ❌ Don't: Submit forms to API endpoints

```html
<!-- ❌ BAD: This causes URL issues -->
<form method="post" action="/api/donations/initialize">
```

### ❌ Don't: Use separate routes for form submission

```html
<!-- ❌ BAD: Creates URL problems -->
<form method="post" action="/donate/submit">
```

### ❌ Don't: Use HTMX without progressive enhancement

```html
<!-- ❌ BAD: Breaks without JavaScript -->
<form hx-post="/submit" hx-target="#result">
```

### ❌ Don't: Use separate handlers for GET and POST

```go
// ❌ BAD: Creates routing complexity
func ShowFormHandler(c buffalo.Context) error {
    return c.Render(http.StatusOK, r.HTML("form.plush.html"))
}

func ProcessFormHandler(c buffalo.Context) error {
    // ... processing
    return c.Redirect(http.StatusSeeOther, "/success")
}
```

## 🎯 Best Practices Summary

1. **Single route pattern** - Use same route for both GET (show) and POST (process)
2. **Progressive enhancement first** - Always include `action` attribute with same route
3. **Method-based routing** - Handle GET and POST in the same handler function
4. **Proper URL management** - Use `hx-push-url="true"` to keep URLs in sync
5. **Full page responses** - Return complete pages for HTMX requests, not just fragments
6. **Flash message integration** - Use Buffalo's flash messages for user feedback
7. **Error handling consistency** - Same error handling pattern for both request types

## 🧪 Testing Your Forms

### Test Both Request Types

```bash
# Test regular form submission (same route)
curl -X POST http://localhost:3000/contact \
  -d "name=Test&email=test@example.com&message=Hello"

# Test HTMX form submission (same route)
curl -X POST http://localhost:3000/contact \
  -H "HX-Request: true" \
  -d "name=Test&email=test@example.com&message=Hello"
```

### Verify URL Behavior

1. **Submit form via HTMX** - URL should update to reflect current page
2. **Refresh page** - Should load the current page correctly (no 404)
3. **Direct navigation** - Typing URL directly should work
4. **Back button** - Browser back/forward should work correctly

### Check Progressive Enhancement

1. **Disable JavaScript** - Form should still work
2. **Enable JavaScript** - HTMX should enhance the experience
3. **Error scenarios** - Errors should display properly in both modes
4. **Success scenarios** - Success flow should work in both modes

This pattern ensures robust, accessible forms that work for all users while providing enhanced experience for modern browsers.