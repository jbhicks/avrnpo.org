# Coolify Deployment Guide - PocketBase Migration

> **Migration Guide**: This document covers deploying the new PocketBase/GOTH-based version of AVRNPO to Coolify. The application has been migrated from Buffalo framework to PocketBase.

## Prerequisites

1. **Coolify instance running** with Docker support
2. **Persistent storage** configured in Coolify for PocketBase data
3. **Domain/subdomain** set up for your application
4. **No PostgreSQL needed** - PocketBase uses embedded SQLite

## Architecture Changes from Buffalo

| Component | Buffalo (Old) | PocketBase (New) |
|-----------|---------------|------------------|
| Framework | Buffalo | PocketBase + GOTH |
| Database | PostgreSQL | SQLite (embedded) |
| Migrations | Soda | PocketBase migrations |
| Templates | Plush | Templ (compiled to Go) |
| Admin UI | Custom | PocketBase Admin UI at `/_/` |
| Build | Buffalo CLI | Standard Go build |
| Port | 3001 | 8090 (default) |

## Step-by-Step Deployment

### 1. Update Dockerfile

The existing Dockerfile needs to be replaced to support PocketBase instead of Buffalo:

**New Dockerfile:**
```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o avrnpo .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/avrnpo .

# Copy static assets and templates (if not embedded)
COPY --from=builder /app/pb_public ./pb_public

# Create directory for PocketBase data
RUN mkdir -p pb_data

# Expose PocketBase default port
EXPOSE 8090

# Run the application
CMD ["./avrnpo", "serve", "--http=0.0.0.0:8090"]
```

### 2. Create New Application in Coolify

1. **Log into Coolify dashboard**
2. **Click "Add New Resource" → "Application"**
3. **Choose "Git Repository"**
4. **Connect your repository:** `https://github.com/jbhicks/avrnpo.org.git`
5. **Select branch:** `main`

### 3. Configure Build Settings

**Build Configuration:**
- **Build Pack:** Dockerfile
- **Dockerfile Location:** `./Dockerfile` (root)
- **Port:** 8090
- **Health Check Path:** `/api/health` (or just `/`)

### 4. Configure Persistent Storage

**CRITICAL:** PocketBase stores data in SQLite files that MUST persist across deployments.

**Add Volume Mount in Coolify:**
```
Source: /var/lib/coolify/volumes/avrnpo-pb-data
Destination: /app/pb_data
```

This ensures your database, uploaded files, and settings survive container restarts.

### 5. Configure Environment Variables

Add these environment variables in Coolify application settings:

#### Core Application Settings
```bash
# PocketBase Admin User (auto-created on first run)
PB_ADMIN_EMAIL=admin@avrnpo.org
PB_ADMIN_PASSWORD=your-very-secure-admin-password-here

# CSRF Protection (32 characters minimum)
CSRF_KEY=your_32_character_csrf_key_here_must_be_exactly_32_chars

# Environment
GO_ENV=production
```

#### Payment Processing (Helcim)
```bash
HELCIM_PRIVATE_API_KEY=your_helcim_private_api_key_here
HELCIM_WEBHOOK_VERIFIER_TOKEN=your_webhook_token_here
HELCIM_CURRENCY=USD
HELCIM_TEST_MODE=false
```

#### Email Configuration
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
FROM_EMAIL=donations@avrnpo.org
FROM_NAME=American Veterans Rebuilding
EMAIL_ENABLED=true
CONTACT_EMAIL=AmericanVeteransRebuilding@avrnpo.org
```

#### Organization Information
```bash
ORGANIZATION_EIN=12-3456789
ORGANIZATION_ADDRESS=1234 Main St, Your City, ST 12345
```

#### Security Settings
```bash
# SSL/Security
FORCE_SSL=true

# PocketBase Auto-Migration (optional, use with caution)
PB_AUTOMIGRATE=1
```

### 6. Deploy

1. **Click "Deploy"** in Coolify
2. **Monitor deployment logs** for:
   ```
   ✅ Build successful
   📦 Extracting Go modules
   🔨 Building application binary
   🚀 Starting PocketBase server
   👤 Admin user created/verified
   🌐 Server listening on http://0.0.0.0:8090
   ```

### 7. Post-Deployment Setup

#### A. Verify Application is Running
Visit your application URL - you should see the homepage.

#### B. Access PocketBase Admin UI
1. **Navigate to:** `https://your-domain.com/_/`
2. **Log in with:**
   - Email: Value from `PB_ADMIN_EMAIL`
   - Password: Value from `PB_ADMIN_PASSWORD`

#### C. Verify Admin User in Application
1. **Log in at:** `https://your-domain.com/auth/login`
2. **Use same credentials** as PocketBase admin
3. **Check navigation** - should see "Admin" link
4. **Visit:** `https://your-domain.com/cms/posts`

#### D. Create Initial Blog Posts
1. **Go to PocketBase Admin:** `/_/`
2. **Navigate to "Collections" → "posts"**
3. **Click "New record"**
4. **Fill in:**
   - Title
   - Slug (URL-friendly, e.g., `welcome-to-avrnpo`)
   - Content (supports Markdown)
   - Excerpt
   - Published At (set to current date/time to publish)
   - Featured Image (optional file upload)

## Key Differences for Operations

### No More Database Migrations to Run Manually
- PocketBase handles migrations automatically via Go migration files
- Set `PB_AUTOMIGRATE=1` for automatic migration on startup (production)
- Or run migrations manually: `./avrnpo migrate up`

### Admin User Creation is Automatic
- Admin user created on first startup from environment variables
- No need to run separate admin creation tasks
- User is created in `users` collection with `role = "admin"`

### Static File Uploads
- Files uploaded via PocketBase Admin UI stored in `pb_data/storage/`
- This directory MUST be in persistent volume
- Featured images for blog posts stored here

### Logs
- Application logs go to stdout/stderr (visible in Coolify)
- PocketBase internal logs in `pb_data/logs/`

### Backups
**Critical Files to Backup:**
```
pb_data/
├── data.db          # Main database
├── data.db-shm      # Shared memory file
├── data.db-wal      # Write-ahead log
├── logs/            # Application logs
└── storage/         # Uploaded files
```

**Backup Strategy:**
1. Stop the container
2. Copy entire `pb_data/` directory
3. Restart container

**Or use Coolify's volume backup feature.**

## Testing the Deployment

### 1. Homepage Loads
```bash
curl https://your-domain.com/
```
Should return HTML homepage.

### 2. Admin Login Works
```bash
curl -X POST https://your-domain.com/auth/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "email=admin@avrnpo.org&password=yourpassword"
```
Should redirect or return success.

### 3. PocketBase Admin UI Accessible
Visit `https://your-domain.com/_/` and log in.

### 4. Blog Posts Display
1. Create a post in PocketBase admin
2. Visit `https://your-domain.com/blog`
3. Post should appear

### 5. File Uploads Work
1. Create a blog post with featured image
2. Image should display on blog post page

## Troubleshooting

### Admin User Not Created
**Symptoms:** Can't log in at `/_/` or `/auth/login`

**Check logs for:**
```
Created new admin user: admin@avrnpo.org
```
or
```
Admin user already exists: admin@avrnpo.org
```

**Fix:**
1. Verify `PB_ADMIN_EMAIL` and `PB_ADMIN_PASSWORD` are set
2. Restart container
3. Check that user exists in PocketBase Admin UI under Users collection

### Database/Files Not Persisting
**Symptoms:** Data disappears after deployment/restart

**Fix:**
1. Verify volume mount in Coolify: `/app/pb_data`
2. Check volume permissions
3. Restart container and verify volume is mounted

### Port Issues
**Symptoms:** App not accessible or connection refused

**Fix:**
1. Verify Coolify port mapping: Container port `8090` → Host port
2. Check that application is listening on `0.0.0.0:8090` (not `127.0.0.1`)
3. Check firewall rules

### CSRF Token Errors
**Symptoms:** Form submissions fail with CSRF errors

**Fix:**
1. Verify `CSRF_KEY` is set (exactly 32 characters)
2. Check that forms include `<input type="hidden" name="csrf_token">`
3. Restart application

### Featured Images Not Loading
**Symptoms:** Images uploaded but don't display

**Check:**
1. Files exist in `pb_data/storage/`
2. File permissions are correct
3. PocketBase admin UI can access files
4. Image URL construction in template is correct

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PB_ADMIN_EMAIL` | Yes | - | Admin user email (auto-created) |
| `PB_ADMIN_PASSWORD` | Yes | - | Admin user password |
| `CSRF_KEY` | Yes | - | CSRF protection key (32 chars) |
| `GO_ENV` | No | development | Application environment |
| `PB_AUTOMIGRATE` | No | 0 | Auto-run migrations on startup (1=yes) |
| `HELCIM_PRIVATE_API_KEY` | Yes* | - | Helcim API key (*if using donations) |
| `HELCIM_CURRENCY` | No | USD | Currency for payments |
| `HELCIM_TEST_MODE` | No | false | Use Helcim test mode |
| `EMAIL_ENABLED` | No | false | Enable email sending |
| `SMTP_HOST` | Yes* | - | SMTP server (*if EMAIL_ENABLED=true) |
| `SMTP_PORT` | No | 587 | SMTP port |
| `SMTP_USERNAME` | Yes* | - | SMTP username |
| `SMTP_PASSWORD` | Yes* | - | SMTP password |
| `FROM_EMAIL` | No | - | From address for emails |
| `FROM_NAME` | No | - | From name for emails |
| `CONTACT_EMAIL` | No | - | Contact form destination email |
| `ORGANIZATION_EIN` | No | - | Tax ID for receipts |
| `ORGANIZATION_ADDRESS` | No | - | Org address for receipts |
| `FORCE_SSL` | No | false | Redirect HTTP to HTTPS |

## Rollback Plan

If deployment fails or issues arise:

### Option 1: Revert to Previous Container
1. In Coolify, select previous deployment
2. Click "Redeploy"
3. Verify functionality

### Option 2: Restore from Backup
1. Stop current container
2. Replace `pb_data/` with backup
3. Restart container

### Option 3: Emergency Buffalo Rollback
If you need to rollback to Buffalo version:
1. Create new Coolify app pointing to pre-migration commit
2. Restore PostgreSQL database from backup
3. Update environment variables for Buffalo (see old docs)

## Security Checklist

- [ ] `CSRF_KEY` is random and exactly 32 characters
- [ ] `PB_ADMIN_PASSWORD` is strong (16+ chars, mixed)
- [ ] `FORCE_SSL=true` in production
- [ ] Volume mount for `pb_data` is configured
- [ ] Database backups scheduled
- [ ] `HELCIM_TEST_MODE=false` in production
- [ ] Email credentials secure
- [ ] PocketBase Admin UI (`/_/`) access restricted or monitored
- [ ] SMTP credentials are app-specific passwords, not main passwords

## Migration from Old Buffalo Deployment

If you're migrating from an existing Buffalo deployment:

### Data Migration Required
1. **Export data from PostgreSQL**
   ```sql
   -- Export blog posts
   COPY (SELECT * FROM posts) TO '/tmp/posts.csv' CSV HEADER;
   
   -- Export users
   COPY (SELECT * FROM users) TO '/tmp/users.csv' CSV HEADER;
   ```

2. **Import into PocketBase**
   - Use PocketBase Admin UI at `/_/`
   - Import CSV files into respective collections
   - Or write a migration script in Go

### DNS Update
1. Test new PocketBase deployment on staging domain
2. Verify all functionality works
3. Update DNS to point to new deployment
4. Monitor for issues
5. Keep old deployment running for 24h as fallback

## Monitoring & Maintenance

### Health Checks
- **Endpoint:** `/` or `/api/health`
- **Expected:** 200 OK response
- **Frequency:** Every 30 seconds

### Log Monitoring
**Check for:**
- CSRF errors
- Authentication failures
- Payment processing errors
- Email sending failures

**Logs location:**
- Application logs: Coolify dashboard or `docker logs`
- PocketBase logs: `pb_data/logs/`

### Performance
- **SQLite performance:** Excellent for read-heavy workloads
- **Write-heavy?** Consider read replicas or optimization
- **File storage:** Monitor `pb_data/storage/` size

### Updates
**To update the application:**
1. Push changes to Git repository
2. Coolify auto-deploys (if enabled)
3. Or manually trigger deployment in Coolify
4. Monitor logs for successful startup

## Support & Documentation

- **PocketBase Docs:** https://pocketbase.io/docs/
- **GOTH (Buffalo-like for Go):** https://github.com/go-goth/goth
- **Application Structure:** See `/docs/DEVELOPMENT_GUIDE.md`
- **Local Development:** See `/docs/getting-started/quick-start.md`
- **Testing:** See `/README_TESTING.md`

## Summary

**Key Takeaways:**
1. ✅ No PostgreSQL needed - SQLite embedded
2. ✅ Persistent volume for `pb_data/` is CRITICAL
3. ✅ Admin user auto-created from env vars
4. ✅ Port changed from 3001 → 8090
5. ✅ PocketBase Admin UI at `/_/` for content management
6. ✅ Simpler deployment - just build Go binary and run
