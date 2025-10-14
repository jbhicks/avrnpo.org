# Archived Documentation

This document contains links to archived documentation from the Buffalo framework era of this project. These docs are preserved for historical reference but are **no longer applicable** to the current PocketBase implementation.

**Note:** The project was refactored from Buffalo to PocketBase in October 2025. See [REFACTORING_STATUS.md](../REFACTORING_STATUS.md) for current status.

## Buffalo-Era Documentation (Archived)

### Features & Roadmaps
- [Buffalo Donation System Implementation](./changelog/buffalo-donation-system-archived.md) - Original Buffalo-based donation system plans
- [Buffalo Progressive Enhancement Refactor](./changelog/buffalo-progressive-enhancement-archived.md) - Buffalo/HTMX/Plush refactoring plans
- [Buffalo Test Hardening](./changelog/buffalo-test-hardening-archived.md) - Buffalo test infrastructure improvements
- [Buffalo Template Deficiencies](./changelog/buffalo-template-fixes-archived.md) - Plush template improvements

### Buffalo Framework Documentation
- [Buffalo Authentication](./buffalo-framework/authentication.md) - Buffalo auth patterns (archived)
- [Buffalo CSRF/HTMX Implementation](./buffalo-framework/csrf-htmx-implementation.md) - Buffalo CSRF patterns (archived)
- [Buffalo Database](./buffalo-framework/database.md) - Pop/Fizz database patterns (archived)
- [Buffalo Form URL Fix](./buffalo-framework/form-url-fix-implementation.md) - Buffalo form handling (archived)
- [Buffalo Templates](./buffalo-framework/templates.md) - Plush template documentation (archived)
- [Buffalo Troubleshooting](./buffalo-framework/troubleshooting.md) - Buffalo-specific issues (archived)

### Changelog (Buffalo Era)
- [Blog Fix Status](./changelog/blog-fix-current-status.md)
- [Blog Visibility Fix](./changelog/blog-visibility-fix-complete.md)
- [CSRF Implementation](./changelog/csrf-implementation.md)
- [Helcim Standardization](./changelog/helcim-standardization.md)

## Current Active Documentation

For current documentation, see:
- [README.md](./README.md) - Main documentation index
- [REFACTORING_STATUS.md](../REFACTORING_STATUS.md) - Current refactoring status
- [Development Guide](./DEVELOPMENT_GUIDE.md) - Current development workflows
- [Current Feature](./development/current-feature.md) - Active feature work

## Why Archive?

The Buffalo framework codebase has been moved to `archive/buffalo/` and is no longer the active codebase. All documentation referencing Buffalo, Plush templates, Pop/Fizz migrations, and Buffalo-specific patterns is now historical reference only.

The new stack:
- **PocketBase** (replaces Buffalo)
- **Templ** (replaces Plush templates)
- **SQLite** (replaces PostgreSQL)
- **Go stdlib + PocketBase SDK** (replaces Buffalo framework)
