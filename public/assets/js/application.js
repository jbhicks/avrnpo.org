// Application JavaScript for HTMX integration with CSRF support

document.addEventListener('DOMContentLoaded', function() {
    // Ensure HTMX requests include CSRF tokens
    document.body.addEventListener('htmx:configRequest', function(event) {
        // Find the closest form that contains a CSRF token
        const element = event.detail.elt;
        let form = element.closest('form');

        if (!form) {
            // If no form found, look for any form on the page with CSRF token
            const csrfInput = document.querySelector('input[name="authenticity_token"]');
            if (csrfInput) {
                form = csrfInput.closest('form');
            }
        }

        if (form) {
            const csrfInput = form.querySelector('input[name="authenticity_token"]');
            if (csrfInput && csrfInput.value) {
                // Add CSRF token to HTMX request parameters
                if (!event.detail.parameters) {
                    event.detail.parameters = {};
                }
                event.detail.parameters['authenticity_token'] = csrfInput.value;
            }
        }
    });

    // Add loading indicators for better UX
    document.body.addEventListener('htmx:beforeRequest', function(event) {
        const elt = event.detail.elt;
        if (elt.tagName === 'BUTTON' || elt.tagName === 'INPUT') {
            elt.style.opacity = '0.7';
            elt.disabled = true;
        }
    });

    document.body.addEventListener('htmx:afterRequest', function(event) {
        const elt = event.detail.elt;
        if (elt.tagName === 'BUTTON' || elt.tagName === 'INPUT') {
            elt.style.opacity = '1';
            elt.disabled = false;
        }
    });

    // Handle CSRF errors gracefully
    document.body.addEventListener('htmx:responseError', function(event) {
        if (event.detail.xhr.status === 403) {
            console.warn('Request blocked by CSRF protection. This may indicate a session issue.');
            // Try to refresh the page to get new tokens
            if (confirm('Your session may have expired. Refresh the page to continue?')) {
                window.location.reload();
            }
        }
    });
});

    // Ensure all forms have CSRF tokens (fallback for non-HTMX forms)
    document.addEventListener('submit', function(event) {
        const form = event.target;
        if (form.tagName === 'FORM') {
            window.CSRFUtils.addTokenToForm(form);
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

    // CSRF Error Handling
    document.body.addEventListener('htmx:responseError', function(event) {
        if (event.detail.xhr.status === 403) {
            // CSRF token error - try to refresh the page to get a new token
            console.warn('CSRF token validation failed. This may be due to session expiry.');
            // Optionally show user-friendly error message
            if (confirm('Your session may have expired. Would you like to refresh the page?')) {
                window.location.reload();
            }
        }
    });
});
