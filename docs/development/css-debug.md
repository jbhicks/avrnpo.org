# Proper Pico.css Implementation for Army Theme

## What We Implemented
Following official Pico.css best practices from their documentation, we now properly override colors using the exact CSS selector patterns that Pico expects.

## Pico CSS Override Pattern
Pico requires specific CSS selectors for proper color scheme handling:

### Light Mode (Default)
```css
[data-theme="light"],
:root:not([data-theme="dark"]) {
    --pico-primary: #4a5d23; /* Army Green */
    /* ... other variables ... */
}
```

### Dark Mode (Auto Detection)
```css
@media only screen and (prefers-color-scheme: dark) {
    :root:not([data-theme]) {
        --pico-primary: #6b7c32; /* Brighter Army Green for dark */
        /* ... other variables ... */
    }
}
```

### Dark Mode (Forced)
```css
[data-theme="dark"] {
    --pico-primary: #6b7c32; /* Brighter Army Green for dark */
    /* ... other variables ... */
}
```

## Army Color Scheme
- **Light Mode Primary**: `#4a5d23` (Army Green)
- **Dark Mode Primary**: `#6b7c32` (Brighter Army Green)
- **Contrast**: `#d2691e` / `#ff8c42` (Military Orange)
- **Secondary**: `#ffb627` (Army Gold)

## How It Works
1. **No theme attributes** - Let Pico handle automatic detection
2. **Proper CSS selectors** - Use Pico's exact selector patterns
3. **Complete variable sets** - Define all related variables for each color
4. **Theme-aware colors** - Different shades for light vs dark mode

This follows Pico's official documentation and ensures our army theme works in all scenarios:
- Default light mode
- Auto dark mode (based on user's system preference)
- Forced dark mode (if manually set)

## Expected Result
All buttons and primary elements should now show army green colors consistently across both light and dark modes.