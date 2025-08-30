# Current Features and Roadmap

This document tracks active features and planned improvements for the project, prioritized by urgency and impact.

## Table of Contents

1. [Progressive Enhancement Refactor (Current Feature)](#progressive-enhancement-refactor-current-feature)
2. [Harden Tests](#feature-harden-tests--fix-404s-and-webhook-signature-flakiness)
3. [Template Codebase Deficiencies and Fixes](#feature-template-codebase-deficiencies-and-fixes)

---

# Progressive Enhancement Refactor (Current Feature)

This document describes the comprehensive refactor plan to convert the site to a progressively enhanced application using out-of-the-box Buffalo features and HTMX best practices. It includes an actionable checklist to track progress.

## Goal

Refactor the codebase to follow Buffalo OOTB patterns for forms, rendering, CSRF, flash messaging, and asset handling while strictly following HTMX guidelines (single-template architecture, hx-boost for navigation, progressive enhancement).

---

## Summary of Changes

- Remove header-based conditional rendering (no HX-Request header branching)
- Enable global `hx-boost` in `templates/application.plush.html`
- Consolidate to a single render engine and remove the `htmx.plush.html` layout
- Standardize form handlers to the single-route GET/POST pattern
- Ensure CSRF tokens are included via hidden inputs and rely on Buffalo CSRF middleware
- Use Buffalo flash messages for alerts
- Use Buffalo asset helpers (stylesheetTag/javascriptTag) and fix asset pipeline integration
- Remove form submissions to API endpoints for user-facing pages

---

## Implementation Steps (High Level)

1. Core architecture
   - Remove `renderForRequest()` and `IsHTMX()` helpers and `rHTMX` render engine
   - Remove `templates/htmx.plush.html`
   - Replace all calls to `renderForRequest()` with standard `c.Render(..., r.HTML(...))`
   - Add `hx-boost="true"` to `<body>` in `templates/application.plush.html`

2. Form handlers
   - Update all form handlers to implement GET (show) and POST (process) in the same function
   - Ensure success and error handling uses Buffalo flash messages and standard redirects for non-HTMX
   - Ensure HTMX-enhanced requests still return full pages

3. Templates
   - Ensure all page templates are full-page templates with proper `<html>`, `<head>`, and `<body>` via `application.plush.html`
   - Ensure form partials include hidden `authenticity_token` inputs
   - Remove HTMX-only partials where unnecessary

4. Assets
   - Replace direct asset links with Buffalo helpers (stylesheetTag/javascriptTag)
   - Ensure manifest and asset pipeline are configured

5. Routes
   - Consolidate routes for forms (same path for GET and POST)
   - Remove user-facing forms posting to `/api/*` endpoints

6. Tests & QA
   - Test with JS disabled and enabled
   - Verify bookmarking, refresh, back/forward behavior
   - Add/adjust unit and integration tests where necessary

7. Documentation
   - Update docs in `docs/buffalo-framework/` to reference the new patterns
   - Add this CURRENT_FEATURE.md as the active progress tracker

---

## File Changes Required (non-exhaustive)

- actions/render.go: remove HTMX helpers and rHTMX engine
- templates/htmx.plush.html: remove file
- templates/application.plush.html: add hx-boost on body and replace direct assets with helpers
- Multiple handlers: replace renderForRequest(...) with c.Render(...)
- templates/pages/* and templates/pages/_donate_form.plush.html: ensure CSRF hidden inputs and progressive enhancement attributes
- actions/pages.go, actions/donations.go: consolidate GET/POST handlers

---

## Checklist (Track Progress)

- [ ] Core: Remove `renderForRequest()` and `IsHTMX()` helpers
- [ ] Core: Remove `rHTMX` render engine and `templates/htmx.plush.html`
- [ ] Core: Replace all `renderForRequest()` usages
- [ ] Core: Add `hx-boost="true"` to `templates/application.plush.html`
- [ ] Forms: Convert all form handlers to GET/POST single-handler pattern
- [ ] Forms: Ensure all form templates include hidden `authenticity_token` inputs
- [ ] Alerts: Replace custom alerts with Buffalo `c.Flash()` usage across handlers
- [ ] Assets: Replace direct asset links with `stylesheetTag`/`javascriptTag` helpers
- [ ] Routes: Remove forms posting to `/api/*` and consolidate routes
- [ ] Tests: Add/adjust tests for progressive enhancement and HTMX
- [ ] Docs: Update docs to reflect the new architecture
- [ ] QA: Manual verification with JS disabled/enabled for all major flows

---

## How to Use This Tracker

- Mark checklist items as completed when changes are implemented and tested.
- For each completed item, add a short note below with the commit or PR reference.

---

## Notes & Rationale

We will follow Buffalo OOTB features and HTMX docs: single-template architecture, progressive enhancement (forms with action attributes), and global hx-boost for navigation. Removing header-based rendering prevents bookmark/refresh issues and simplifies templates and handlers.

---

Last updated: (auto-generated)

---

Feature: Harden tests — fix 404s and webhook signature flakiness
Status: in_progress

Description
- Reduce intermittent test failures caused by missing DB fixtures, Helcim webhook signature verification, and missing template partials. Make tests deterministic and CI-friendly.

Short checklist
- Create required DB fixtures for blog post tests (ensure PublishedAt/slugs match handlers).
- Add a reusable test helper: CreatePostForTest (centralizes post creation).
- Ensure CI/test runner prepares DB (migrations) before running DB-dependent tests or clearly mark tests that require DB.
- Make Helcim webhook signature verification test-friendly:
  - Add test helpers to generate valid signatures, and/or
  - Bypass verification only in test environment (GO_ENV=test) or via a guarded env var.
- Add AttachHelcimSignature test helper to simplify signed request creation.
- Scan templates for missing partials; add thin wrapper partials or correct partial references.
- Add a template-prewarm CI test to fail fast on missing partials.
- Update developer docs with testing instructions and helper usage.

Acceptance criteria
- Previously-failing blog slug tests create their fixtures and pass consistently in CI.
- Helcim webhook tests pass using test signatures or bypass mechanism; verification logic is covered by unit tests.
- No missing-partial template render errors during tests; template prewarm passes in CI.
- Docs updated with instructions to run DB-backed tests and webhook-test environment variables.

---

Feature: Template Codebase Deficiencies and Fixes
Status: pending

Description
- Address deficiencies in the Plush template codebase to improve maintainability, security, and consistency. Move logic out of templates into helpers, standardize partial naming, fix error handling, and ensure XSS safety.

Short checklist
- Audit and inventory templates and partial calls (generate a report mapping partial names -> files).
- Create common helpers in Go: fieldClass(errors, name), fieldValue(ctx, name, default), checkedAttr(val, expected), selectedAttr(val, expected), csrfToken(), renderErrors(errors).
- Update handlers to set defaults in context rather than mutating templates.
- Replace inline logic in templates with partial/helper calls incrementally.
- Add/extend tests for helpers and templates; run make test.
- Security review for any helper that returns template.HTML; add sanitizer where needed.
- Lint templates for partial naming consistency and update docs if conventions change.
- Fix specific issues: mutation/defaulting inside templates, large inline logic blocks, error rendering duplication, CSRF handling inconsistency, inline attribute concatenation, raw value insertion risking XSS, inconsistent partial naming, type/append errors for arrays, iterator usage.

Acceptance criteria
- All templates follow Plush best practices with minimal inline logic.
- Common patterns (errors, CSRF, attributes) use standardized helpers/partials.
- No XSS vulnerabilities from unsanitized output.
- Partial naming is consistent and follows repository conventions.
- Unit tests cover new helpers and template rendering edge cases.
- Template prewarm and linting pass without errors.

---
