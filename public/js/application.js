/**
 * Application JavaScript Entry Point
 * Main JavaScript file for AVR NPO website
 */

// Initialize application when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    console.log('AVR NPO Application initialized');
    
    // Initialize any global functionality here
    initializeGlobalFeatures();
});

function initializeGlobalFeatures() {
    // Initialize HTMX if needed
    if (typeof htmx !== 'undefined') {
        console.log('HTMX loaded and ready');
    }
    
    // Initialize theme switching
    if (typeof initializeTheme === 'function') {
        initializeTheme();
    }
    
    // Initialize donation system if on donate page
    if (window.location.pathname === '/donate' && typeof DonationSystem !== 'undefined') {
        new DonationSystem();
    }
}
