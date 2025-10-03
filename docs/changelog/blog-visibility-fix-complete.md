# Blog Post Visibility - Complete Fix Summary

## Date: October 2, 2025

## Issues Found and Fixed

### 1. ✅ Missing `published_at` Timestamp in Create Handler
**Problem:** When creating posts via `/admin/posts`, the `published_at` field was not being set, even when the post was marked as published.

**Fix:** Updated `AdminPostsCreate` in `actions/admin.go` to set both `published = true` AND `published_at = now()` when publishing.

```go
if shouldPublish {
    now := time.Now()
    post.PublishedAt = &now
} else {
    post.PublishedAt = nil
}
```

### 2. ✅ Missing `published_at` Timestamp in Update Handler
**Problem:** When updating posts to published status, the `published_at` field wasn't being set.

**Fix:** Updated `AdminPostsUpdate` in `actions/admin.go` to automatically set `published_at` when a post becomes published.

```go
if post.Published && post.PublishedAt == nil {
    now := time.Now()
    post.PublishedAt = &now
} else if !post.Published {
    post.PublishedAt = nil
}
```

### 3. ✅ Incomplete Bulk Operations
**Problem:** Bulk publish/unpublish operations only set one field, leaving data inconsistent.

**Fix:** Updated `AdminPostsBulk` in `actions/admin.go`:

**Publish:**
```sql
UPDATE posts SET published = true, published_at = ? WHERE id IN (?)
```

**Unpublish:**
```sql
UPDATE posts SET published = false, published_at = NULL WHERE id IN (?)
```

### 4. ✅ Incorrect Plush Template Syntax (ROOT CAUSE)
**Problem:** The blog index template was using `<% if` and `<% for` instead of `<%= if` and `<%= for`.

**The Critical Issue:** In Plush templates:
- `<%` = Execute code but don't output anything
- `<%= ` = Execute code AND output the result

For control flow statements (`if`, `for`), you MUST use `<%= ` to render the content blocks.

**Fix:** Changed `templates/blog/index.plush.html`:
```erb
<!-- BEFORE (WRONG) -->
<% if (len(posts) > 0) { %>
  <% for (post) in posts { %>

<!-- AFTER (CORRECT) -->
<%= if (len(posts) > 0) { %>
  <%= for (post) in posts { %>
```

## Files Modified

1. `/home/josh/avrnpo.org/actions/admin.go`
   - `AdminPostsCreate` - Sets `published_at` when publishing
   - `AdminPostsUpdate` - Auto-sets `published_at` on status change
   - `AdminPostsBulk` - Updates both fields in bulk operations

2. `/home/josh/avrnpo.org/templates/blog/index.plush.html`
   - Fixed Plush syntax: `<% if` → `<%= if`
   - Fixed Plush syntax: `<% for` → `<%= for`

3. `/home/josh/avrnpo.org/grifts/`
   - Added helper tasks: `posts:list`, `posts:publish:all`, `db:fix:published`

## Data Integrity Rules

Going forward, these rules are enforced:

| State | `published` | `published_at` |
|-------|-------------|----------------|
| Published | `true` | Current timestamp |
| Draft | `false` | `NULL` |

Both fields must always be in sync.

## Testing the Fix

1. Create a new post at `/admin/posts/new`
2. Click "Publish Post" button
3. Visit `/blog` - post should appear immediately
4. Check database - both `published = true` AND `published_at` should be set

## Key Learnings

### Plush Template Syntax
- **Always use `<%= ` for control flow** (`if`, `for`, `each`)
- Use `<% ` only for assignments and helper calls that don't output
- The admin templates had the same bug but appeared to work due to different conditions

### Data Consistency
- When you have related fields (like `published` and `published_at`), always update them together
- Bulk operations must maintain the same data integrity rules as individual operations

## Migration for Existing Data

If you have existing posts with inconsistent data:

```bash
# List all posts
buffalo task posts:list

# Publish all unpublished posts
buffalo task posts:publish:all

# Fix published posts missing timestamps
buffalo task db:fix:published
```

Or run SQL directly:
```sql
UPDATE posts 
SET published_at = created_at 
WHERE published = true AND published_at IS NULL;
```
