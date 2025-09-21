# Coolify Deployment Guide for AVRNPO

## Prerequisites

1. **Coolify instance running** with Docker support
2. **PostgreSQL database** configured in Coolify
3. **Domain/subdomain** set up for your application

## Step-by-Step Deployment

### 1. Create New Application in Coolify

1. **Log into Coolify dashboard**
2. **Click "Add New Resource" → "Application"**
3. **Choose "Git Repository"**
4. **Connect your repository:** `https://github.com/jbhicks/avrnpo.org.git`
5. **Select branch:** `main`

### 2. Configure Environment Variables

In Coolify application settings, add these environment variables:

#### Required Variables
```bash
# Database (auto-configured by Coolify if using their PostgreSQL)
DATABASE_URL=postgresql://username:password@host:5432/database_name

# Application Settings
GO_ENV=production
PORT=3001
SESSION_SECRET=your-very-long-random-session-secret-at-least-64-characters

# Admin User Creation
ADMIN_EMAIL=admin@avrnpo.org
ADMIN_PASSWORD=your-secure-admin-password
ADMIN_FIRST_NAME=Admin
ADMIN_LAST_NAME=User
```

#### Optional Variables
```bash
# Email Configuration (if using email features)
SMTP_HOST=your-smtp-host
SMTP_PORT=587
SMTP_USER=your-smtp-username
SMTP_PASSWORD=your-smtp-password

# SSL/Security (if needed)
FORCE_SSL=true
```

### 3. Database Setup

#### Option A: Use Coolify PostgreSQL Service
1. **Add PostgreSQL service** in Coolify
2. **Link it to your application**
3. **Coolify will auto-configure DATABASE_URL**

#### Option B: External Database
1. **Set DATABASE_URL manually** in environment variables
2. **Ensure database is accessible** from Coolify

### 4. Configure Build & Deployment

#### Build Configuration
- **Framework:** Dockerfile
- **Build Command:** (automatic - uses Dockerfile)
- **Port:** 3001

#### Health Check (Optional)
- **Path:** `/`
- **Port:** 3001
- **Timeout:** 30 seconds

### 5. Deploy

1. **Click "Deploy"** in Coolify
2. **Monitor deployment logs** for any issues
3. **Check for these success messages:**
   ```
   📊 Running database migrations...
   👤 Creating admin user...
   ✅ Admin user setup completed!
   🌐 Starting web server...
   ```

### 6. Post-Deployment

1. **Visit your application URL**
2. **Log in with admin credentials:**
   - Email: The one you set in `ADMIN_EMAIL`
   - Password: The one you set in `ADMIN_PASSWORD`
3. **Verify admin access** - you should see "Admin" in navigation
4. **Change admin password** if using defaults

## Troubleshooting

### Admin User Issues

**Problem:** Admin user not created
```bash
# Check logs in Coolify for error messages
# Manually create admin user via console:
docker exec -it your-container-name /bin/sh
cd /app
./bin/app task db:create_admin
```

**Problem:** "User already exists" error
- This is normal if admin user was already created
- Check if user has admin role in database

### Database Connection Issues

**Problem:** Database connection failed
- Verify `DATABASE_URL` is correct
- Check if database service is running
- Ensure network connectivity between app and database

### Build Issues

**Problem:** Build fails
- Check if all dependencies are available
- Verify Go version compatibility
- Check Dockerfile syntax

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `GO_ENV` | Yes | production | Application environment |
| `SESSION_SECRET` | Yes | - | Secret for session encryption |
| `PORT` | No | 3001 | Port for web server |
| `ADMIN_EMAIL` | Yes | admin@avrnpo.org | Admin user email |
| `ADMIN_PASSWORD` | Yes | - | Admin user password |
| `ADMIN_FIRST_NAME` | No | Admin | Admin user first name |
| `ADMIN_LAST_NAME` | No | User | Admin user last name |

## Security Checklist

- [ ] Strong `SESSION_SECRET` (64+ random characters)
- [ ] Strong `ADMIN_PASSWORD` (12+ characters, mixed case, numbers, symbols)
- [ ] `GO_ENV=production` set
- [ ] Database credentials secure
- [ ] SSL/HTTPS enabled via Coolify
- [ ] Admin password changed after first login

## Monitoring

- **Application logs:** Available in Coolify dashboard
- **Health check:** Coolify monitors application status
- **Database:** Monitor via Coolify PostgreSQL service

## Backup

- **Database:** Use Coolify PostgreSQL backup features
- **Application:** Code is backed up in Git repository