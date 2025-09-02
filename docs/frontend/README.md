# Frontend Development

User interface, styling, and interactive patterns for the AVR NPO donation system.

## 📋 Frontend Documentation

### 🎨 Styling with Pico CSS
- **[Pico CSS Guide](./pico-css.md)** - CSS variables, theming, and customization
- **[Pico Implementation](./pico-implementation.md)** - Semantic HTML patterns and best practices

### ⚡ Interactive Patterns  
- **[HTMX Patterns](./htmx-patterns.md)** - Progressive enhancement and navigation
- **[HTMX Reference](./htmx-reference.md)** - Complete HTMX integration guide

### 📦 Asset Management
- **[Assets](./assets.md)** - Asset pipeline, optimization, and serving *(planned)*

## 🎯 Current Frontend Stack

### Core Technologies
- **Pico CSS** - Semantic styling with CSS variables
- **HTMX** - Progressive enhancement and AJAX navigation
- **Buffalo Asset Pipeline** - CSS/JS compilation and serving
- **Plush Templates** - Server-side HTML rendering

### Styling Philosophy
- **Semantic HTML first** - Use proper HTML elements before adding classes
- **CSS variables for theming** - Support light/dark modes automatically
- **Progressive enhancement** - Works without JavaScript, better with it
- **Minimal custom CSS** - Leverage Pico's built-in styles

## 🧭 Navigation Patterns

### User Experience Flow
1. **Landing pages** - Clean, accessible, fast-loading
2. **Donation flow** - Streamlined, secure, trustworthy
3. **Account management** - Clear navigation, helpful feedback
4. **Admin interface** - Functional, efficient, well-organized

### HTMX Integration
- **Page navigation** - Smooth transitions without full reloads
- **Form submissions** - Inline validation and feedback
- **Dynamic content** - Real-time updates where appropriate
- **Fallback support** - Works with JavaScript disabled

## 🎨 Design System

### Color Scheme
- **Primary colors** - AVR brand colors for CTAs and headers
- **Semantic colors** - Success, warning, error states
- **Theme support** - Automatic light/dark mode switching
- **Accessibility** - WCAG AA contrast compliance

### Typography
- **System fonts** - Fast loading, excellent readability
- **Scale hierarchy** - Clear information hierarchy
- **Responsive text** - Scales appropriately across devices

### Layout Patterns
- **Container widths** - Optimal reading line lengths
- **Grid systems** - Flexible, responsive layouts
- **Spacing scale** - Consistent visual rhythm
- **Component spacing** - Predictable element relationships

## 🧪 Frontend Testing

### Manual Testing Checklist
- [ ] **Accessibility** - Keyboard navigation, screen readers
- [ ] **Responsive design** - Mobile, tablet, desktop layouts  
- [ ] **Theme switching** - Light/dark mode functionality
- [ ] **Progressive enhancement** - Works without JavaScript
- [ ] **Performance** - Fast loading, minimal resource usage

### Browser Support
- **Modern browsers** - Chrome, Firefox, Safari, Edge
- **Mobile browsers** - iOS Safari, Chrome Mobile
- **Graceful degradation** - Basic functionality in older browsers

## 🔧 Development Workflow

### Making Frontend Changes
1. **Check existing patterns** in Pico CSS documentation
2. **Use CSS variables** instead of hardcoded values
3. **Test both themes** - Light and dark modes
4. **Verify accessibility** - Keyboard and screen reader support
5. **Test responsiveness** - Multiple screen sizes

### Asset Pipeline
- **Development** - Files served individually with hot reload
- **Production** - Concatenated and minified automatically
- **Caching** - Proper cache headers for performance

## 🎯 Key Implementation Guidelines

### Pico CSS Best Practices
- Use semantic HTML elements (`<article>`, `<section>`, `<nav>`)
- Leverage built-in classes (`secondary`, `outline`, `contrast`)
- Customize via CSS variables (`--pico-primary`, `--pico-background-color`)
- Avoid utility classes or extensive custom CSS

### HTMX Integration
- Progressive enhancement for better user experience
- Graceful fallback when JavaScript unavailable
- Proper HTTP status codes for HTMX responses
- Clear loading states and error handling

### Performance Considerations
- **Minimal CSS** - Only necessary styles loaded
- **Efficient HTMX** - Strategic use of AJAX requests
- **Image optimization** - Proper formats and sizes
- **Caching strategy** - Browser and CDN caching

For detailed implementation guidance, see the specific documentation files listed above.
