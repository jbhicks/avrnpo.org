# Blog Post Visibility Fix

## Date: October 2, 2025

## Problem
Newly created blog posts were not appearing at `/blog` even when marked as published.

## Root Causes Identified

### 1. Missing `published_at` Timestamp in AdminPostsCreate
When creating a new post through the admin interface at `/admin/posts`, the code was setting `post.Published = true` but **not setting** `post.PublishedAt`. This left the `published_at` field as NULL in the database.

**Location:** `actions/admin.go` - `AdminPostsCreate` function

### 2. Blog Query Ordering Issue
The blog index page at `/blog` queries posts with:
- `WHERE published = ?` (true) ✅
- `ORDER BY published_at desc` ❌

Since newly created posts had `published = true` but `published_at = NULL`, they matched the WHERE clause but the NULL `published_at` value caused them to be sorted last or not displayed properly.

**Location:** `actions/blog.go` - `BlogIndex` function (line 22)

### 3. Bulk Publish Action Incomplete
The admin bulk publish action was only setting `published_at` without setting the `published` field:
```sql
UPDATE posts SET published_at = ? WHERE id IN (?)
```

This created the opposite problem - posts got a timestamp but remained unpublished.

**Location:** `actions/admin.go` - `AdminPostsBulk` function

## Solutions Implemented

### 1. Fixed AdminPostsCreate Handler
```go
// Handle published status - use form data if provided, otherwise check action
action := c.Param("action")
if action == "publish" {
    post.Published = true
    now := time.Now()
    post.PublishedAt = &now  // ✅ Now sets timestamp
} else if action == "draft" {
    post.Published = false
    post.PublishedAt = nil
}
```

### 2. Fixed AdminPostsUpdate Handler
```go
// Handle published status changes
if post.Published && post.PublishedAt == nil {
    // Post is being published for the first time
    now := time.Now()
    post.PublishedAt = &now
} else if !post.Published {
    // Post is being unpublished
    post.PublishedAt = nil
}
```

### 3. Fixed AdminPostsBulk Publish Action
```sql
UPDATE posts SET published = true, published_at = ? WHERE id IN (?)
```

Now sets **both** the `published` field and the `published_at` timestamp.

### 4. Fixed AdminPostsBulk Unpublish Action
```sql
UPDATE posts SET published = false, published_at = NULL WHERE id IN (?)
```

Now sets **both** fields to ensure consistency.

## Data Integrity Rules

Going forward, the following rules are enforced:

1. **When publishing a post:**
   - `published` = `true`
   - `published_at` = current timestamp

2. **When unpublishing a post:**
   - `published` = `false`
   - `published_at` = `NULL`

3. **Both fields must be in sync** for posts to appear correctly in the blog listing.

## Testing

The fixes ensure that:
- New posts created via admin with "publish" action appear immediately at `/blog`
- Posts edited to published status get a `published_at` timestamp
- Bulk publish/unpublish operations maintain data consistency
- Blog query ordering works correctly with `published_at desc`

## Files Changed

- `actions/admin.go` - AdminPostsCreate, AdminPostsUpdate, AdminPostsBulk functions

## Migration Notes

**Important:** Existing posts in the database may have inconsistent states:
- Posts with `published = true` but `published_at = NULL` won't display properly
- These should be fixed by either:
  1. Re-publishing them through the admin interface
  2. Running a database migration to set `published_at` for already-published posts

Example migration query:
```sql
UPDATE posts 
SET published_at = COALESCE(published_at, created_at) 
WHERE published = true AND published_at IS NULL;
```
