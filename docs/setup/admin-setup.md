# Admin User Creation Guide

PocketBase automatically creates an admin user on first run using environment variables. This guide explains the setup process.

## Automatic Admin Creation

The application automatically creates an admin user on first startup if the `users` collection is empty.

### Required Environment Variables

Set these before starting the application:

```bash
export PB_ADMIN_EMAIL="admin@avrnpo.org"
export PB_ADMIN_PASSWORD="your-secure-password"
```

### Optional Variables

```bash
export PB_ADMIN_FIRST_NAME="Admin"    # Default: "Admin"
export PB_ADMIN_LAST_NAME="User"      # Default: "User"
```

## Setup Methods

### 1. Local Development (Recommended)

Create a `.env` file in the project root:

```bash
# Admin User Configuration
PB_ADMIN_EMAIL=admin@avrnpo.org
PB_ADMIN_PASSWORD=your-secure-password
PB_ADMIN_FIRST_NAME=Admin
PB_ADMIN_LAST_NAME=User
```

Then start the server:
```bash
make dev
```

The admin user will be created automatically on first run.

### 2. Production Deployment

For Coolify or other cloud platforms:

1. **Set environment variables** in your deployment platform:
   - `PB_ADMIN_EMAIL`
   - `PB_ADMIN_PASSWORD`
   - `PB_ADMIN_FIRST_NAME` (optional)
   - `PB_ADMIN_LAST_NAME` (optional)

2. **Deploy application** - Admin user created on first startup

3. **Verify creation** by accessing the admin UI at `https://your-domain.com/_/`

### 3. Manual Server Setup

SSH into your server and set environment variables:

```bash
export PB_ADMIN_EMAIL="admin@avrnpo.org"
export PB_ADMIN_PASSWORD="secure-password-here"
./avrnpo serve --http=0.0.0.0:8090
```

## Accessing the Admin Interface

### PocketBase Admin UI

PocketBase provides a built-in admin dashboard:

1. **Navigate to**: `http://localhost:8090/_/` (development) or `https://your-domain.com/_/` (production)
2. **Login** with your admin credentials
3. **Manage**:
   - Collections (posts, users, etc.)
   - Records (blog posts, user accounts)
   - Files and uploads
   - API rules and permissions

### Application Admin Features

The application also has custom admin features:

1. **Navigate to**: `/login`
2. **Login** with admin credentials
3. **Access**: Admin-only routes (if role = "admin")

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PB_ADMIN_EMAIL` | Yes | none | Admin user email address |
| `PB_ADMIN_PASSWORD` | Yes | none | Admin user password |
| `PB_ADMIN_FIRST_NAME` | No | `Admin` | Admin first name |
| `PB_ADMIN_LAST_NAME` | No | `User` | Admin last name |

## Security Considerations

1. **Use strong passwords** (16+ characters, mixed case, numbers, symbols)
2. **Never commit `.env` files** to version control
3. **Change default passwords** immediately in production
4. **Use HTTPS** for all admin access
5. **Rotate passwords regularly**
6. **Limit admin access** to trusted IP addresses if possible
7. **Enable 2FA** in production (if available)

## Verification

### Verify Admin User Created

Check the logs after first startup:

```bash
# Look for admin creation message
tail -f /tmp/avrnpo.log | grep -i "admin"
```

You should see a message like:
```
[INFO] Admin user created successfully: admin@avrnpo.org
```

### Verify Admin Access

1. **PocketBase Admin UI**: Visit `/_/` and login
2. **Application Admin**: Login at `/login` and verify admin features visible
3. **Database**: Check the `users` collection for admin user with `role = "admin"`

## Troubleshooting

### Admin User Not Created

**Symptom**: No admin user exists after startup

**Solutions**:
1. Check environment variables are set: `echo $PB_ADMIN_EMAIL`
2. Check logs for errors: `tail -f /tmp/avrnpo.log`
3. Verify `pb_data/` directory exists and is writable
4. Ensure `users` collection exists (created by migrations)

### Can't Login to PocketBase Admin

**Symptom**: Login fails at `/_/`

**Solutions**:
1. Verify correct email/password
2. Check PocketBase version is correct (`v0.22+`)
3. Clear browser cache
4. Try incognito/private mode
5. Check server logs for authentication errors

### Database Already Has Users

If users already exist, the auto-creation is skipped. To manually create an admin:

1. **Using PocketBase Admin UI** (if you have admin access):
   - Navigate to `/_/collections/users`
   - Create or edit user
   - Set `role` field to `"admin"`

2. **Using SQLite directly**:
   ```bash
   sqlite3 pb_data/data.db "UPDATE users SET role = 'admin' WHERE email = 'your@email.com';"
   ```

### Permission Errors

**Symptom**: Can't write to database

**Solutions**:
1. Check `pb_data/` directory permissions: `ls -la pb_data/`
2. Ensure application user has write access
3. Check disk space: `df -h`

## Docker/Container Deployments

Add environment variables to your Docker configuration:

**Dockerfile:**
```dockerfile
ENV PB_ADMIN_EMAIL=admin@avrnpo.org
ENV PB_ADMIN_PASSWORD=change-me-in-production
```

**docker-compose.yml:**
```yaml
services:
  avrnpo:
    image: your-image
    environment:
      - PB_ADMIN_EMAIL=admin@avrnpo.org
      - PB_ADMIN_PASSWORD=${ADMIN_PASSWORD}
    volumes:
      - pb_data:/app/pb_data
```

## Coolify Deployment

1. **Navigate to** your app in Coolify
2. **Go to** "Environment Variables"
3. **Add variables**:
   ```
   PB_ADMIN_EMAIL=admin@avrnpo.org
   PB_ADMIN_PASSWORD=your-secure-password
   ```
4. **Save and redeploy**

## Manual Admin Creation (Advanced)

If you need to manually create an admin user via code:

```go
// In a migration or custom script
collection, _ := app.Dao().FindCollectionByNameOrId("users")
record := models.NewRecord(collection)

record.Set("email", "admin@avrnpo.org")
record.Set("password", "secure-password")
record.Set("first_name", "Admin")
record.Set("last_name", "User")
record.Set("role", "admin")
record.Set("verified", true)

if err := app.Dao().SaveRecord(record); err != nil {
    log.Println("Failed to create admin:", err)
}
```

## Next Steps

After creating your admin user:

1. **Login to PocketBase Admin**: `/_/` to manage collections
2. **Login to Application**: `/login` to test admin features
3. **Change password**: Update to a unique, strong password
4. **Create content**: Add blog posts via admin UI
5. **Test features**: Verify donation system, contact forms, etc.

## Additional Security

### Rate Limiting

Admin login is rate-limited to prevent brute force attacks:
- 5 attempts per minute per IP
- Configured in `middleware/ratelimit.go`

### CSRF Protection

All admin forms include CSRF protection:
- Tokens automatically included in forms
- Validated on submission
- Configured in `middleware/csrf.go`

### Session Security

- Sessions expire after inactivity
- Secure cookie flags enabled in production
- HTTP-only cookies to prevent XSS

## Support

If you encounter issues:

1. Check application logs: `tail -f /tmp/avrnpo.log`
2. Check PocketBase logs in Coolify dashboard
3. Verify environment variables are set correctly
4. Review this guide for common solutions
