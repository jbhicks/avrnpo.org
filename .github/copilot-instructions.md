# GitHub Copilot Instructions for AVR NPO Website

## Project Context
This is the official website for American Veterans Rebuilding (AVR), a 501(c)(3) nonprofit organization. The project handles donations through Helcim API and requires high security standards.

## ⚠️ CRITICAL: Always Read README First
**ALWAYS** check `README.md` before starting any work to understand:
- Current project status and phase
- What task is next in the improvement plan
- Which files need to be modified
- Progress tracking with checkboxes

## Technology Stack & Constraints

### Languages & Frameworks
- **Go**: Primary backend language with Gin framework
- **HTMX**: For frontend interactivity (prefer over vanilla JavaScript)
- **Tailwind CSS + DaisyUI**: For styling (ALWAYS use DaisyUI components)
- **HTML Templates**: Go template system

### Code Style Requirements
1. **Minimal Comments**: Don't over-comment code, only explain complex logic
2. **Clean Go Practices**: Follow standard Go conventions
3. **Security First**: This handles financial transactions - security is paramount
4. **DaisyUI Components**: Always prefer DaisyUI over custom Tailwind classes
5. **NEVER** overwrite Tailwind classes in CSS files - remove the class from HTML instead

### Architecture Patterns
- RESTful API endpoints with proper HTTP methods
- Template-based rendering with HTMX for dynamic content
- Environment-based configuration
- Structured error handling and logging

## Current Development Focus

### Active Improvement Plan
The project is following a phased improvement plan tracked in README.md:

1. **Phase 1**: Security & API Improvements (CURRENT)
2. **Phase 2**: Webhook Integration  
3. **Phase 3**: Database Integration
4. **Phase 4**: Receipt & Email System
5. **Phase 5**: Admin Dashboard

### Immediate Priorities
- Convert donation API from GET to POST (security improvement)
- Add proper request validation
- Implement webhook handling for payment confirmations

## Development Environment
- **OS**: WSL Arch Linux
- **Terminal**: Configured for Bash in WSL
- **Port**: 3001 (default, configurable via PORT env var)
- **Hot Reload**: Manual restart required

## File Structure Guidelines
```
/
├── main.go              # Main application with all routes
├── templates/           # HTML templates
├── static/             # CSS, JS, images
├── .env                # Environment variables (never commit)
├── README.md           # Project status and improvement plan
└── .github/
    └── copilot-instructions.md  # This file
```

## Security Considerations
- Validate all user inputs
- Use POST for sensitive data (never GET with query params)
- Implement proper error handling without exposing internals
- Log security events
- Validate API tokens and signatures
- Use HTTPS in production

## Common Tasks & Patterns

### Adding New Routes
```go
r.POST("/api/endpoint", func(c *gin.Context) {
    // Always validate input first
    var request struct {
        Field string `json:"field" binding:"required"`
    }
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }
    
    // Process request...
    c.JSON(http.StatusOK, gin.H{"result": "success"})
})
```

### Frontend HTMX Patterns
```html
<!-- Prefer HTMX over JavaScript -->
<form hx-post="/api/endpoint" hx-target="#result">
    <input type="text" name="field" class="input input-bordered" required>
    <button type="submit" class="btn btn-primary">Submit</button>
</form>
<div id="result"></div>
```

### DaisyUI Component Usage
```html
<!-- Always use DaisyUI classes -->
<div class="card bg-base-100 shadow-xl">
    <div class="card-body">
        <h2 class="card-title">Title</h2>
        <div class="card-actions justify-end">
            <button class="btn btn-primary">Action</button>
        </div>
    </div>
</div>
```

## Critical Don'ts (Past Mistakes to Avoid)
1. **NEVER** add excessive comments explaining what you're doing - only explain complex logic
2. **NEVER** overwrite Tailwind classes in CSS files - remove the class from HTML instead
3. **NEVER** use vanilla JavaScript when HTMX can accomplish the task
4. **NEVER** create custom components when DaisyUI has an equivalent
5. **NEVER** use GET requests for sensitive data like payment information

## Reminder Checklist for Every Session

### Before Starting Work
- [ ] Read README.md to understand current status
- [ ] Check which phase and task is active in README.md
- [ ] Review PROJECT_TRACKING.md for detailed task requirements
- [ ] Review environment setup (.env, port 3001)

### During Development
- [ ] Consider security implications for any changes
- [ ] Use DaisyUI components for any UI changes
- [ ] Follow minimal commenting style
- [ ] Test changes thoroughly

### After Completing Work - CRITICAL DOCUMENTATION UPDATES
- [ ] **Update README.md progress checkboxes** when tasks are completed
- [ ] **Update PROJECT_TRACKING.md** with task status changes
- [ ] **Update this instructions file** if new patterns or issues are discovered
- [ ] Mark completed tasks with ✅ and update status from 📋 PLANNED to 🔄 IN PROGRESS to ✅ COMPLETED
- [ ] Move to next task in sequence and update current priority
- [ ] Document any new issues or improvements discovered during development

## Contact & Support
This is a nonprofit project handling financial transactions. Always prioritize security and reliability over speed of development.