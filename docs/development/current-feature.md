# Current Feature: PocketBase Migration - Phase 3

**Status:** In Progress  
**Date:** October 7, 2025

## Overview

Completing Phase 3 of the PocketBase migration by implementing the UI layer using Templ templates, integrating services with PocketBase, and creating all necessary pages for the AVR website.

## Current Phase: Phase 3 - UI Implementation & Service Integration

### Template Strategy: Templ

Using [Templ](https://templ.guide/) for type-safe, component-based Go templates, consistent with our SaaS template repository approach.

**Benefits:**
- ✅ Type-safe templates compiled to Go code
- ✅ Component-based architecture
- ✅ IDE support with autocomplete
- ✅ No runtime template parsing
- ✅ Easy integration with HTMX
- ✅ Hot reload in development

### Implementation Checklist

#### 1. Setup & Configuration
- [ ] Install Templ CLI tool
- [ ] Configure Templ generation in development workflow
- [ ] Set up hot reload for `.templ` files
- [ ] Configure build process to generate Templ code

#### 2. Base Layout Components
- [ ] Create base HTML layout component
- [ ] Create navigation component (AVR branding)
- [ ] Create footer component
- [ ] Create flash message component
- [ ] Create error display component
- [ ] Create SEO meta tags component

#### 3. Page Templates
- [ ] Home page with hero section and latest posts
- [ ] Blog listing page with pagination/infinite scroll
- [ ] Blog post detail page with SEO metadata
- [ ] Donation page with Helcim.js integration
- [ ] Contact form page
- [ ] About page
- [ ] Success/error pages

#### 4. Service Integration
- [ ] Adapt Helcim service to work with PocketBase
  - [ ] Create/update donation records in `donations` collection
  - [ ] Handle checkout token generation
  - [ ] Process webhook callbacks
  - [ ] Update subscription records
- [ ] Adapt email service to use PocketBase mailer
  - [ ] Donation receipts
  - [ ] Contact form notifications
  - [ ] Admin notifications

#### 5. Route Handlers
- [ ] Update handlers to use PocketBase SDK
- [ ] Implement blog routes (listing, detail)
- [ ] Implement donation routes (form, payment, webhook)
- [ ] Implement contact form route
- [ ] Implement static page routes (about, home)

#### 6. HTMX Integration
- [ ] Add HTMX for infinite scroll on blog
- [ ] Add HTMX for form validation feedback
- [ ] Add HTMX for contact form submission
- [ ] Progressive enhancement patterns

## Previous Phases Status

### ✅ Phase 1: Project Structure & Initialization (Complete)
- Buffalo code archived to `archive/buffalo/`
- PocketBase application initialized
- Basic route structure created
- Services preserved and ready

### ✅ Phase 2: Database Collections & Migrations (Complete)
- All collections created (`posts`, `donations`, `contact_submissions`, `users`)
- Migration system implemented
- Access rules configured
- Database schema matches requirements

## Next Steps After Phase 3

### Phase 4: Data Migration (If Needed)
- Export data from Buffalo/PostgreSQL
- Transform to PocketBase format
- Import via PocketBase API
- Verify data integrity

### Phase 5: Testing & Deployment
- End-to-end testing
- Payment flow testing
- Email delivery testing
- Production deployment to Coolify

## Reference Documentation

- [REFACTORING_STATUS.md](../../REFACTORING_STATUS.md) - Overall refactoring status
- [Templ Documentation](https://templ.guide/)
- [PocketBase Documentation](https://pocketbase.io/docs/)
- [Helcim Integration](../payment-system/helcim-integration.md)

## Notes

- All Buffalo/Plush references are now deprecated
- Focus is on PocketBase + Templ stack
- HTMX patterns from Buffalo project can be adapted
- Pico CSS remains the styling framework
