# Setup & Configuration

Setup guides and configuration documentation for the AVR NPO platform.

## 🚀 Setup Guides

### 💳 Payment System Setup
- **[Receipt Setup](./receipt-setup.md)** - Configure email receipts and payment confirmations
- **[Admin Setup](./admin-setup.md)** - Administrator account creation and configuration

### 📧 Development Tools  
- **[Dev Email Tests](./dev-email-tests.md)** - Email testing in development environment

## 🔧 Configuration Files

### Environment Configuration
- **`.env`** - Environment variables (not committed to git)
- **`.env.example`** - Template for required environment variables
- **`pb_data/`** - PocketBase data directory (SQLite database)

### Key Environment Variables

```bash
# PocketBase Admin
PB_ADMIN_EMAIL=admin@avrnpo.org
PB_ADMIN_PASSWORD=secure-password-here

# Payment Processing (Helcim)
HELCIM_API_TOKEN=your_api_token_here
HELCIM_TEST_MODE=true  # Set to false in production

# Email Service (SMTP)
EMAIL_ENABLED=true
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
FROM_EMAIL=noreply@avrnpo.org
FROM_NAME=American Veterans Rebuilding
CONTACT_EMAIL=info@avrnpo.org
```

## 📂 Data Storage

### PocketBase Data Directory (`pb_data/`)

Contains all application data:
- **`data.db`** - SQLite database (users, posts, donations)
- **`auxiliary.db`** - PocketBase system data
- **`logs.db`** - Application logs
- **`backups/`** - Automatic database backups

**Important**: This directory must be persisted in production deployments.

## 🔗 Related Documentation

- **[Getting Started](../getting-started/README.md)** - Initial environment setup
- **[Payment System](../payment-system/README.md)** - Payment integration details
- **[Deployment](../deployment/README.md)** - Production configuration
- **[Development Guide](../DEVELOPMENT_GUIDE.md)** - PocketBase development patterns

## ⚠️ Security Notes

- **NEVER commit `.env` files** to version control
- **NEVER commit real API keys** to the repository
- Use placeholder values in documentation examples
- Store ALL sensitive configuration in environment variables only
- Use strong passwords for admin accounts (16+ characters)
- Review [Security Guidelines](../deployment/security.md) before production setup
- Ensure `pb_data/` directory has proper permissions (readable/writable by app only)

## 🧪 Development Setup Checklist

- [ ] Copy `.env.example` to `.env`
- [ ] Set `PB_ADMIN_EMAIL` and `PB_ADMIN_PASSWORD`
- [ ] Configure SMTP settings for email testing
- [ ] Set Helcim API token (use test mode)
- [ ] Run `make dev` to start development server
- [ ] Access PocketBase admin at `http://localhost:8090/_/`
- [ ] Verify admin user created successfully
- [ ] Test email sending with contact form
- [ ] Test donation flow with Helcim test cards

## 🚀 Production Setup Checklist

- [ ] Set all environment variables in deployment platform
- [ ] Change `HELCIM_TEST_MODE` to `false`
- [ ] Use strong, unique `PB_ADMIN_PASSWORD`
- [ ] Configure production SMTP credentials
- [ ] Set up persistent volume for `pb_data/`
- [ ] Enable HTTPS/SSL certificate
- [ ] Configure regular database backups
- [ ] Test email delivery in production
- [ ] Test payment processing with real cards
- [ ] Monitor application logs for errors
