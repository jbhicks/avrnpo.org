# Buffalo Test System - Debugging Summary

*Created: June 9, 2025*
*Context: Blog/Admin System Development*

## 🚨 CRITICAL: Always Use Buffalo Test Commands

**NEVER use `go test` directly in Buffalo applications!**

### Root Cause of Previous Issues

1. **PostgreSQL Version Mismatch**: 
   - Problem: pg_dump v17 generating schema incompatible with PostgreSQL v15 server
   - Solution: Upgraded PostgreSQL from v15 to v17 in docker-compose.yml
   - Result: Eliminated "transaction_timeout" errors completely

2. **Incorrect Test Command Usage**:
   - Problem: Using `go test ./actions` directly instead of Buffalo's test system
   - Solution: Use `buffalo test ./actions` which handles database setup properly
   - Result: Proper test database management and Buffalo environment setup

## ✅ Correct Buffalo Test Usage

### Commands That Work:
```bash
# Test specific packages (RECOMMENDED)
buffalo test ./actions
buffalo test ./models
buffalo test ./pkg
buffalo test ./actions ./models ./pkg

# Test with verbose output
buffalo test ./actions -v

# Test everything (USE WITH CAUTION - excludes backup dir)
buffalo test ./actions ./models ./pkg
```

### Commands to AVOID:
```bash
# ❌ DO NOT USE - bypasses Buffalo test setup
go test ./actions
go test ./...

# ❌ DO NOT USE - includes problematic backup directory
buffalo test ./...
```

## Buffalo Test Process (Automated)

When you run `buffalo test`, Buffalo automatically:

1. **Drops test database**: `[POP] dropped database avrnpo_test`
2. **Creates fresh test database**: `[POP] created database avrnpo_test`  
3. **Dumps development schema**: `[POP] dumped schema for avrnpo_development`
4. **Loads schema into test DB**: `[POP] loaded schema for avrnpo_test`
5. **Runs Go tests with Buffalo flags**: `go test -p 1 -tags development`

This ensures:
- Clean test environment for each run
- Proper Buffalo environment variables and configuration
- Database schema matches development environment
- Buffalo-specific build tags and dependencies

## 🏗️ Database Infrastructure

### PostgreSQL Configuration:
- **Version**: 17 (upgraded from 15)
- **Container**: `my_go_saas_template_postgres`
- **Port**: 5432
- **Management**: Podman Compose

### Database Management Commands:
```bash
# Check container status
podman ps

# Start/stop database
podman-compose up -d
podman-compose down

# Database operations (use soda, NOT buffalo pop)
soda create -a          # Create all databases
soda migrate up         # Run migrations
GO_ENV=test soda migrate up  # Run test migrations
soda reset              # Reset database
```

## 📁 Project Structure for Testing

### Test Files Location:
- `actions/*_test.go` - HTTP handler tests
- `models/*_test.go` - Model validation and database tests  
- `pkg/logging/*_test.go` - Logging service tests

### Exclusions:
- `backup/` directory - Contains old dependencies, exclude from testing
- Auto-generated files - Let Buffalo manage schema dumps

## 🔧 Debugging Test Issues

### If tests fail to start:
1. Check PostgreSQL is running: `podman ps`
2. Verify database connectivity: `soda --version`
3. Check for schema issues: `GO_ENV=test soda migrate status`

### If tests hang:
1. Check for infinite loops in test code
2. Verify database connections are closed properly
3. Look for deadlocks in concurrent tests

### If compilation fails:
1. Check for missing imports
2. Verify model property access syntax
3. Ensure Buffalo build tags are present

## 📋 Current State (June 9, 2025)

### ✅ Working:
- PostgreSQL 17 database infrastructure
- Buffalo test command execution
- Database schema management
- Go code compilation

### 🔧 Needs Fixing:
- Template syntax errors in blog templates
- Post model alignment with database schema
- Property access patterns in Plush templates

### 🎯 Next Steps:
1. Fix Plush template syntax for User/Author property access
2. Update Post model to match database schema exactly
3. Run full Buffalo test suite to verify all functionality
4. Document template patterns for future development

## 💡 Key Lessons

1. **Always upgrade dependencies together** - Don't mix PostgreSQL client/server versions
2. **Use Buffalo's test system** - It handles complex environment setup automatically  
3. **Test database is ephemeral** - Gets recreated on every test run, design tests accordingly
4. **Schema consistency matters** - Keep model structs aligned with database migrations
5. **Buffalo has specific patterns** - Follow Buffalo conventions for testing and development

This debugging session successfully resolved the core infrastructure issues blocking the blog/admin system development.
