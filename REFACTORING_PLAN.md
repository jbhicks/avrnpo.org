# Buffalo Refactoring Plan

## Critical Issues to Fix

### 1. Asset Serving (Priority 1)
**Problem**: Complex asset serving causing test failures
**Solution**: Simplify to Buffalo best practices

```go
// Replace complex asset logic with:
app.ServeFiles("/", assetsFS) // Single line, embed all assets
```

### 2. Middleware Architecture (Priority 2) 
**Problem**: 200+ character middleware skip line
**Solution**: Group routes by authentication needs

```go
// Public routes
public := app.Group("/")
public.GET("/", HomeHandler)

// Auth required routes  
auth := app.Group("/")
auth.Use(SetCurrentUser)
auth.Use(Authorize)
auth.GET("/account", AccountSettings)

// Admin routes
admin := app.Group("/admin")
admin.Use(SetCurrentUser)
admin.Use(AdminRequired)
```

### 3. Test Simplification (Priority 3)
**Problem**: Testing infrastructure instead of business logic
**Solution**: Remove/skip asset tests, focus on critical functionality

### 4. Route Organization (Priority 4)
**Problem**: 40+ routes in one block
**Solution**: Group by logical purpose

## Benefits Expected
- Simpler asset serving
- Clearer route organization  
- Easier testing
- Better maintainability
- Reduced entropy

## Implementation Order
1. Fix asset serving (15 minutes)
2. Refactor middleware (30 minutes)
3. Remove/simplify asset tests (10 minutes)
4. Organize routes into groups (45 minutes)

Total estimated time: 1.5 hours for significant improvement
