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
      if (window.getIcon) {
        themeIcon.innerHTML = window.getIcon('computerDesktop', 'w-4 h-4');
      } else {
        themeIcon.textContent = '🖥️';
      }
    }
  } else {
    localStorage.setItem('picoPreferredColorScheme', theme);
    document.documentElement.setAttribute('data-theme', theme);
    if (themeIcon) {
      if (window.getIcon) {
        themeIcon.innerHTML = theme === 'dark' ? window.getIcon('moon', 'w-4 h-4') : window.getIcon('sun', 'w-4 h-4');
      } else {
        themeIcon.textContent = theme === 'dark' ? '🌙' : '☀️';
      }
    }
  }
}

function toggleTheme() {
  const currentTheme = localStorage.getItem('picoPreferredColorScheme');
  
  if (!currentTheme || currentTheme === 'auto') {
    setTheme('light');
  } else if (currentTheme === 'light') {
    setTheme('dark');
  } else {
    setTheme('light');
  }
}

// Initialize theme on page load
document.addEventListener('DOMContentLoaded', function() {
  const themeIcon = document.getElementById('theme-icon');
  const savedTheme = localStorage.getItem('picoPreferredColorScheme');
  
  if (savedTheme) {
    document.documentElement.setAttribute('data-theme', savedTheme);
    if (themeIcon) {
      if (window.getIcon) {
        if (savedTheme === 'dark') {
          themeIcon.innerHTML = window.getIcon('moon', 'w-4 h-4');
        } else if (savedTheme === 'light') {
          themeIcon.innerHTML = window.getIcon('sun', 'w-4 h-4');
        } else {
          themeIcon.innerHTML = window.getIcon('computerDesktop', 'w-4 h-4');
        }
      } else {
        // Fallback to emoji if getIcon is not available
        if (savedTheme === 'dark') {
          themeIcon.textContent = '🌙';
        } else if (savedTheme === 'light') {
          themeIcon.textContent = '☀️';
        } else {
          themeIcon.textContent = '🖥️';
        }
      }
    }
  } else {
    // Default to dark mode if no preference is set
    document.documentElement.setAttribute('data-theme', 'dark');
    if (themeIcon) {
      if (window.getIcon) {
        themeIcon.innerHTML = window.getIcon('moon', 'w-4 h-4');
      } else {
        themeIcon.textContent = '🌙';
      }
    }
  }
});
