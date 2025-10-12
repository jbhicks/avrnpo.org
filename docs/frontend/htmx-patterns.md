# HTMX Best Practices (Based on Official v2.0.4 Docs)

## 🎯 SINGLE TEMPLATE ARCHITECTURE: The HTMX Way

**Our application follows the official HTMX "Single Template Architecture" pattern using `hx-boost`.**

### ✅ CORRECT PATTERN: Single Template with hx-boost

```go
// ✅ ALWAYS USE: Simple full page handlers
func handleDonate(e *core.RequestEvent) error {
    csrfToken, _ := middleware.GetCSRFToken(e)
    component := templates.DonatePage(csrfToken)
    return component.Render(context.Background(), e.Response())
}

func handleBlog(e *core.RequestEvent) error {
    posts, _ := getBlogPosts(app)
    component := templates.BlogPage(posts)
    return component.Render(context.Background(), e.Response())
}
```

**Key Benefits:**
- ✅ **Works with and without JavaScript**
- ✅ **Bookmarks work correctly**
- ✅ **Page refreshes work correctly**
- ✅ **No complex server-side header checking**
- ✅ **Progressive enhancement by default**

### ❌ ANTI-PATTERN: Header Checking (NEVER USE)

```go
// ❌ NEVER USE: This breaks bookmarks and direct URL access
func BadHandler(e *core.RequestEvent) error {
    if e.Request.Header.Get("HX-Request") == "true" {
        // Don't do this - returns partial without full page context
        component := templates.PartialComponent()
        return component.Render(context.Background(), e.Response())
    }
    component := templates.FullPage()
    return component.Render(context.Background(), e.Response())
}
```

**Why this is wrong:**
- ❌ Breaks when users bookmark URLs
- ❌ Breaks when users refresh pages
- ❌ Breaks when users access URLs directly
- ❌ Creates maintenance burden with duplicate templates
- ❌ Violates progressive enhancement principles

## ✅ RECOMMENDED: Use `hx-boost` for Navigation

The **official HTMX documentation recommends `hx-boost`** as the optimal approach:

```templ
// ✅ GOOD: Already implemented in base.templ
templ Base(title string, csrfToken string, contents templ.Component) {
    <!DOCTYPE html>
    <html lang="en">
    <body hx-boost="true">
        @Navigation()
        @contents
        @Footer()
    </body>
    </html>
}
```
```

**How `hx-boost` works:**
- ✅ **Intercepts all clicks** on `<a>` tags and form submissions
- ✅ **Makes AJAX requests** automatically
- ✅ **Swaps entire `<body>` content** with new page
- ✅ **Updates browser history** correctly
- ✅ **Graceful degradation** - falls back to normal navigation if JS fails

### Server-Side Implementation (SIMPLE)

```go
// ✅ Perfect for hx-boost - just return full pages
func handleAbout(e *core.RequestEvent) error {
    csrfToken, _ := middleware.GetCSRFToken(e)
    component := templates.AboutPage(csrfToken)
    return component.Render(context.Background(), e.Response())
}

func handleBlog(e *core.RequestEvent) error {
    posts, _ := getBlogPosts(app)
    component := templates.BlogPage(posts)
    return component.Render(context.Background(), e.Response())
}
```

**No header checking needed!** HTMX handles everything automatically.

## 🔧 SPECIFIC USE CASES: When to Use Explicit HTMX

Only use explicit HTMX attributes for specialized functionality:

### ✅ Form Validation & Updates
```html
<!-- ✅ GOOD: Specific form interaction -->
<button hx-patch="/donate/update-amount" 
        hx-include="closest form"
        hx-target="#donation-form-content">
    Update Amount
</button>
```

### ✅ Real-time Content Updates  
```html
<!-- ✅ GOOD: Loading content into specific containers -->
<div hx-get="/api/notifications" 
     hx-trigger="every 30s"
     hx-target="#notification-panel">
</div>
```

### ✅ Progressive Enhancement Forms
```html
<!-- ✅ GOOD: Works with and without JS -->
<form method="post" action="/contact" 
      hx-post="/contact" 
      hx-target="body" 
      hx-swap="outerHTML">
</form>
```

## 📋 IMPLEMENTATION CHECKLIST

### ✅ Current Status
- [x] **Global `hx-boost` enabled** in `templates/base.templ`
- [x] **Navigation links work** with and without JavaScript  
- [x] **Forms use progressive enhancement** patterns

### 🎯 Our Single Template Architecture

Every page handler follows this simple pattern:

```go
func handlePage(e *core.RequestEvent) error {
    csrfToken, _ := middleware.GetCSRFToken(e)
    // Always return full page - hx-boost handles the rest!
    component := templates.PageComponent(csrfToken)
    return component.Render(context.Background(), e.Response())
}
```

### 🚫 What We DON'T Do

- ❌ No `HX-Request` header checking
- ❌ No separate partial templates for HTMX vs direct access
- ❌ No complex conditional rendering logic
- ❌ No duplicate template maintenance

### 🔍 Template Architecture

**Full page templates** include:
- Complete HTML structure (`<html>`, `<head>`, `<body>`)
- Navigation (via `@Navigation()` component)
- Footer (via `@Footer()` component)
- Main content area with HTMX target containers

**HTMX automatically handles:**
- Extracting `<body>` content for navigation
- Updating browser history
- Managing loading states
- Graceful fallback for disabled JavaScript

## 🛡️ Progressive Enhancement Principles

1. **Links first**: Every action starts with a proper `<a href="">` or `<form action="">`
2. **HTMX second**: Add HTMX attributes to enhance the experience
3. **JavaScript optional**: Site works perfectly without JavaScript
4. **Accessibility built-in**: Screen readers and keyboard navigation work correctly

## ✨ Best Practices Summary

✅ **DO:**
- Use `hx-boost="true"` for navigation
- Return full HTML pages from all handlers
- Include proper `href` and `action` attributes
- Test functionality with JavaScript disabled

❌ **DON'T:**
- Check `HX-Request` headers
- Create separate partial templates
- Use `hx-get` without `href` fallbacks
- Assume JavaScript is enabled

## Fragment Contract (Donation Form)

Any fragment swapped into `#donation-form-content` must include a hidden `authenticity_token` input and be safe to `innerHTML` swap. This ensures CSRF protection is preserved across HTMX swaps and avoids relying on `hx-vals` to transmit the token. Prefer `hx-include="closest form"` and use the hidden input for form submissions.