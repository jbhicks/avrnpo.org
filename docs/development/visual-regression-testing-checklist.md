# Visual Regression Testing Checklist - Pico CSS Refactoring

**Purpose**: Verify all refactored templates render correctly after removing inline styles and Tailwind classes.

## Testing Setup

1. Start development server: `make dev` or `./dev.sh`
2. Open browser to `http://127.0.0.1:8090`
3. Test in both light and dark modes
4. Test responsive breakpoints: mobile (375px), tablet (768px), desktop (1280px)

## Pages to Test

### ✓ Public Pages

#### Home Page (`/`)
- [ ] Mission section displays with proper card styling
- [ ] Blog section shows latest posts with proper grid layout
- [ ] "The Process" section shows 3-column grid on desktop
- [ ] Donation form displays with proper card hierarchy
- [ ] Impact items show correct styling
- [ ] Amount buttons in 3x2 grid
- [ ] Social links display correctly
- [ ] All cards have proper borders and shadows
- [ ] Color hierarchy: concrete gray body → white cards → light gray sections

#### About Page (`/about`)
- [ ] Content sections render with proper spacing
- [ ] Grid layouts work correctly
- [ ] Typography is consistent

#### Team Page (`/team`)
- [ ] Team intro section centered, max-width 800px
- [ ] Team member cards in 2-column grid
- [ ] Team member images 300px height, proper object-fit
- [ ] Overlay gradient on images works
- [ ] Team member titles in army gold color
- [ ] Card content padding correct
- [ ] Grid spacing (2rem margin-bottom) between rows

#### Projects Page (`/projects`)
- [ ] "The Process" section displays correctly
- [ ] Process cards in 3-column grid
- [ ] Process numbers styled properly
- [ ] Process lists render correctly

#### Blog Page (`/blog`)
- [ ] Blog header styled correctly
- [ ] Empty state shows when no posts
- [ ] Blog post list items display properly
- [ ] Post metadata (date) shows as muted text
- [ ] Post detail page layout correct
- [ ] Post header, content, footer sections styled

#### Contact Page (`/contact`)
- [ ] Contact form displays in grid layout
- [ ] Social links grid renders correctly
- [ ] Info sections styled properly
- [ ] Success message displays correctly
- [ ] Contact result page uses correct styling

#### Donate Page (`/donate`)
- [ ] Donation intro centered
- [ ] Impact grid (2 columns) displays correctly
- [ ] Donation form card styled properly
- [ ] Amount grid (3x2 on desktop, responsive) works
- [ ] Donor info section renders correctly
- [ ] Address info section displays
- [ ] Tax info and disclaimers styled
- [ ] Success page displays correctly
- [ ] Error page displays correctly

### ✓ Admin Pages (Requires Login)

#### Login Page (`/admin/login`)
- [ ] Login form centered and styled
- [ ] Error messages display correctly
- [ ] Full-width button works
- [ ] Minimal, clean design

#### Admin Posts (`/admin/posts`)
- [ ] Admin container max-width correct
- [ ] Admin card styling applied
- [ ] Admin header with actions displays
- [ ] Posts table renders correctly
- [ ] Table responsive wrapper works on mobile
- [ ] Status badges (published/draft) styled correctly
- [ ] Post actions buttons work
- [ ] Delete action has proper red styling
- [ ] Empty state displays when no posts
- [ ] Hover states on table rows work

#### Post Form (`/admin/posts/create`, `/admin/posts/:id/edit`)
- [ ] Admin nav breadcrumb displays
- [ ] Form container narrower than table view
- [ ] Form fields styled correctly
- [ ] Quill editor integration works
- [ ] Form actions (Save/Cancel) display properly
- [ ] Small button modifier works
- [ ] Published toggle renders correctly

## Color System Testing

### Light Mode
- [ ] Body background: `#e8e9ea` (concrete gray)
- [ ] Card backgrounds: `#ffffff` (white)
- [ ] Section backgrounds: `#f5f5f5` (light gray)
- [ ] Card borders visible with subtle color
- [ ] Card shadows applied (2-level depth)
- [ ] Army green accents: `#4a5d23`
- [ ] Army gold primary: `#ffb627`

### Dark Mode
- [ ] Body background: `#0d1117` (deep dark)
- [ ] Card backgrounds: medium dark (from Pico variables)
- [ ] Section backgrounds: darker than cards
- [ ] Card borders maintain hierarchy
- [ ] Card shadows work in dark mode
- [ ] Army green accents visible
- [ ] Army gold stands out

## Responsive Breakpoints

### Mobile (< 768px)
- [ ] Grid layouts stack to single column
- [ ] Navigation collapses appropriately
- [ ] Donation amount buttons stack properly
- [ ] Admin tables scroll horizontally
- [ ] Card padding adjusts
- [ ] Typography scales down

### Tablet (768px - 1024px)
- [ ] Grid layouts show 2 columns where appropriate
- [ ] Spacing adjusts appropriately
- [ ] Admin interface usable

### Desktop (> 1024px)
- [ ] Full 3-column grids display
- [ ] Optimal spacing and padding
- [ ] Admin interface comfortable

## WCAG Accessibility

- [ ] All text meets WCAG AA contrast ratios (4.5:1 for normal text)
- [ ] Headings meet AAA contrast (7:1)
- [ ] Focus states visible on all interactive elements
- [ ] Color not sole indicator of information

## Pico CSS Compatibility

- [ ] All styling uses Pico CSS variables
- [ ] No CSS conflicts with Pico base styles
- [ ] Semantic HTML elements used correctly
- [ ] Button roles work correctly
- [ ] Form validation states work
- [ ] Grid system works with Pico

## Browser Testing

- [ ] Chrome/Chromium
- [ ] Firefox
- [ ] Safari (if available)
- [ ] Mobile browsers (Chrome Mobile, Safari iOS)

## Known Issues to Watch For

1. Card border visibility in dark mode
2. Admin table overflow on small screens
3. Quill editor z-index with Pico modals
4. Team member image overlay readability
5. Donation amount button active states

## Test Results

**Date Tested**: ___________  
**Tested By**: ___________  
**Browser/Version**: ___________  
**Issues Found**: ___________

## Notes

- All inline `style=""` attributes removed
- All Tailwind utility classes removed
- 100% semantic/component-based CSS
- Enhanced 3-tier color system implemented
- Admin panel fully refactored
