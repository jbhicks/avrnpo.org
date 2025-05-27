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
      themeIcon.innerHTML = `<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25A2.25 2.25 0 015.25 3h13.5A2.25 2.25 0 0121 5.25z"></path>
      </svg>`;
    }
  } else {
    localStorage.setItem('picoPreferredColorScheme', theme);
    document.documentElement.setAttribute('data-theme', theme);
    if (themeIcon) {
      if (theme === 'dark') {
        themeIcon.innerHTML = `<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"></path>
        </svg>`;
      } else { // light
        themeIcon.innerHTML = `<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"></path>
        </svg>`;
      }
    }
  }
}

window.toggleTheme = function() { // Explicitly global
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
  const savedTheme = localStorage.getItem('picoPreferredColorScheme');
  const themeToggleButtons = document.querySelectorAll('.theme-toggle-btn');

  let initialTheme = savedTheme;
  if (!savedTheme) {
    initialTheme = 'dark'; // Default to dark
  }
  
  setTheme(initialTheme); // Call setTheme to initialize data-theme, localStorage (if 'dark' or 'light'), and icon.
  
  // Add event listener to all theme toggle buttons
  themeToggleButtons.forEach(button => {
    button.addEventListener('click', toggleTheme);
  });
});
