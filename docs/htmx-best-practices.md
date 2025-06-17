# HTMX Best Practices (Based on Official v2.0.4 Docs)

## ✅ RECOMMENDED: Use `hx-boost` for Simple Navigation

The **official HTMX documentation recommends `hx-boost`** as the simplest approach for basic navigation:

```html
<div hx-boost="true">
    <a href="/about">About</a>
    <a href="/contact">Contact</a>
    <a href="/donate">Donate</a>
</div>
```

**Why `hx-boost` is better:**
- ✅ **Progressive Enhancement**: Links work without JavaScript
- ✅ **Simplest implementation**: No server-side header checking needed
- ✅ **Automatic history management**: Built-in browser history support
- ✅ **Graceful degradation**: Falls back to normal page loads
- ✅ **Full page responses**: Server returns normal HTML pages

### Server-Side with `hx-boost` (SIMPLE)

```go
func (app *App) AboutHandler(c buffalo.Context) error {
    // Just render the full page - hx-boost handles everything!
    return c.Render(http.StatusOK, r.HTML("pages/about.plush.html"))
}
```

**No header checking needed!** HTMX automatically swaps the `<body>` content.

## ❌ AVOID: Manual Header Checking (Unless Necessary)

This pattern is more complex and error-prone:

```go
// ❌ More complex - only use if you need fine control
func (app *App) AboutHandler(c buffalo.Context) error {
    if c.Request().Header.Get("HX-Request") == "true" {
        return c.Render(http.StatusOK, rHTMX.HTML("partial.plush.html"))
    }
    return c.Render(http.StatusOK, r.HTML("full_page.plush.html"))
}
```

**Problems with this approach:**
- ❌ More server-side complexity
- ❌ Requires separate partial templates
- ❌ No progressive enhancement
- ❌ Breaks if JavaScript fails

## When to Use Each Approach

### Use `hx-boost` for:
- ✅ Simple navigation between pages
- ✅ Forms that should work without JS
- ✅ Applications that need accessibility
- ✅ Progressive enhancement

### Use explicit HTMX attributes for:
- 🔧 Loading content into specific containers
- 🔧 Complex form interactions
- 🔧 Real-time updates
- 🔧 Fine-grained control over swapping

## Progressive Enhancement Example

```html
<!-- ✅ GOOD: Works with and without JavaScript -->
<div hx-boost="true">
    <nav>
        <a href="/">Home</a>
        <a href="/about">About</a>
        <a href="/donate">Donate</a>
    </nav>
</div>

<!-- ❌ AVOID: Breaks without JavaScript -->
<nav>
    <button hx-get="/about" hx-target="#content">About</button>
</nav>
```

## Key Takeaways from HTMX Docs

1. **`hx-boost` is the recommended starting point** for most applications
2. **Progressive enhancement is a core principle** - always include `href` attributes
3. **Server-side complexity should be minimized** when possible
4. **Full HTML responses are preferred** over fragments when using `hx-boost`

## Our Implementation Fix

Based on the official docs, we should:

1. ✅ Use `hx-boost="true"` on navigation
2. ✅ Return full HTML pages from handlers
3. ✅ Remove complex header checking logic
4. ✅ Ensure all links have proper `href` attributes

This gives us the simplest, most robust HTMX implementation that follows official best practices.
