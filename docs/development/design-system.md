# AVR Design System Documentation

**Last Updated**: October 11, 2025  
**Status**: Complete - Pico CSS Refactoring Phase 1 & 2

## Overview

The American Veterans Rebuilding (AVR) website uses a semantic-first, component-based CSS architecture built on top of [Pico CSS](https://picocss.com/). This design system prioritizes maintainability, accessibility, and adherence to the Pico CSS philosophy while implementing a military/tactical aesthetic.

## Design Philosophy

### Core Principles

1. **Semantic-First**: Use meaningful class names that describe *what* something is, not *how* it looks
2. **Component-Based**: Group related styles into cohesive components
3. **Pico CSS Native**: Leverage Pico's variables and conventions, never fight against them
4. **Zero Utility Classes**: No Tailwind-style utilities; all styling is semantic
5. **Accessible by Default**: WCAG AA minimum, AAA preferred

### Zero Inline Styles

**Rule**: Never use `style=""` attributes in templates.  
**Reason**: Inline styles are unmaintainable, defeat caching, and break the separation of concerns.

**Example**:
```html
<!-- ❌ BAD -->
<div style="margin-bottom: 2rem;">

<!-- ✅ GOOD -->
<div class="section-spacing">
```

## Color System

### 3-Tier Visual Hierarchy

The AVR design uses a three-tier color system to create depth and hierarchy:

**Light Mode:**
- **Tier 1 (Body)**: `#e8e9ea` - Concrete gray background
- **Tier 2 (Cards)**: `#ffffff` - White cards stand out from background
- **Tier 3 (Sections)**: `#f5f5f5` - Light gray sections within cards

**Dark Mode:**
- **Tier 1 (Body)**: `#0d1117` - Deep dark background
- **Tier 2 (Cards)**: Pico's medium dark (via `--pico-card-background-color`)
- **Tier 3 (Sections)**: Darker than cards (via `--pico-card-sectional-background-color`)

### Color Palette

#### Primary Colors
- **Army Gold**: `#ffb627` - Primary actions, CTAs, highlights
- **Army Green**: `#4a5d23` - Secondary actions, accents
- **Burnt Orange**: `#d2691e` - Contrast actions, warnings

#### Implementation
```css
/* In custom.css - Pico variable overrides */
[data-theme="light"] {
    --pico-primary: #ffb627;
    --pico-secondary: #4a5d23;
    --pico-contrast: #d2691e;
    
    /* 3-tier system */
    --pico-background-color: #e8e9ea;
    --pico-card-background-color: #ffffff;
    --pico-card-sectional-background-color: #f5f5f5;
    --pico-card-border-color: rgba(74, 93, 35, 0.1);
    --pico-card-box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08), 
                             0 1px 4px rgba(0, 0, 0, 0.04);
}
```

### Card Styling

All cards automatically get:
- Background color from tier 2
- Subtle border in army green (`rgba(74, 93, 35, 0.1)`)
- Two-level box shadow for depth
- Rounded corners via `--pico-border-radius`

## Typography

### Headings

Semantic heading tags with Pico's built-in hierarchy:

```html
<h1>Page Title</h1>        <!-- 2rem, bold -->
<h2>Major Section</h2>      <!-- 1.75rem, bold -->
<h3>Subsection</h3>         <!-- 1.5rem, bold -->
<h4>Component Title</h4>    <!-- 1.25rem, bold -->
```

### Text Colors

```html
<p class="text-muted">Secondary information</p>
```

**Note**: `text-muted` is a Pico CSS utility, not a custom class.

## Layout Components

### Grid System

Pico CSS provides a responsive `.grid` class:

```html
<!-- Automatically responsive: 1 column mobile, 2+ desktop -->
<div class="grid">
    <article>Column 1</article>
    <article>Column 2</article>
    <article>Column 3</article>
</div>
```

### Custom Grid Variants

For specific layouts, we define semantic variants:

```css
.donation-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;  /* Impact + Form */
    gap: 2rem;
}

@media (max-width: 768px) {
    .donation-grid {
        grid-template-columns: 1fr;  /* Stack on mobile */
    }
}
```

## Component Patterns

### Navigation

```html
<nav class="avr-navigation">
    <div class="container">
        <div class="header-grid">
            <a href="/" class="logo-link">
                <img src="/logo.svg" alt="AVR Logo" class="main-logo">
            </a>
            <div class="nav-links">
                <a href="/about">About</a>
                <a href="/donate" role="button">Donate</a>
            </div>
        </div>
    </div>
</nav>
```

**Classes**:
- `.avr-navigation` - Main navigation container
- `.header-grid` - Logo + links layout
- `.nav-links` - Navigation link group

### Blog Components

```html
<section id="blog">
    <div class="blog-header">
        <h2>Latest News</h2>
        <p class="text-muted">Description</p>
    </div>
    
    <div class="blog-posts">
        <article class="blog-post-item">
            <h3 class="post-title">Post Title</h3>
            <p class="post-excerpt">Excerpt...</p>
            <small class="text-muted">Published: Jan 1, 2025</small>
        </article>
    </div>
</section>
```

**Classes**:
- `.blog-header` - Blog section header
- `.blog-posts` - Container for post list
- `.blog-post-item` - Individual post card
- `.post-title`, `.post-excerpt` - Post content elements

### Donation Components

```html
<div class="donation-grid">
    <div class="donation-impact">
        <h3>Your Impact</h3>
        <div class="grid">
            <article class="impact-item">
                <h4>Title</h4>
                <p>Description</p>
            </article>
        </div>
    </div>
    
    <div class="donation-form">
        <div class="donation-card">
            <form>
                <div class="donation-amounts">
                    <div class="amount-grid">
                        <button type="button" class="outline amount-btn">$25</button>
                        <button type="button" class="outline amount-btn">$50</button>
                    </div>
                </div>
            </form>
        </div>
    </div>
</div>
```

**Classes**:
- `.donation-grid` - Impact + form 2-column layout
- `.donation-impact` - Impact section
- `.impact-item` - Individual impact card
- `.donation-form` - Form container
- `.donation-card` - Form card wrapper
- `.donation-amounts` - Amount selection section
- `.amount-grid` - 3x2 button grid
- `.amount-btn` - Amount button (with Pico's `.outline`)

### Contact Components

```html
<section id="contact" class="grid">
    <article class="contact-info">
        <div class="info-section">
            <h3>Get In Touch</h3>
            <div class="grid social-links">
                <a href="mailto:...">Email</a>
                <a href="tel:...">Phone</a>
            </div>
        </div>
    </article>
    
    <article class="contact-form-wrapper">
        <form>
            <!-- Form fields -->
        </form>
    </article>
</section>
```

**Classes**:
- `.info-section` - Contact information container
- `.social-links` - Social media / contact links grid

### Team Components

```html
<section class="team-intro">
    <h1>Meet The Team</h1>
    <p>Description...</p>
</section>

<section class="team-members">
    <div class="grid">
        <article class="team-card">
            <header class="team-card-image">
                <img src="/member.jpg" alt="Name">
                <div class="team-card-overlay">
                    <small>Title</small>
                    <h2>Name</h2>
                </div>
            </header>
            <div class="team-card-content">
                <p>Bio...</p>
            </div>
        </article>
    </div>
</section>
```

**Classes**:
- `.team-intro` - Centered intro section
- `.team-members` - Team member grid container
- `.team-card` - Individual member card
- `.team-card-image` - Image container with overlay
- `.team-card-overlay` - Gradient overlay on image
- `.team-card-content` - Bio content area

### Admin Components

```html
<div class="admin-container">
    <div class="admin-card">
        <header class="admin-header">
            <h2>Manage Posts</h2>
            <div class="admin-actions">
                <a href="/admin/posts/create" role="button">Create Post</a>
            </div>
        </header>
        
        <div class="admin-table-wrapper">
            <table class="admin-table">
                <thead>
                    <tr>
                        <th>Title</th>
                        <th>Status</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>
                            <span class="post-title">Title</span>
                            <small class="post-excerpt">Excerpt</small>
                        </td>
                        <td>
                            <span class="status-badge published">Published</span>
                        </td>
                        <td class="post-actions">
                            <a href="/edit" role="button" class="outline small">Edit</a>
                            <button type="button" class="outline small delete-action">Delete</button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</div>
```

**Classes**:
- `.admin-container` - Max-width container for admin pages
- `.admin-container.narrow` - Narrower container for forms
- `.admin-card` - Card wrapper for admin content
- `.admin-header` - Header with title + actions
- `.admin-actions` - Action button group
- `.admin-table-wrapper` - Responsive table container
- `.admin-table` - Table with admin-specific styling
- `.admin-table-actions` - Table action button group
- `.post-title` - Post title in table
- `.post-excerpt` - Post excerpt in table
- `.post-date` - Post date
- `.post-actions` - Action buttons cell
- `.status-badge` - Status indicator
  - `.status-badge.published` - Published status (green)
  - `.status-badge.draft` - Draft status (gray)
- `.delete-action` - Delete button (red)
- `.admin-empty` - Empty state message
- `.admin-nav` - Admin navigation breadcrumb
- `.form-field` - Form field wrapper
- `.form-actions` - Form action buttons
- `button.small` - Small button modifier

## Button Patterns

### Pico CSS Roles

```html
<!-- Primary action -->
<button>Default Button</button>

<!-- Secondary action -->
<button class="secondary">Secondary</button>

<!-- Outline style -->
<button class="outline">Outline</button>

<!-- Contrast (high emphasis) -->
<button class="contrast">Contrast</button>
```

### Custom Button Utilities

```html
<!-- Full-width button -->
<button class="full-width">Sign In</button>

<!-- Small button -->
<button class="small">Edit</button>

<!-- Delete action -->
<button class="delete-action">Delete</button>
```

## Form Patterns

### Standard Form

```html
<form>
    <div class="form-field">
        <label for="name">Name</label>
        <input type="text" id="name" name="name" required>
    </div>
    
    <div class="form-actions">
        <button type="submit">Submit</button>
        <a href="/cancel" role="button" class="secondary">Cancel</a>
    </div>
</form>
```

### Validation States

Pico CSS provides automatic validation styling:

```html
<input type="email" required aria-invalid="true">
<small>Please enter a valid email</small>
```

## Responsive Design

### Breakpoints

```css
/* Mobile-first approach */

/* Tablet: 768px+ */
@media (min-width: 768px) {
    .grid { grid-template-columns: repeat(2, 1fr); }
}

/* Desktop: 1024px+ */
@media (min-width: 1024px) {
    .grid { grid-template-columns: repeat(3, 1fr); }
}
```

### Mobile Patterns

```css
/* Stack on mobile */
@media (max-width: 768px) {
    .donation-grid {
        grid-template-columns: 1fr;
    }
    
    .amount-grid {
        grid-template-columns: 1fr 1fr;  /* 2 columns instead of 3 */
    }
}
```

## Accessibility

### Color Contrast

**WCAG AA Minimum**:
- Normal text: 4.5:1 contrast ratio
- Large text (18pt+): 3:1 contrast ratio
- Interactive elements: 3:1 against background

**Our Standards**:
- Body text: 7:1+ (AAA)
- Headings: 7:1+ (AAA)
- Buttons: 4.5:1+ (AA)
- Borders: 3:1+ (AA)

### Focus States

All interactive elements have visible focus states via Pico CSS:

```css
/* Pico automatically provides focus styles */
button:focus-visible {
    outline: var(--pico-outline-width) solid var(--pico-primary-focus);
}
```

### Semantic HTML

Always use the correct HTML5 semantic elements:

```html
<header> - Page/section header
<nav>    - Navigation
<main>   - Main content
<article>- Self-contained content
<section>- Thematic grouping
<aside>  - Complementary content
<footer> - Page/section footer
```

## CSS Architecture

### File Structure

```
pb_public/assets/css/
├── pico.min.css          # Pico CSS framework
├── custom.css            # AVR custom styles (THIS IS WHERE WE WORK)
├── quill.snow.css        # Quill editor theme
└── quill-custom.css      # Quill overrides
```

### custom.css Organization

```css
/**
 * 1. Pico CSS Variable Overrides
 * 2. Typography
 * 3. HTMX Loading Indicators
 * 4. Header & Navigation
 * 5. Semantic Page Sections (#mission, #blog, etc.)
 * 6. Donation Components
 * 7. Contact Page Components
 * 8. Blog Components
 * 9. Team Page Components
 * 10. Admin Panel Components
 * 11. Form Validation
 * 12. Quill Editor Overrides
 */
```

### Naming Conventions

**Component Classes**: 
- Use noun-based names: `.blog-post`, `.team-card`, `.donation-form`
- Nested elements: `.card-header`, `.card-content`, `.card-footer`

**Modifier Classes**:
- Use adjectives: `.admin-container.narrow`, `.status-badge.published`

**Utility Classes** (minimal):
- Pico CSS utilities only: `.text-muted`, `.grid`, `.container`
- Custom utilities only when absolutely necessary: `.full-width`, `.small`

**Anti-Pattern**:
```css
/* ❌ DON'T: Tailwind-style utilities */
.mt-4 { margin-top: 1rem; }
.bg-gray-100 { background: #f5f5f5; }

/* ✅ DO: Semantic components */
.section-spacing { margin-top: var(--pico-spacing); }
.card-background { background: var(--pico-card-background-color); }
```

## Common Patterns Cheatsheet

### Card with Header and Actions

```html
<article>
    <header>
        <h3>Title</h3>
        <small>Subtitle</small>
    </header>
    <p>Content</p>
    <footer>
        <button>Action</button>
    </footer>
</article>
```

### Two-Column Layout

```html
<div class="grid">
    <div>Left column</div>
    <div>Right column</div>
</div>
```

### Form with Validation

```html
<form>
    <label for="field">Label</label>
    <input type="text" id="field" name="field" required aria-invalid="false">
    <small>Helper text</small>
    
    <button type="submit">Submit</button>
</form>
```

### Status Badge

```html
<span class="status-badge published">Published</span>
<span class="status-badge draft">Draft</span>
```

## Migration Guide

### From Inline Styles

**Before**:
```html
<div style="margin-bottom: 2rem; padding: 1rem; background: white;">
```

**After**:
```html
<div class="card-section">  <!-- Define in custom.css -->
```

### From Tailwind Utilities

**Before**:
```html
<div class="max-w-4xl mx-auto mt-8 bg-white rounded-lg shadow-md">
```

**After**:
```html
<div class="admin-container">  <!-- Define in custom.css -->
```

```css
/* custom.css */
.admin-container {
    max-width: 1200px;
    margin: 0 auto;
    margin-top: 2rem;
    background: var(--pico-card-background-color);
    border-radius: var(--pico-border-radius);
    box-shadow: var(--pico-card-box-shadow);
}
```

## Testing Checklist

When adding new components:

- [ ] Uses semantic HTML5 elements
- [ ] All text meets WCAG AA contrast (4.5:1+)
- [ ] Focus states visible on all interactive elements
- [ ] Works in light and dark modes
- [ ] Responsive on mobile, tablet, desktop
- [ ] Uses Pico CSS variables (never hardcoded colors/spacing)
- [ ] No inline styles
- [ ] No utility classes (except Pico's)
- [ ] Semantic class names
- [ ] Documented in this file

## Resources

- **Pico CSS Docs**: https://picocss.com/docs
- **Pico CSS Variables**: https://picocss.com/docs/css-variables
- **WCAG Guidelines**: https://www.w3.org/WAI/WCAG21/quickref/
- **HTML5 Elements**: https://developer.mozilla.org/en-US/docs/Web/HTML/Element

## Changelog

### October 11, 2025 - Phase 1 & 2 Complete
- Removed all inline styles from templates (117+ instances)
- Removed all Tailwind utility classes (83+ instances)
- Implemented 3-tier color system
- Added comprehensive component library
- Refactored admin panel components
- Added team page components
- Enhanced card styling with borders and shadows
- Achieved 100% Pico CSS compatibility
