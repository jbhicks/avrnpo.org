// Theme switching functionality for Pico CSS
(function() {
    'use strict';
    
    // Get theme preference from localStorage or default to 'auto'
    const getStoredTheme = () => localStorage.getItem('picoPreferredColorScheme') || 'auto';
    const setStoredTheme = (theme) => localStorage.setItem('picoPreferredColorScheme', theme);
    
    // Apply theme to document
    const applyTheme = (theme) => {
        if (theme === 'auto') {
            document.documentElement.removeAttribute('data-theme');
        } else {
            document.documentElement.setAttribute('data-theme', theme);
        }
    };
    
    // Initialize theme on page load
    const initTheme = () => {
        const storedTheme = getStoredTheme();
        applyTheme(storedTheme);
        
        // Update theme switcher UI if it exists
        const themeSwitcher = document.querySelector('[data-theme-switcher]');
        if (themeSwitcher) {
            themeSwitcher.value = storedTheme;
        }
    };
    
    // Handle theme change
    const changeTheme = (theme) => {
        setStoredTheme(theme);
        applyTheme(theme);
    };
    
    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initTheme);
    } else {
        initTheme();
    }
    
    // Listen for theme switcher changes
    document.addEventListener('change', (e) => {
        if (e.target.matches('[data-theme-switcher]')) {
            changeTheme(e.target.value);
        }
    });
    
    // Export for global access
    window.ThemeManager = {
        getTheme: getStoredTheme,
        setTheme: changeTheme,
        applyTheme: applyTheme
    };
})();
