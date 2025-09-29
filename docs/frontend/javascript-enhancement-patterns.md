# JavaScript Progressive Enhancement Patterns

This guide provides specific patterns for progressively enhancing Buffalo + Plush applications with JavaScript.

## Core Philosophy

1. **HTML First**: All functionality works without JavaScript
2. **Layer Enhancement**: JavaScript adds convenience, not core functionality  
3. **Graceful Degradation**: Failures fall back to standard behavior
4. **Performance Focused**: Minimal JavaScript, maximum impact

## Essential Patterns

### 1. Form Enhancement

#### Basic Pattern
```javascript
// Enhance forms to submit via AJAX while maintaining fallback
document.addEventListener('DOMContentLoaded', () => {
    enhanceForms();
});

function enhanceForms() {
    const forms = document.querySelectorAll('[data-enhance="ajax"]');
    
    forms.forEach(form => {
        form.addEventListener('submit', handleEnhancedSubmit);
    });
}

async function handleEnhancedSubmit(e) {
    e.preventDefault();
    const form = e.target;
    const submitButton = form.querySelector('[type="submit"]');
    
    // Show loading state
    const originalText = submitButton.textContent;
    submitButton.disabled = true;
    submitButton.textContent = 'Submitting...';
    
    try {
        const response = await fetch(form.action, {
            method: form.method,
            body: new FormData(form),
            headers: {
                'X-Requested-With': 'XMLHttpRequest'
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            handleFormSuccess(form, data);
        } else {
            const errorData = await response.json();
            handleFormErrors(form, errorData.errors);
        }
    } catch (error) {
        console.error('Enhanced form submission failed:', error);
        // Fallback to normal submission
        form.submit();
    } finally {
        // Restore button state
        submitButton.disabled = false;
        submitButton.textContent = originalText;
    }
}
```

#### Success/Error Handling
```javascript
function handleFormSuccess(form, data) {
    // Clear any existing errors
    clearFormErrors(form);
    
    // Show success message
    if (data.message) {
        showNotification(data.message, 'success');
    }
    
    // Handle redirect or reset
    if (data.redirect) {
        window.location.href = data.redirect;
    } else {
        form.reset();
    }
}

function handleFormErrors(form, errors) {
    clearFormErrors(form);
    
    Object.entries(errors).forEach(([field, messages]) => {
        const input = form.querySelector(`[name="${field}"]`);
        if (input) {
            showFieldError(input, messages[0]);
        }
    });
}

function showFieldError(input, message) {
    // Remove existing error
    const existingError = input.parentNode.querySelector('.field-error');
    if (existingError) {
        existingError.remove();
    }
    
    // Add error class
    input.classList.add('error');
    
    // Create error message
    const errorEl = document.createElement('div');
    errorEl.className = 'field-error';
    errorEl.textContent = message;
    
    // Insert after input
    input.parentNode.insertBefore(errorEl, input.nextSibling);
}
```

### 2. Dynamic Content Loading

#### Content Sections
```javascript
// Load content sections without full page reload
function enhanceNavigation() {
    const links = document.querySelectorAll('[data-load-content]');
    
    links.forEach(link => {
        link.addEventListener('click', async (e) => {
            e.preventDefault();
            
            const url = link.href;
            const target = link.dataset.loadContent;
            
            await loadContent(url, target);
            
            // Update URL without page reload
            history.pushState({ url }, '', url);
        });
    });
}

async function loadContent(url, targetSelector) {
    const target = document.querySelector(targetSelector);
    if (!target) return;
    
    // Show loading state
    target.classList.add('loading');
    
    try {
        const response = await fetch(url, {
            headers: {
                'X-Requested-With': 'XMLHttpRequest'
            }
        });
        
        if (response.ok) {
            const html = await response.text();
            target.innerHTML = html;
            
            // Re-enhance new content
            enhanceNewContent(target);
        }
    } catch (error) {
        console.error('Content loading failed:', error);
        // Fallback to navigation
        window.location.href = url;
    } finally {
        target.classList.remove('loading');
    }
}
```

### 3. Search and Filtering

#### Live Search
```javascript
function enhanceSearch() {
    const searchInputs = document.querySelectorAll('[data-live-search]');
    
    searchInputs.forEach(input => {
        let timeoutId;
        
        input.addEventListener('input', (e) => {
            clearTimeout(timeoutId);
            
            timeoutId = setTimeout(() => {
                performSearch(input);
            }, 300); // Debounce 300ms
        });
    });
}

async function performSearch(input) {
    const query = input.value.trim();
    const form = input.closest('form');
    const resultsContainer = document.querySelector(input.dataset.liveSearch);
    
    if (!resultsContainer) return;
    
    try {
        const formData = new FormData(form);
        const params = new URLSearchParams(formData);
        
        const response = await fetch(`${form.action}?${params}`, {
            headers: {
                'X-Requested-With': 'XMLHttpRequest'
            }
        });
        
        if (response.ok) {
            const html = await response.text();
            resultsContainer.innerHTML = html;
            enhanceNewContent(resultsContainer);
        }
    } catch (error) {
        console.error('Search failed:', error);
    }
}
```

### 4. UI Enhancement Utilities

#### Notifications
```javascript
function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;
    
    // Add close button
    const closeBtn = document.createElement('button');
    closeBtn.className = 'notification-close';
    closeBtn.innerHTML = '×';
    closeBtn.addEventListener('click', () => {
        notification.remove();
    });
    
    notification.appendChild(closeBtn);
    
    // Add to page
    document.body.appendChild(notification);
    
    // Auto-remove after delay
    setTimeout(() => {
        if (notification.parentNode) {
            notification.remove();
        }
    }, 5000);
}
```

#### Modal Dialogs
```javascript
function enhanceModals() {
    const modalTriggers = document.querySelectorAll('[data-modal]');
    
    modalTriggers.forEach(trigger => {
        trigger.addEventListener('click', (e) => {
            e.preventDefault();
            
            const modalId = trigger.dataset.modal;
            const modal = document.getElementById(modalId);
            
            if (modal) {
                showModal(modal);
            }
        });
    });
}

function showModal(modal) {
    modal.style.display = 'block';
    document.body.classList.add('modal-open');
    
    // Close on backdrop click
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            hideModal(modal);
        }
    });
    
    // Close on escape key
    const escapeHandler = (e) => {
        if (e.key === 'Escape') {
            hideModal(modal);
            document.removeEventListener('keydown', escapeHandler);
        }
    };
    
    document.addEventListener('keydown', escapeHandler);
}

function hideModal(modal) {
    modal.style.display = 'none';
    document.body.classList.remove('modal-open');
}
```

### 5. Utility Functions

#### Content Re-enhancement
```javascript
function enhanceNewContent(container) {
    // Re-apply all enhancements to new content
    enhanceForms();
    enhanceNavigation();
    enhanceSearch();
    enhanceModals();
    
    // Dispatch custom event for other scripts
    container.dispatchEvent(new CustomEvent('contentEnhanced', {
        bubbles: true
    }));
}
```

#### Error Handling
```javascript
function clearFormErrors(form) {
    // Remove error classes
    form.querySelectorAll('.error').forEach(el => {
        el.classList.remove('error');
    });
    
    // Remove error messages
    form.querySelectorAll('.field-error').forEach(el => {
        el.remove();
    });
}

function clearFieldError(e) {
    const field = e.target;
    field.classList.remove('error');
    
    const errorEl = field.parentNode.querySelector('.field-error');
    if (errorEl) {
        errorEl.remove();
    }
}
```

## Implementation Strategy

### 1. Initialize Enhancements
```javascript
document.addEventListener('DOMContentLoaded', () => {
    enhanceForms();
    enhanceNavigation();
    enhanceSearch();
    enhanceModals();
    
    console.log('Progressive enhancements applied');
});
```

### 2. Handle Browser History
```javascript
window.addEventListener('popstate', (e) => {
    if (e.state && e.state.url) {
        // Handle back/forward navigation for AJAX-loaded content
        const mainContent = document.querySelector('#main-content');
        if (mainContent) {
            loadContent(e.state.url, '#main-content');
        }
    }
});
```

### 3. Performance Considerations
```javascript
// Use event delegation for better performance
document.addEventListener('click', (e) => {
    // Handle enhanced navigation
    if (e.target.matches('[data-load-content]')) {
        e.preventDefault();
        const url = e.target.href;
        const target = e.target.dataset.loadContent;
        loadContent(url, target);
        history.pushState({ url }, '', url);
    }
    
    // Handle modal triggers
    if (e.target.matches('[data-modal]')) {
        e.preventDefault();
        const modalId = e.target.dataset.modal;
        const modal = document.getElementById(modalId);
        if (modal) showModal(modal);
    }
});
```

## Template Integration

### HTML Data Attributes
```html
<!-- Enhanced form -->
<form method="POST" action="/contact" data-enhance="ajax">
    <%= csrf() %>
    <!-- form fields -->
    <button type="submit">Send Message</button>
</form>

<!-- Enhanced navigation -->
<a href="/admin/users" data-load-content="#main-content">Users</a>

<!-- Live search -->
<input type="text" name="search" data-live-search="#search-results">

<!-- Modal trigger -->
<button data-modal="confirm-delete">Delete</button>
```

### CSS Support
```css
/* Loading states */
.loading {
    opacity: 0.6;
    pointer-events: none;
}

/* Form errors */
.error {
    border-color: var(--pico-del-color);
}

.field-error {
    color: var(--pico-del-color);
    font-size: 0.875rem;
    margin-top: 0.25rem;
}

/* Notifications */
.notification {
    position: fixed;
    top: 1rem;
    right: 1rem;
    padding: 1rem;
    border-radius: var(--pico-border-radius);
    z-index: 1000;
}

.notification-success {
    background: var(--pico-ins-color);
    color: white;
}

.notification-error {
    background: var(--pico-del-color);
    color: white;
}
```

This approach provides rich interactivity while maintaining the reliability and simplicity of standard HTML forms and navigation.