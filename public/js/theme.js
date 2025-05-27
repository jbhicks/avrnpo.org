/**
 * Theme switching functionality for Pico.css
 * Handles dark/light/auto theme switching with localStorage persistence
 */

function setTheme(theme) {
  const themeIcon = document.getElementById('theme-icon');
  
  if (theme === 'auto') {
    localStorage.removeItem('picoPreferredColorScheme');
    document.documentElement.removeAttribute('data-theme');
    if (themeIcon) {
      themeIcon.textContent = '🔄';
    }
  } else {
    localStorage.setItem('picoPreferredColorScheme', theme);
    document.documentElement.setAttribute('data-theme', theme);
    if (themeIcon) {
      themeIcon.textContent = theme === 'dark' ? '🌙' : '☀️';
    }
  }
}

// Initialize theme on page load
document.addEventListener('DOMContentLoaded', function() {
  const themeIcon = document.getElementById('theme-icon');
  const savedTheme = localStorage.getItem('picoPreferredColorScheme');
  
  if (savedTheme) {
    document.documentElement.setAttribute('data-theme', savedTheme);
    if (themeIcon) {
      themeIcon.textContent = savedTheme === 'dark' ? '🌙' : '☀️';
    }
  } else {
    // Default to dark mode if no preference is set
    document.documentElement.setAttribute('data-theme', 'dark');
    if (themeIcon) {
      themeIcon.textContent = '🌙';
    }
  }
});
