# Testing Guide

How to properly test the AVR NPO donation system with Buffalo framework.

## 🚨 CRITICAL: Buffalo Testing Rules

**NEVER use `go test` directly** - Buffalo requires special setup.

**✅ Always use these Makefile commands:**

```bash
# Comprehensive test suite (recommended)
make test

# Quick testing (assumes database running)  
make test-fast

# Automatic database management
make test-resilient
```

## 🎯 Why Buffalo Testing is Special

Buffalo tests need special setup that `go test` alone cannot provide:

- **PostgreSQL connection** - Test database must be running
- **Environment variables** - `GO_ENV=test` properly configured
- **Database migrations** - Test database with proper schema
- **Transaction isolation** - ActionSuite handles test data cleanup
- **Session management** - Buffalo test session handling

## 🧪 Test Commands Explained

### `make test-fast` (Most Common)
- **Use when**: Buffalo is already running, database is up
- **Speed**: Fast - assumes environment is ready
- **Best for**: Regular development testing

### `make test` (Comprehensive) 
- **Use when**: Full verification needed
- **Speed**: Slower - sets up everything
- **Best for**: CI/CD, major changes, troubleshooting

### `make test-resilient` (Automatic)
- **Use when**: Database issues or unknown state
- **Speed**: Medium - handles database setup
- **Best for**: After database changes or when unsure

## 📋 Testing Workflow

### Daily Development Testing
```bash
# 1. Start development environment (once)
make dev

# 2. Make your code changes

# 3. Run tests (Buffalo keeps running)
make test-fast

# 4. Continue development
```

### After Database Changes
```bash
# Run migrations first
soda migrate up

# Then test with resilient option
make test-resilient
```

### Full System Verification
```bash
# Stop Buffalo if running (Ctrl+C)
# Run comprehensive test suite
make test
```

## ✅ Buffalo Testing Best Practices

### ActionSuite Pattern
```go
// ✅ Correct test structure
func (as *ActionSuite) Test_DonationFlow() {
    res := as.HTML("/donation").Get()
    as.Equal(http.StatusOK, res.Code)
    as.Contains(res.Body.String(), "Donate to AVR")
}
```

### User Creation with Unique Data
```go
// ✅ Avoid database conflicts
timestamp := time.Now().UnixNano()
user := &models.User{
    Email: fmt.Sprintf("test-%d@example.com", timestamp),
    // ... other fields
}
```

### Test Both Direct and HTMX Navigation
```go
// ✅ Test regular page load
res := as.HTML("/account/subscriptions").Get()
as.Equal(http.StatusOK, res.Code)

// ✅ Test HTMX navigation
res = as.HTML("/account/subscriptions").
    Header("HX-Request", "true").Get()
as.Equal(http.StatusOK, res.Code)
```

## 🚨 Common Testing Errors

### "Database Not Found"
```bash
# Solution: Create test database
GO_ENV=test soda create
GO_ENV=test soda migrate up
```

### "Port Already in Use"
```bash
# Check what's using port 3000
lsof -i :3000

# If it's an old Buffalo process, kill it carefully
pkill -f buffalo
```

### "Template Not Found"
- Check partial naming: `partial("dir/file")` not `partial("dir/_file.plush.html")`
- Verify template file exists with correct underscore prefix

### "Connection Refused"
```bash
# Check if PostgreSQL is running
podman-compose ps

# Start database if needed
podman-compose up -d postgres
```

## 🧪 Test Categories

### Unit Tests
- Test individual functions and methods
- Fast execution, no external dependencies
- Focus on business logic

### Integration Tests  
- Test database interactions
- Test API endpoints
- Test payment flow integration

### ActionSuite Tests
- Test full HTTP request/response cycle
- Test authentication and authorization
- Test template rendering

## 📊 Test Coverage

### Payment System Testing
- Donation form validation
- Payment initialization flow
- Subscription management operations
- Webhook event processing

### Authentication Testing
- User login/logout flows
- Session management
- Role-based access control
- Password reset functionality

### Template Testing
- Page rendering without errors
- Partial inclusion works correctly
- HTMX navigation functions properly
- Form submissions process correctly

## 🔧 Debugging Test Failures

### 1. Check Buffalo Logs
```bash
# Buffalo console shows detailed error information
# Look for stack traces and SQL queries
```

### 2. Database State Issues
```bash
# Reset test database to clean state
GO_ENV=test soda reset
make test-fast
```

### 3. Environment Problems
```bash
# Verify environment variables
echo $GO_ENV
# Should be "test" during testing

# Check database configuration
cat database.yml
```

### 4. Template Errors
- Verify partial paths don't include underscores
- Check that all referenced variables exist in context
- Ensure template files have correct extensions

## 📋 Pre-Commit Testing

Before committing code changes:

```bash
# 1. Run quick tests
make test-fast

# 2. If database changes were made
make test-resilient

# 3. For major changes  
make test
```

## 🎯 Continuous Integration

For CI/CD environments:

```bash
# Full setup and test cycle
make test
```

This ensures:
- Database is properly initialized
- All migrations are applied
- Complete test suite runs
- Environment is properly configured

## 🆘 When Tests Still Fail

1. **Check [Buffalo Troubleshooting](../buffalo-framework/troubleshooting.md)** for Buffalo-specific issues
2. **Verify [Development Workflow](./development-workflow.md)** for environment setup
3. **Review logs** in Buffalo console for specific error details
4. **Reset environment** with `make clean && make dev`
