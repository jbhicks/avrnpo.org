// Application JavaScript for HTMX integration with CSRF support

document.addEventListener('DOMContentLoaded', function () {
  // Ensure HTMX requests include CSRF tokens
  // Get CSRF token from meta tag for HTMX requests
  document.body.addEventListener('htmx:configRequest', function(event) {
    const csrfToken = document.querySelector('meta[name="csrf-token"]');
    if (csrfToken) {
      event.detail.parameters['authenticity_token'] = csrfToken.getAttribute('content');
      console.log('HTMX: Added CSRF token to request:', csrfToken.getAttribute('content').substring(0, 20) + '...');
    } else {
      console.warn('HTMX: No CSRF token found in meta tag');
    }
    console.log('HTMX: Request parameters:', event.detail.parameters);
  });

  // Add loading indicators for better UX
  document.body.addEventListener('htmx:beforeRequest', function (event) {
    const elt = event.detail.elt;
    if (elt.tagName === 'BUTTON' || elt.tagName === 'INPUT') {
      elt.style.opacity = '0.7';
      elt.disabled = true;
    }
  });

  document.body.addEventListener('htmx:afterRequest', function (event) {
    const elt = event.detail.elt;
    if (elt.tagName === 'BUTTON' || elt.tagName === 'INPUT') {
      elt.style.opacity = '1';
      elt.disabled = false;
    }
  });

  // Handle HTMX errors gracefully
  document.body.addEventListener('htmx:responseError', function (event) {
    const xhr = event.detail.xhr;
    const status = xhr.status;

    // Handle different error types
    switch (status) {
      case 403:
        console.warn('Request blocked by CSRF protection. This may indicate a session issue.');
        if (confirm('Your session may have expired. Refresh the page to continue?')) {
          window.location.reload();
        }
        break;
      case 404:
        console.error('Resource not found:', xhr.responseURL);
        // Could show a user-friendly message
        break;
      case 422:
        console.warn('Validation error - check form fields');
        // Validation errors are typically handled by server-side rendering
        break;
      case 500:
        console.error('Server error occurred');
        if (confirm('A server error occurred. Would you like to refresh the page?')) {
          window.location.reload();
        }
        break;
      default:
        console.error('HTMX request failed with status:', status, xhr.responseText);
    }
  });

  // Handle network errors
  document.body.addEventListener('htmx:sendError', function (event) {
    console.error('Network error during HTMX request:', event.detail.error);
    // Could show offline/network error message
  });
});


