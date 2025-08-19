# Buffalo Routing and HTMX Integration Guide

*Updated: June 24, 2025*

This document addresses critical issues with Buffalo routing, asset pipeline configuration, and HTMX integration based on official Buffalo documentation and best practices.

## 🚨 CRITICAL ISSUES IDENTIFIED

### 1. Asset Pipeline Not Properly Configured
**Problem**: Assets are served directly from `/public/` without Buffalo's asset pipeline
- No asset fingerprinting configured
- No manifest.json for cache busting
- Direct file serving instead of Buffalo asset helpers
- CSS/JS loads fail on page refresh due to improper asset handling

### 2. HTMX Integration Issues
**Problem**: Current HTMX setup doesn't follow Buffalo best practices
- Using custom layout switching instead of proper render patterns
- Not leveraging Buffalo's built-in template rendering for HTMX
- Complex rendering logic that could be simplified

### 3. Route Structure Issues
**Problem**: Inconsistent route naming and organization
- Route `/donate` exists but user tried `/donations`
- No proper asset route handling
- Missing asset fingerprinting support

## ✅ OFFICIAL BUFFALO PATTERNS

### Asset Pipeline Best Practices

**From Buffalo Documentation**:
> Buffalo uses asset fingerprinting to generate hashed file names (e.g., `application.a8adff90f4c6d47529c4.js`)
> This allows for proper caching while ensuring cache invalidation when files change.

**Required Setup**:
1. **Asset Helpers**: Use `stylesheetTag()` and `javascriptTag()` instead of direct `<link>` tags
2. **Manifest File**: Buffalo expects `/public/assets/manifest.json` for asset mapping
3. **Asset Box**: Configure `AssetsBox` in render options for proper asset serving

### Template Rendering Patterns

**Standard Layout Usage**:
```go
// actions/render.go
r = render.New(render.Options{
    HTMLLayout:  "application.plush.html",
    TemplatesFS: templates.FS(),
    AssetsFS:    public.FS(),
})
```

**HTMX Integration**:
```go
// In handler
if IsHTMX(c.Request()) {
    return c.Render(http.StatusOK, r.HTML("partial.html"))
} else {
    return c.Render(http.StatusOK, r.HTML("full_page.html"))
}
```

## 🔧 REQUIRED FIXES

### 1. Fix Asset Pipeline Configuration

**Current Problem**:
```html
<!-- WRONG: Direct asset links -->
<link rel="stylesheet" href="/css/pico.min.css">
<script src="/js/htmx.min.js"></script>
```

**Buffalo Best Practice**:
```html
<!-- CORRECT: Use Buffalo asset helpers -->
<%= stylesheetTag("application.css") %>
<%= javascriptTag("application.js") %>
```

### 2. Implement Proper HTMX Pattern

**Current Complex Pattern**:
- Multiple render engines (r, rHTMX, rNoLayout)
- Custom layout switching logic
- Manual HTMX request detection

**Buffalo HTMX Best Practice**:
```go
func MyHandler(c buffalo.Context) error {
    data := getData()
    c.Set("data", data)
    
    // Let Buffalo handle the rendering
    if c.Request().Header.Get("HX-Request") == "true" {
        return c.Render(http.StatusOK, r.HTML("content_partial.html"))
    }
    return c.Render(http.StatusOK, r.HTML("full_page.html"))
}
```

### 3. Fix Route Organization

**Add Missing Routes**:
```go
// Add redirect for common variations
app.GET("/donations", func(c buffalo.Context) error {
    return c.Redirect(http.StatusMovedPermanently, "/donate")
})
```

## 📋 IMPLEMENTATION PLAN

### Phase 1: Asset Pipeline Fix
1. **Configure Asset Fingerprinting**
   - Set up Webpack or asset build process
   - Generate manifest.json file
   - Update render.go to use AssetsBox

2. **Update Templates**
   - Replace direct asset links with Buffalo helpers
   - Use `stylesheetTag()` and `javascriptTag()`
   - Remove hardcoded asset paths

### Phase 2: HTMX Integration Cleanup
1. **Simplify Render Logic**
   - Remove multiple render engines
   - Use single render engine with proper templates
   - Implement Buffalo HTMX pattern

2. **Template Structure**
   - Create proper full-page templates
   - Create HTMX partial templates
   - Use Buffalo's built-in template resolution

### Phase 3: Route Optimization
1. **Add Route Aliases**
   - Add redirects for common variations
   - Implement proper SEO redirects
   - Organize routes by functionality

2. **Asset Route Handling**
   - Ensure proper asset serving
   - Configure cache headers
   - Handle development vs production

## 🛠️ SPECIFIC FIXES NEEDED

### Update application.plush.html
```html
<head>
    <!-- Replace direct links -->
    <%= stylesheetTag("pico.min.css") %>
    <%= stylesheetTag("application.css") %>
    
    <!-- At end of body -->
    <%= javascriptTag("application.js") %>
    <%= javascriptTag("htmx.min.js") %>
</head>
```

### Update render.go
```go
func init() {
    r = render.New(render.Options{
        HTMLLayout:  "application.plush.html",
        TemplatesFS: templates.FS(),
        AssetsFS:    public.FS(),
        AssetsBox:   public.FS(), // Add this for proper asset serving
        Helpers:     commonHelpers,
    })
}
```

### Update app.go routes
```go
// Add route aliases and redirects
app.GET("/donations", func(c buffalo.Context) error {
    return c.Redirect(http.StatusMovedPermanently, "/donate")
})

// Ensure asset serving
app.ServeFiles("/assets", public.FS())
```

## 🔍 DIAGNOSTIC STEPS

### Check Asset Pipeline Status
```bash
# Check if manifest.json exists
ls -la public/assets/manifest.json

# Check asset compilation
buffalo build --extract-assets

# Test asset helpers in templates
grep -r "stylesheetTag\|javascriptTag" templates/
```

### Test HTMX Integration
```bash
# Test with HTMX header
curl -H "HX-Request: true" http://localhost:3000/donate

# Test without HTMX header
curl http://localhost:3000/donate
```

## 📚 REFERENCES

### Buffalo Documentation
- [Asset Pipeline](https://gobuffalo.io/documentation/frontend-layer/assets/)
- [Rendering](https://gobuffalo.io/documentation/frontend-layer/rendering/)
- [Layouts](https://gobuffalo.io/documentation/frontend-layer/layouts/)
- [Routing](https://gobuffalo.io/documentation/request_handling/routing/)

### Key Buffalo Concepts
- **Asset Fingerprinting**: Automatic hash generation for cache busting
- **Asset Helpers**: `stylesheetTag()`, `javascriptTag()`, `assetPath()`
- **Template Resolution**: Automatic .html extension and layout application
- **Render Engines**: Single engine with multiple layout options

---

*This guide addresses the root causes of styling loss on page refresh and improper HTMX integration in Buffalo applications.*
