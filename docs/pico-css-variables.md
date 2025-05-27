# Pico.css CSS Variables Documentation

> **Source**: https://picocss.com/docs/css-variables

Customize Pico's design system with over 130 CSS variables to create a unique look and feel.

## Overview

Pico includes many custom properties (variables) that allow easy access to frequently used values such as:
- `font-family`
- `font-size`
- `border-radius`
- `margin`
- `padding`
- Colors and color schemes
- Spacing and typography

## Key Principles

### Prefixed Variables
All CSS variables are prefixed with `--pico-` to avoid collisions with other CSS frameworks or your own vars. You can remove or customize this prefix by recompiling the CSS files with SASS.

### Global vs Local Application
- **Global**: Define CSS variables within the `:root` selector to apply changes globally
- **Local**: Overwrite CSS variables on specific selectors to apply changes locally

## Example Usage

```css
:root {
  --pico-border-radius: 2rem;
  --pico-typography-spacing-vertical: 1.5rem;
  --pico-form-element-spacing-vertical: 1rem;
  --pico-form-element-spacing-horizontal: 1.25rem;
}

h1 {
  --pico-font-family: Pacifico, cursive;
  --pico-font-weight: 400;
  --pico-typography-spacing-vertical: 0.5rem;
}

button {
  --pico-font-weight: 700;
}
```

## Color Schemes

### Light Mode (Default)
To add or edit CSS variables for light mode only:

```css
/* Light color scheme (Default) */
/* Can be forced with data-theme="light" */
[data-theme="light"],
:root:not([data-theme="dark"]) {
  --pico-color: #000;
  --pico-background-color: #fff;
  /* ... other light mode variables */
}
```

### Dark Mode
To add or edit CSS variables for dark mode, define them twice:

1. **Auto Dark Mode** (based on user's device settings):
```css
/* Dark color scheme (Auto) */
/* Automatically enabled if user has Dark mode enabled */
@media only screen and (prefers-color-scheme: dark) {
  :root:not([data-theme]) {
    --pico-color: #fff;
    --pico-background-color: #000;
    /* ... other dark mode variables */
  }
}
```

2. **Forced Dark Mode** (manual toggle):
```css
/* Dark color scheme (Forced) */
/* Enabled if forced with data-theme="dark" */
[data-theme="dark"] {
  --pico-color: #fff;
  --pico-background-color: #000;
  /* ... other dark mode variables */
}
```

## Variable Categories

There are two main categories of CSS variables:

1. **Style variables** - Do not depend on color scheme (typography, spacing, borders)
2. **Color variables** - Depend on color scheme (backgrounds, text colors, borders)

## Common CSS Variables

### Typography
- `--pico-font-family`
- `--pico-font-size`
- `--pico-font-weight`
- `--pico-line-height`
- `--pico-typography-spacing-vertical`

### Colors
- `--pico-color` - Main text color
- `--pico-background-color` - Main background
- `--pico-primary` - Primary brand color
- `--pico-secondary` - Secondary color
- `--pico-muted-color` - Muted text
- `--pico-muted-border-color` - Subtle borders

### Spacing
- `--pico-spacing`
- `--pico-form-element-spacing-vertical`
- `--pico-form-element-spacing-horizontal`

### Borders
- `--pico-border-radius`
- `--pico-border-width`
- `--pico-outline-width`

### Form Elements
- `--pico-form-element-background-color`
- `--pico-form-element-border-color`
- `--pico-form-element-color`

## Theme Implementation in Buffalo SaaS Template

Our template implements theme switching using:

```javascript
function setTheme(theme) {
  if (theme === 'auto') {
    localStorage.removeItem('picoPreferredColorScheme');
    document.documentElement.removeAttribute('data-theme');
  } else {
    localStorage.setItem('picoPreferredColorScheme', theme);
    document.documentElement.setAttribute('data-theme', theme);
  }
}
```

## Best Practices

1. **Use semantic variable names** - Prefer `--pico-primary` over hardcoded colors
2. **Test both themes** - Always verify customizations work in light and dark modes
3. **Override sparingly** - Use Pico's design system as much as possible
4. **Maintain accessibility** - Ensure color contrast ratios remain compliant
5. **Consider auto theme** - Respect user's system preferences when possible

## Resources

- [Pico.css Documentation](https://picocss.com/docs)
- [Pico.css GitHub](https://github.com/picocss/pico)
- [CSS Variables Specification](https://developer.mozilla.org/en-US/docs/Web/CSS/Using_CSS_custom_properties)
