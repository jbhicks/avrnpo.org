# Buffalo Plush Template Syntax

Based on the official Plush documentation: https://github.com/gobuffalo/plush

## 🚨 CRITICAL: Buffalo Partial Naming Convention

**⚠️ COMMON GOTCHA - This causes recurring 500 errors:**

Buffalo automatically adds an underscore prefix to partial names. When calling partials:

**❌ WRONG:**
```html
<%= partial("pages/_donate_content.plush.html") %>
```

**✅ CORRECT:**
```html
<%= partial("pages/donate_content.plush.html") %>
```

**Why this happens:**
- Buffalo looks for `_donate_content.plush.html` when you call `partial("donate_content.plush.html")`
- If you add the underscore yourself, Buffalo looks for `__donate_content.plush.html` (double underscore)
- This results in "could not find template" errors

**Rule: Never include the underscore in partial() calls - Buffalo adds it automatically**

## String Manipulation

Plush provides built-in helpers for string operations:

- `capitalize(string)` - capitalizes the first letter
- `len(string)` - gets the length of a string
- String slicing should use helper functions, not Go syntax

## Template Syntax

- Use `<%= %>` for output
- Use `<% %>` for execution without output
- Conditionals: `<%= if (condition) { %> ... <% } %>`
- String comparisons: `string != ""`

## Common Patterns

For getting the first character of a string, use helper functions or create custom helpers.
Avoid Go-style syntax like `string[0:1]` as Plush doesn't support this directly.

## Buffalo Context Variables

- `current_user` - contains the authenticated user object
- Standard string operations should use Plush helpers