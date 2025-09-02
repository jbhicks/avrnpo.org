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

## Backend Best Practices

Our application uses `hx-boost` for navigation, which simplifies our backend logic considerably.

### Response Patterns

#### Full Page Responses
For most requests, including navigation and form submissions, handlers should return a full HTML page. `hx-boost` will automatically extract the `<body>` content to update the page. This is the standard and recommended approach.

```go
// Standard handler for a boosted link or form
func MyPageHandler(c buffalo.Context) error {
    // ... logic to fetch data ...
    return c.Render(http.StatusOK, r.HTML("pages/my_page.plush.html"))
}
```

#### Partial Page Responses (for specific components)
In some cases, you may want to update only a small part of a page (e.g., a search results container, a dynamic chart). In these situations, you can use explicit `hx-` attributes (`hx-get`, `hx-post`, etc.) and return an HTML fragment.

```html
<!-- Requesting a partial update -->
<div hx-get="/search-results?query=htmx" hx-target="#results" hx-swap="innerHTML">
    Search
</div>
<div id="results"></div>
```

The handler for this would return just the HTML for the results:

```go
// Handler for a partial update
func SearchResultsHandler(c buffalo.Context) error {
    // ... logic to fetch search results ...
    return c.Render(http.StatusOK, r.HTML("partials/_search_results.plush.html"))
}
```

#### Error Handling
Return appropriate HTTP status codes:
- 200: Success with content
- 204: Success with no content (e.g., for a `hx-swap="delete"`)
- 400: Client error (e.g., validation failure)
- 500: Server error

When a form submission fails validation, re-render the full page containing the form, including the validation errors. `hx-boost` will handle swapping the body and showing the errors to the user.

## Common Patterns

### Progressive Enhancement
Always include proper `href` and `action` attributes as fallbacks. This is the foundation of our `hx-boost` strategy.

```html
<!-- Works with and without JavaScript -->
<a href="/donate">Donate Now</a>

<form action="/contact" method="post">
    <!-- Form fields -->
</form>
```
`hx-boost` will automatically enhance these standard HTML elements.

### Explicit HTMX for Components
For interactions that don't involve a full page navigation, use explicit `hx-` attributes.

```html
<!-- Good for components like modals, inline editing, etc. -->
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
Listen for HTMX events in JavaScript to re-initialize components after a swap.

```javascript
// After content swap
document.addEventListener('htmx:afterSwap', function(event) {
    console.log('Content swapped:', event.detail);
    // Re-initialize JavaScript components like charts, maps, etc.
});

// After settle (animations complete)
document.addEventListener('htmx:afterSettle', function(event) {
    console.log('Content settled:', event.detail);
});
```

## Debugging

### HTMX Headers
HTMX sends these headers with requests:
- `HX-Request: true` - Indicates HTMX request
- `HX-Current-URL` - Current page URL
- `HX-Target` - ID of target element (if specified)
- `HX-Trigger` - ID of triggered element (if specified)

### Browser DevTools
- The Network tab shows HTMX requests.
- Look for the `HX-Request: true` header.
- The response should be a full HTML document for boosted links/forms, or an HTML fragment for partial swaps.

### Logging
Enable HTMX logging in your main JavaScript file for detailed console output:

```javascript
htmx.logAll();
```

## Security Considerations

### CSRF Protection
Buffalo's `formFor` helper automatically includes a CSRF token. For manual forms or AJAX requests with `hx-post`, ensure the token is included. `hx-boost` automatically handles this for forms.

### Validation
Always validate data on the server side. HTMX is just a transport mechanism.

## Performance Tips

### Minimize Response Size for Partials
When returning partials, return only the necessary HTML.

### Lazy Loading
Load content only when it's needed or becomes visible.

```html
<div hx-get="/api/comments" hx-trigger="intersect">
    Loading comments...
</div>
```

## Common Issues and Solutions

### JavaScript Not Working After Swap
This is a common issue. Use an `htmx:afterSwap` event listener to re-initialize any JavaScript components within the newly loaded content.

This reference covers the essential HTMX concepts needed for the AVR NPO project. For complete documentation, visit https://htmx.org/docs/.
