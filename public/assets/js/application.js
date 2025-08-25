// Application JavaScript will be added here

// Enhanced HTMX setup for better UX
document.addEventListener('DOMContentLoaded', function() {
    // Add CSRF token to all HTMX requests
    document.body.addEventListener('htmx:configRequest', function(event) {
        const csrfToken = document.querySelector('meta[name="csrf-token"]');
        if (csrfToken && csrfToken.content) {
            event.detail.headers['X-CSRF-Token'] = csrfToken.content;
        }
    });

    // Add loading indicators for HTMX requests
    document.body.addEventListener('htmx:beforeRequest', function(event) {
        // Add loading state to buttons/links that trigger requests
        if (event.detail.elt.tagName === 'BUTTON' || event.detail.elt.tagName === 'INPUT') {
            event.detail.elt.style.opacity = '0.7';
            event.detail.elt.disabled = true;
        }
    });
    
    document.body.addEventListener('htmx:afterRequest', function(event) {
        // Remove loading state
        if (event.detail.elt.tagName === 'BUTTON' || event.detail.elt.tagName === 'INPUT') {
            event.detail.elt.style.opacity = '1';
            event.detail.elt.disabled = false;
        }
    });
});
