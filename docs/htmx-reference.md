# HTMX Reference Documentation

This document contains key HTMX concepts and best practices for the AVR NPO project. Based on HTMX documentation as of June 2025.

## Core Concepts

### HTMX Attributes

#### Basic Request Attributes
- `hx-get="/url"` - Issues GET request to URL
- `hx-post="/url"` - Issues POST request to URL
- `hx-put="/url"` - Issues PUT request to URL
- `hx-patch="/url"` - Issues PATCH request to URL
- `hx-delete="/url"` - Issues DELETE request to URL

#### Targeting and Swapping
- `hx-target="#id"` - Target element for response content
- `hx-swap="innerHTML"` - How to swap content (default)
- `hx-swap="outerHTML"` - Replace entire target element
- `hx-swap="beforebegin"` - Insert before target
- `hx-swap="afterbegin"` - Insert at start of target
- `hx-swap="beforeend"` - Insert at end of target
- `hx-swap="afterend"` - Insert after target
- `hx-swap="delete"` - Delete target element
- `hx-swap="none"` - Don't swap content

#### Navigation and History
- `hx-push-url="true"` - Push URL to browser history
- `hx-push-url="/custom-url"` - Push custom URL to history
- `hx-replace-url="true"` - Replace current URL in history

#### Triggers
- `hx-trigger="click"` - Trigger on click (default for buttons/links)
- `hx-trigger="submit"` - Trigger on form submit (default for forms)
- `hx-trigger="change"` - Trigger on input change
- `hx-trigger="keyup"` - Trigger on key release
- `hx-trigger="load"` - Trigger when element loads
- `hx-trigger="every 2s"` - Trigger every 2 seconds
- `hx-trigger="click once"` - Trigger only once

#### Headers and Parameters
- `hx-headers='{"X-Custom": "value"}'` - Custom headers
- `hx-params="*"` - Include all form parameters
- `hx-params="param1,param2"` - Include specific parameters
- `hx-params="none"` - Include no parameters

#### Indicators and Loading
- `hx-indicator="#loading"` - Show loading indicator
- `hx-disabled-elt="this"` - Disable element during request

## Best Practices for Buffalo/Go Backend

### Response Patterns

#### Partial Templates
Return HTML fragments that match your target element:

```html
<!-- For hx-target="#content" hx-swap="innerHTML" -->
<article>
    <h2>New Content</h2>
    <p>This replaces the innerHTML of #content</p>
</article>
```

#### Full Page Responses
For `hx-push-url`, return complete page HTML or partial that includes navigation updates.

#### Error Handling
Return appropriate HTTP status codes:
- 200: Success with content
- 204: Success with no content
- 400: Client error
- 500: Server error

### Buffalo Action Patterns

```go
// Handle both HTMX and regular requests
func (app *App) SomeAction(c buffalo.Context) error {
    // Your logic here
    
    if c.Request().Header.Get("HX-Request") == "true" {
        // HTMX request - return partial
        return c.Render(http.StatusOK, r.HTML("partial.plush.html"))
    }
    
    // Regular request - return full page
    return c.Render(http.StatusOK, r.HTML("full_page.plush.html"))
}
```

## Common Patterns

### Navigation with History
```html
<nav>
    <a href="/home" hx-get="/home" hx-target="#main-content" hx-push-url="true">Home</a>
    <a href="/about" hx-get="/about" hx-target="#main-content" hx-push-url="true">About</a>
    <a href="/donate" hx-get="/donate" hx-target="#main-content" hx-push-url="true">Donate</a>
</nav>

<main id="main-content">
    <!-- Content swapped here -->
</main>
```

### Forms with HTMX
```html
<form hx-post="/api/contact" hx-target="#form-result">
    <input type="text" name="name" required>
    <input type="email" name="email" required>
    <button type="submit">Submit</button>
</form>
<div id="form-result"></div>
```

### Progressive Enhancement
Always include proper `href` attributes as fallbacks:

```html
<!-- Works with and without JavaScript -->
<a href="/donate" 
   hx-get="/donate" 
   hx-target="#main-content" 
   hx-push-url="true">
   Donate Now
</a>
```

### Loading States
```html
<button hx-post="/api/action" 
        hx-target="#result" 
        hx-indicator="#spinner"
        hx-disabled-elt="this">
    Submit
</button>
<div id="spinner" class="htmx-indicator">Loading...</div>
```

## Event Handling

### HTMX Events
Listen for HTMX events in JavaScript:

```javascript
// Before request
document.addEventListener('htmx:beforeRequest', function(event) {
    console.log('Request starting:', event.detail);
});

// After request
document.addEventListener('htmx:afterRequest', function(event) {
    console.log('Request completed:', event.detail);
});

// After content swap
document.addEventListener('htmx:afterSwap', function(event) {
    console.log('Content swapped:', event.detail);
    // Re-initialize JavaScript components
});

// After settle (animations complete)
document.addEventListener('htmx:afterSettle', function(event) {
    console.log('Content settled:', event.detail);
});
```

### Custom Events
Trigger custom events from server responses:

```html
<!-- Server response -->
<div hx-trigger="customEvent from:body">
    <!-- Content -->
</div>

<script>
// Trigger from JavaScript
document.body.dispatchEvent(new CustomEvent('customEvent'));
</script>
```

## Debugging

### HTMX Headers
HTMX sends these headers with requests:
- `HX-Request: true` - Indicates HTMX request
- `HX-Current-URL` - Current page URL
- `HX-Target` - ID of target element
- `HX-Trigger` - ID of triggered element

### Browser DevTools
- Network tab shows HTMX requests
- Look for `HX-Request` header
- Response should contain HTML fragments

### Logging
Enable HTMX logging:

```javascript
// Add to your main JavaScript
htmx.logAll();
```

## Security Considerations

### CSRF Protection
Include CSRF tokens in forms:

```html
<form hx-post="/api/action">
    <input type="hidden" name="csrf_token" value="{{ .CSRFToken }}">
    <!-- Other fields -->
</form>
```

### Content Security Policy
Allow HTMX inline handlers if needed:

```
Content-Security-Policy: script-src 'self' 'unsafe-inline';
```

### Validation
Always validate on server side, HTMX is just transport:

```go
func (app *App) APIAction(c buffalo.Context) error {
    // Validate request
    if c.Request().Header.Get("HX-Request") != "true" {
        return c.Error(http.StatusBadRequest, errors.New("invalid request"))
    }
    
    // Process request
    // Return response
}
```

## Performance Tips

### Minimize Response Size
Return only necessary HTML:

```html
<!-- Good: Minimal response -->
<article class="post">
    <h2>{{ .Title }}</h2>
    <p>{{ .Content }}</p>
</article>

<!-- Avoid: Unnecessary wrapper elements -->
```

### Cache Static Assets
Cache JavaScript and CSS:

```html
<script src="/js/htmx.min.js" cache-control="max-age=31536000"></script>
```

### Lazy Loading
Load content when needed:

```html
<div hx-get="/api/comments" hx-trigger="intersect">
    Loading comments...
</div>
```

## Common Issues and Solutions

### Navigation Not Working
Check for:
- Missing `hx-push-url="true"`
- Wrong `hx-target` selector
- Server not returning appropriate content

### Content Not Swapping
Verify:
- Target element exists
- Server returns 200 status
- Response contains valid HTML

### History Issues
Ensure:
- URLs are properly pushed with `hx-push-url`
- Server handles direct URL access
- Back button returns correct content

### JavaScript Not Working After Swap
Re-initialize after content changes:

```javascript
document.addEventListener('htmx:afterSwap', function() {
    // Re-bind event listeners
    // Re-initialize components
});
```

## Integration with Buffalo

### Template Structure
```
templates/
├── application.plush.html    # Main layout
├── pages/
│   ├── home.plush.html      # Full page
│   └── _home_partial.plush.html  # Partial for HTMX
```

### Action Pattern
```go
func (app *App) HomePage(c buffalo.Context) error {
    if c.Request().Header.Get("HX-Request") == "true" {
        return c.Render(http.StatusOK, r.HTML("pages/_home_partial.plush.html"))
    }
    return c.Render(http.StatusOK, r.HTML("pages/home.plush.html"))
}
```

This reference covers the essential HTMX concepts needed for the AVR NPO project. For complete documentation, visit https://htmx.org/docs/.
