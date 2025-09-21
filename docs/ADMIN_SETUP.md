# Admin User Creation Guide

When deploying to an empty database, you need to create an initial admin user. This guide provides several methods to do this.

## Methods

### 1. Environment Variables (Recommended for Production)

Set these environment variables before deployment:

```bash
export ADMIN_EMAIL="your-email@avrnpo.org"
export ADMIN_PASSWORD="your-secure-password"
export ADMIN_FIRST_NAME="Your"
export ADMIN_LAST_NAME="Name"
```

Then run:
```bash
buffalo task db:create_admin
```

**Security Note:** Use a strong password and change the default if using it.

### 2. Deployment Script

Use the provided script that handles both environment variables and interactive mode:

```bash
./scripts/create-admin.sh
```

This script will:
- Check for environment variables first
- Fall back to interactive mode if needed
- Provide helpful guidance

### 3. Interactive Mode

For manual setup or development:

```bash
buffalo task db:create_admin_interactive
```

This will prompt you for:
- Email address
- Password  
- First name
- Last name

### 4. Promote First User

If you want users to sign up normally and then promote the first one:

```bash
buffalo task db:promote_admin
```

This finds the first user (by creation date) and promotes them to admin.

## Deployment Integration

### Docker/Container Deployments

Add to your Dockerfile or docker-compose.yml:

```dockerfile
# Set admin user details
ENV ADMIN_EMAIL=admin@avrnpo.org
ENV ADMIN_PASSWORD=change-me-in-production
ENV ADMIN_FIRST_NAME=Admin
ENV ADMIN_LAST_NAME=User

# Run migrations and create admin user
RUN buffalo pop migrate && buffalo task db:create_admin
```

### Coolify/Cloud Deployments

1. Set environment variables in your deployment platform:
   - `ADMIN_EMAIL`
   - `ADMIN_PASSWORD` 
   - `ADMIN_FIRST_NAME` (optional)
   - `ADMIN_LAST_NAME` (optional)

2. Add to your deployment script:
   ```bash
   buffalo pop migrate
   buffalo task db:create_admin
   ```

### Manual Server Setup

1. SSH into your server
2. Navigate to your application directory
3. Run the setup script:
   ```bash
   ./scripts/create-admin.sh
   ```

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ADMIN_EMAIL` | Yes | `admin@avrnpo.org` | Admin user email |
| `ADMIN_PASSWORD` | Yes | `admin123!` | Admin user password |
| `ADMIN_FIRST_NAME` | No | `Admin` | Admin user first name |
| `ADMIN_LAST_NAME` | No | `User` | Admin user last name |

## Security Considerations

1. **Never use default passwords in production**
2. **Use strong passwords** (12+ characters, mixed case, numbers, symbols)
3. **Change admin password** after first login
4. **Rotate passwords regularly**
5. **Use environment variables** to avoid hardcoding credentials
6. **Verify SSL/HTTPS** is enabled for admin login

## Troubleshooting

### "User already exists" error
If you get this error, the admin user already exists. You can:
- Use `buffalo task db:promote_admin` to ensure they have admin role
- Or manually promote them in the database

### Database connection issues
Ensure:
- Database is running and accessible
- Database credentials are correct in `database.yml`
- Migrations have been run (`buffalo pop migrate`)

### Permission errors
Make sure:
- Script is executable (`chmod +x scripts/create-admin.sh`)
- You have write access to the database
- Environment variables are properly set

## Example Commands

```bash
# Production deployment with environment variables
export ADMIN_EMAIL="admin@avrnpo.org"
export ADMIN_PASSWORD="super-secure-password-123!"
export ADMIN_FIRST_NAME="System"
export ADMIN_LAST_NAME="Administrator"
buffalo pop migrate
buffalo task db:create_admin

# Development setup - interactive
buffalo task db:create_admin_interactive

# Quick promotion of first user
buffalo task db:promote_admin
```

## Next Steps

After creating your admin user:

1. **Log in** at `/auth/new` with your admin credentials
2. **Verify admin access** - you should see "Admin" link in navigation
3. **Change password** if using defaults
4. **Test admin functionality** - access `/admin` dashboard
5. **Create additional users** through the admin panel if needed