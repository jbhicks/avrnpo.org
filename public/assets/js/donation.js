// Donation form functionality for AVR website
// Minimal JavaScript - HTMX handles most interactions declaratively
// Form validation handled server-side with Buffalo flash messages

(function() {
    'use strict';

    // Initialize when DOM is ready
    function initialize() {
        // HTMX handles form submission and validation
        // Buffalo flash messages handle error display
        // CSRF tokens are handled by application.js
        // No client-side validation needed - server handles everything
        console.log('Donation form initialized - using server-side validation with Buffalo flash messages');
    }

    // Wait for DOM to be ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initialize);
    } else {
        initialize();
    }

    // Re-initialize after HTMX loads new content
    if (typeof htmx !== 'undefined') {
        htmx.onLoad(function(content) {
            // Check if the loaded content contains a donation form
            if (content.querySelector && content.querySelector('#donation-form')) {
                initialize();
            }
        });
    }
})();
