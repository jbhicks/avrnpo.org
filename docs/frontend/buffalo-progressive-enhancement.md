# Buffalo + Progressive Enhancement Architecture

This document outlines our simplified frontend architecture using standard Buffalo patterns with Plush templates and progressive JavaScript enhancement.

## Architecture Overview

We use a **single-template architecture** where:
- Each route renders one complete HTML template
- No AJAX/HTMX header detection in handlers
- Progressive JavaScript enhancement for improved UX
- Standard HTML forms with fallback behavior

## Core Principles

### 1. HTML-First Approach
All functionality must work with basic HTML and forms:

```html
<!-- ✅ Works without JavaScript -->
<form method="POST" action="/contact">
    <%= csrf() %>
    <input type="text" name="name" required>
    <input type="email" name="email" required>
    <textarea name="message" required></textarea>
    <button type="submit">Send Message</button>
</form>
```

### 2. Single Template Per Route
Each handler renders one complete template:

```go
// ✅ Simple, predictable handler
func ContactHandler(c buffalo.Context) error {
    return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
}

func ContactCreateHandler(c buffalo.Context) error {
    // Process form...
    if errors.HasAny() {
        c.Set("errors", errors)
        return c.Render(http.StatusUnprocessableEntity, r.HTML("pages/contact.plush.html"))
    }
    return c.Redirect(http.StatusSeeOther, "/contact/success")
}
```

### 3. Progressive JavaScript Enhancement
Add JavaScript to improve UX while maintaining fallback:

```javascript
// Enhance forms without breaking basic functionality
document.addEventListener('DOMContentLoaded', () => {
    const forms = document.querySelectorAll('[data-enhance="ajax"]');
    
    forms.forEach(form => {
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            try {
                const response = await fetch(form.action, {
                    method: form.method,
                    body: new FormData(form),
                    headers: {
                        'X-Requested-With': 'XMLHttpRequest'
                    }
                });
                
                if (response.ok) {
                    // Handle success (show message, redirect, etc.)
                    showSuccessMessage('Form submitted successfully!');
                } else {
                    // Handle validation errors
                    const html = await response.text();
                    updateFormErrors(form, html);
                }
            } catch (error) {
                // Fallback to normal form submission
                form.submit();
            }
        });
    });
});
```

## Template Patterns

### Standard Page Template
```html
<!-- pages/contact.plush.html -->
<main>
    <h1>Contact Us</h1>
    
    <%= if (errors) { %>
        <div class="alert alert-danger">
            <%= for (error) in errors { %>
                <p><%= error %></p>
            <% } %>
        </div>
    <% } %>
    
    <form method="POST" action="/contact" data-enhance="ajax">
        <%= csrf() %>
        
        <label for="name">Name</label>
        <input type="text" name="name" id="name" value="<%= name %>" required>
        
        <label for="email">Email</label>
        <input type="email" name="email" id="email" value="<%= email %>" required>
        
        <label for="message">Message</label>
        <textarea name="message" id="message" required><%= message %></textarea>
        
        <button type="submit">Send Message</button>
    </form>
</main>
```

### Reusable Partials
Keep partials for truly reusable components:

```html
<!-- _flash.plush.html -->
<%= if (flash["success"]) { %>
    <div class="alert alert-success" role="alert">
        <%= flash["success"] %>
    </div>
<% } %>

<%= if (flash["error"]) { %>
    <div class="alert alert-danger" role="alert">
        <%= flash["error"] %>
    </div>
<% } %>
```

## Handler Patterns

### Form Handlers
```go
func DonateHandler(c buffalo.Context) error {
    // GET: Show donation form
    return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
}

func DonateCreateHandler(c buffalo.Context) error {
    // POST: Process donation
    donation := &models.Donation{}
    
    if err := c.Bind(donation); err != nil {
        return err
    }
    
    verrs, err := donation.Validate()
    if err != nil {
        return err
    }
    
    if verrs.HasAny() {
        c.Set("errors", verrs)
        c.Set("donation", donation)
        return c.Render(http.StatusUnprocessableEntity, r.HTML("pages/donate.plush.html"))
    }
    
    // Process successful donation...
    c.Flash().Add("success", "Thank you for your donation!")
    return c.Redirect(http.StatusSeeOther, "/donate/success")
}
```

### API-Style Responses for AJAX
For enhanced JavaScript interactions, detect AJAX and return JSON:

```go
func DonateCreateHandler(c buffalo.Context) error {
    // ... validation logic ...
    
    if verrs.HasAny() {
        // For AJAX requests, return JSON
        if c.Request().Header.Get("X-Requested-With") == "XMLHttpRequest" {
            return c.Render(http.StatusUnprocessableEntity, r.JSON(map[string]interface{}{
                "errors": verrs.Errors,
                "success": false,
            }))
        }
        
        // For regular requests, render full page
        c.Set("errors", verrs)
        return c.Render(http.StatusUnprocessableEntity, r.HTML("pages/donate.plush.html"))
    }
    
    // Success handling...
    if c.Request().Header.Get("X-Requested-With") == "XMLHttpRequest" {
        return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
            "success": true,
            "message": "Donation processed successfully!",
            "redirect": "/donate/success",
        }))
    }
    
    return c.Redirect(http.StatusSeeOther, "/donate/success")
}
```

## JavaScript Enhancement Patterns

### Form Validation
```javascript
// Client-side validation enhancement
function enhanceFormValidation() {
    const forms = document.querySelectorAll('[data-validate="true"]');
    
    forms.forEach(form => {
        const inputs = form.querySelectorAll('input[required], textarea[required]');
        
        inputs.forEach(input => {
            input.addEventListener('blur', validateField);
            input.addEventListener('input', clearFieldError);
        });
    });
}

function validateField(e) {
    const field = e.target;
    const value = field.value.trim();
    
    // Clear previous errors
    clearFieldError({ target: field });
    
    if (field.hasAttribute('required') && !value) {
        showFieldError(field, 'This field is required');
        return false;
    }
    
    if (field.type === 'email' && value && !isValidEmail(value)) {
        showFieldError(field, 'Please enter a valid email address');
        return false;
    }
    
    return true;
}
```

### Dynamic Content Updates
```javascript
// Update page sections without full reload
async function updateContent(url, targetSelector) {
    try {
        const response = await fetch(url, {
            headers: {
                'X-Requested-With': 'XMLHttpRequest'
            }
        });
        
        if (response.ok) {
            const html = await response.text();
            const target = document.querySelector(targetSelector);
            
            if (target) {
                target.innerHTML = html;
                // Re-enhance any new content
                enhanceNewContent(target);
            }
        }
    } catch (error) {
        // Fallback to full page navigation
        window.location.href = url;
    }
}
```

## Benefits of This Architecture

1. **Simplicity**: Standard Buffalo patterns, easy to understand
2. **Reliability**: Works without JavaScript, degrades gracefully
3. **Performance**: No heavy frontend framework, fast initial loads
4. **SEO-Friendly**: Full HTML pages, proper navigation
5. **Maintainable**: Single template per view, clear separation of concerns
6. **Accessible**: Standard HTML forms and navigation

## Migration from HTMX

When migrating from HTMX:

1. **Remove HTMX attributes** from templates
2. **Consolidate templates** - remove duplicate partials
3. **Simplify handlers** - remove header detection logic
4. **Add progressive enhancement** - implement JavaScript features as needed
5. **Update tests** - expect full page responses

This approach gives you the benefits of modern UX while maintaining the simplicity and reliability of traditional web applications.