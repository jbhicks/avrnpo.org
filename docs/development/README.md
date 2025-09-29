# Development Documentation

Development-specific guides, tools, and resources for contributors working on the AVR NPO platform.

## 📋 Development Guides

### 🤖 AI Agents & Automation
- **[Agents Guide](./agents-guide.md)** - Complete guide for AI agents working on this project
- **[Current Feature](./current-feature.md)** - What we're currently working on

### 🔧 Development Planning  
- **[Refactoring Plan](./refactoring-plan.md)** - Strategic refactoring roadmap
- **[CSS Debug](./css-debug.md)** - CSS debugging notes and troubleshooting
- **[CSS Consolidation Progress](./css-consolidation-progress.md)** - CSS consolidation status
- **[Documentation Status](./documentation-status-summary.md)** - Documentation organization status
- **[Donation System Roadmap](./donation-system-roadmap.md)** - Payment system development plan

## 🚀 Quick Start for Contributors

1. **Read the [Agents Guide](./agents-guide.md)** - Critical rules and patterns
2. **Check [Current Feature](./current-feature.md)** - Understand current development focus
3. **Review [Getting Started](../getting-started/README.md)** - Environment setup
4. **Follow [Development Workflow](../getting-started/development-workflow.md)** - Daily commands

## 🔗 Related Documentation

- **[Buffalo Framework](../buffalo-framework/README.md)** - Framework-specific development guides
- **[Testing Guide](../getting-started/testing-guide.md)** - How to test your changes
- **[Frontend Development](../frontend/README.md)** - UI/UX development patterns

## ⚠️ Critical Developer Rules

**Security:**
- NEVER expose API keys or secrets in documentation
- Always use environment variables for sensitive configuration
- Review [Security Guidelines](../deployment/security.md)

**Process Management:**
- NEVER start/kill long-running processes without explicit request
- Buffalo auto-reloads on file changes - no manual restarts needed
- Follow patterns in [Agents Guide](./agents-guide.md)

**Code Style:**
- Follow existing patterns in the codebase
- Use Buffalo conventions for routing and templates
- See [Templates Guide](../buffalo-framework/templates.md) for Plush patterns