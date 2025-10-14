# Templ + Progressive Enhancement Architecture

This document outlines our frontend architecture using Templ templates with HTMX and progressive JavaScript enhancement.

## Architecture Overview

We use a **component-based architecture** where:
- Each route renders complete Templ components
- HTMX provides progressive enhancement for better UX
- Standard HTML forms work without JavaScript
- Components are type-safe and compiled to Go code

## Core Principles

### 1. HTML-First Approach
All functionality must work with basic HTML and forms:

```templ
// Works without JavaScript
templ ContactForm(csrfToken string) {
    <form method="POST" action="/contact">
        <input type="hidden" name="csrf_token" value={ csrfToken }/>
        <input type="text" name="name" required/>
        <input type="email" name="email" required/>
        <textarea name="message" required></textarea>
        <button type="submit">Send Message</button>
    </form>
}
```

### 2. Component-Based Templates
Each view is built from composable Templ components:

```templ
package templates

templ ContactPage(csrfToken string) {
    @Base("Contact Us", csrfToken, contactContent(csrfToken))
}

templ contactContent(csrfToken string) {
    <main class="container">
        <h1>Contact Us</h1>
        @ContactForm(csrfToken)
    </main>
}
```

### 3. Progressive Enhancement with HTMX
Add HTMX attributes to improve UX while maintaining fallback:

```templ
templ DonationForm(csrfToken string) {
    <form 
        method="POST" 
        action="/donate/process"
        hx-post="/donate/process"
        hx-target="#donation-result"
        hx-indicator="#loading">
        
        <input type="hidden" name="csrf_token" value={ csrfToken }/>
        
        <label for="amount">Donation Amount</label>
        <input type="number" name="amount" id="amount" required/>
        
        <button type="submit">Donate Now</button>
        <div id="loading" class="htmx-indicator">Processing...</div>
    </form>
    
    <div id="donation-result"></div>
}
```

## Template Patterns

### Base Layout
```templ
package templates

templ Base(title string, csrfToken string, contents templ.Component) {
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8"/>
        <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
        if csrfToken != "" {
            <meta name="csrf-token" content={ csrfToken }/>
        }
        <title>{ title } - AVR NPO</title>
        <link rel="stylesheet" href="/assets/css/pico.min.css"/>
        <link rel="stylesheet" href="/assets/css/custom.css"/>
        <script src="/assets/js/htmx.min.js" defer></script>
    </head>
    <body>
        @Navigation()
        @contents
        @Footer()
    </body>
    </html>
}
```

### Reusable Components
```templ
templ ErrorAlert(message string) {
    <div role="alert" style="background-color: var(--pico-del-color);">
        { message }
    </div>
}

templ SuccessAlert(message string) {
    <div role="alert" style="background-color: var(--pico-ins-color);">
        { message }
    </div>
}

templ FormField(label, name, fieldType, value string, required bool) {
    <label for={ name }>{ label }</label>
    <input 
        type={ fieldType } 
        id={ name } 
        name={ name } 
        value={ value }
        if required {
            required
        }
    />
}
```

## Handler Patterns

### Simple Page Handler
```go
func handleContact(e *core.RequestEvent) error {
    csrfToken, _ := middleware.GetCSRFToken(e)
    component := templates.ContactPage(csrfToken)
    return component.Render(context.Background(), e.Response())
}
```

### Form Submission Handler
```go
func handleContactPost(e *core.RequestEvent) error {
    csrfToken, _ := middleware.GetCSRFToken(e)
    
    // Get form data
    name := e.Request.FormValue("name")
    email := e.Request.FormValue("email")
    message := e.Request.FormValue("message")
    
    // Validate
    if name == "" || email == "" || message == "" {
        component := templates.ContactForm(csrfToken)
        e.Response().WriteHeader(422)
        return component.Render(context.Background(), e.Response())
    }
    
    // Process form...
    
    // For HTMX requests, return partial
    if e.Request.Header.Get("HX-Request") == "true" {
        component := templates.SuccessAlert("Message sent successfully!")
        return component.Render(context.Background(), e.Response())
    }
    
    // For regular requests, redirect
    return e.Redirect(302, "/contact/success")
}
```

### HTMX Partial Response
```go
func handleSearchPosts(e *core.RequestEvent) error {
    query := e.Request.URL.Query().Get("q")
    
    // Search posts
    posts, err := searchBlogPosts(app, query)
    if err != nil {
        return err
    }
    
    // Return just the results component
    component := templates.PostSearchResults(posts)
    return component.Render(context.Background(), e.Response())
}
```

## HTMX Integration Patterns

### Form with Inline Validation
```templ
templ EmailField(csrfToken string) {
    <form hx-post="/validate/email" hx-trigger="blur from:find input" hx-target="#email-error">
        <input type="hidden" name="csrf_token" value={ csrfToken }/>
        <label for="email">Email</label>
        <input type="email" name="email" id="email" required/>
        <div id="email-error"></div>
    </form>
}
```

### Dynamic Content Loading
```templ
templ BlogPostList() {
    <div id="blog-posts">
        <button 
            hx-get="/blog/posts" 
            hx-target="#blog-posts"
            hx-swap="innerHTML">
            Load Posts
        </button>
    </div>
}
```

### Infinite Scroll
```templ
templ PostsList(posts []*models.Record, page int, hasMore bool) {
    for _, post := range posts {
        @PostCard(post)
    }
    
    if hasMore {
        <div 
            hx-get={ fmt.Sprintf("/blog/posts?page=%d", page+1) }
            hx-trigger="revealed"
            hx-swap="outerHTML">
            <p>Loading more...</p>
        </div>
    }
}
```

## JavaScript Enhancement Patterns

### Theme Toggler
```javascript
// public/assets/js/theme.js
document.addEventListener('DOMContentLoaded', () => {
    const themeToggle = document.getElementById('theme-toggle');
    const currentTheme = localStorage.getItem('theme') || 'light';
    
    document.documentElement.setAttribute('data-theme', currentTheme);
    
    themeToggle?.addEventListener('click', () => {
        const newTheme = document.documentElement.getAttribute('data-theme') === 'light' 
            ? 'dark' 
            : 'light';
        document.documentElement.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
    });
});
```

### Form Enhancement
```javascript
// Enhance forms with client-side validation
document.addEventListener('DOMContentLoaded', () => {
    const forms = document.querySelectorAll('form[data-validate]');
    
    forms.forEach(form => {
        form.addEventListener('submit', (e) => {
            // Client-side validation before submission
            const isValid = validateForm(form);
            if (!isValid) {
                e.preventDefault();
            }
        });
    });
});
```

## Benefits of This Architecture

1. **Type Safety**: Templ provides compile-time type checking
2. **Performance**: Templates compiled to Go code, no runtime parsing
3. **Component Reuse**: Small, focused components easy to reuse
4. **Progressive Enhancement**: Works without JavaScript, better with it
5. **HTMX Integration**: Smooth UX without heavy JavaScript frameworks
6. **SEO-Friendly**: Full HTML pages with proper navigation
7. **Developer Experience**: Go tooling, autocomplete, refactoring support

## Best Practices

### Templ Component Design
- **Keep components small and focused** - Single responsibility
- **Use parameters for data** - Avoid global state
- **Compose larger views** from smaller components
- **Type-safe props** - Leverage Go's type system

### HTMX Usage
- **Add HTMX progressively** - Start with working forms
- **Use appropriate triggers** - `click`, `submit`, `change`, `revealed`
- **Target specific elements** - Use `hx-target` for precise updates
- **Provide loading states** - Use `hx-indicator` for feedback

### CSRF Protection
- **Always include CSRF token** in forms
- **Use meta tag for HTMX** - `<meta name="csrf-token">`
- **Validate on server** - Never trust client-side only

### Accessibility
- **Semantic HTML** - Use proper elements
- **ARIA attributes** - When needed for dynamic content
- **Keyboard navigation** - All interactions accessible
- **Loading states** - Announce to screen readers

## Testing Templ Components

### Unit Tests
```go
func TestContactForm(t *testing.T) {
    csrfToken := "test-token"
    
    // Render to buffer
    buf := new(bytes.Buffer)
    err := templates.ContactForm(csrfToken).Render(context.Background(), buf)
    
    require.NoError(t, err)
    html := buf.String()
    
    // Assert content
    assert.Contains(t, html, "csrf_token")
    assert.Contains(t, html, csrfToken)
    assert.Contains(t, html, "name")
    assert.Contains(t, html, "email")
}
```

### Integration Tests
```go
func TestContactPageE2E(t *testing.T) {
    app, cleanup := setupTestApp(t)
    defer cleanup()
    
    req := httptest.NewRequest("GET", "/contact", nil)
    rec := httptest.NewRecorder()
    
    e := &core.RequestEvent{
        Request: req,
        Response: rec,
    }
    
    err := handleContact(e)
    require.NoError(t, err)
    
    assert.Equal(t, 200, rec.Code)
    assert.Contains(t, rec.Body.String(), "Contact Us")
}
```

This architecture provides a modern, type-safe, and performant approach to building server-rendered applications with progressive enhancement.
