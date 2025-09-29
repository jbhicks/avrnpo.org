# Setup & Configuration

Setup guides and configuration documentation for the AVR NPO platform.

## 🚀 Setup Guides

### 💳 Payment System Setup
- **[Receipt Setup](./receipt-setup.md)** - Configure email receipts and payment confirmations
- **[Admin Setup](./admin-setup.md)** - Administrator account and role configuration

### 📧 Development Tools  
- **[Dev Email Tests](./dev-email-tests.md)** - Email testing in development environment

## 🔧 Configuration Files

### Environment Configuration
- **`.env`** - Development environment variables
- **`database.yml`** - Database connection settings  
- **`config/buffalo-app.toml`** - Buffalo application configuration

### Key Environment Variables
```bash
# Payment Processing
HELCIM_API_TOKEN=your_api_token_here
HELCIM_COMMERCE_ID=your_commerce_id

# Email Service  
MAILGUN_API_KEY=your_mailgun_key
MAILGUN_DOMAIN=your_domain

# Database
DATABASE_URL=postgres://user:pass@localhost/dbname
```

## 🔗 Related Documentation

- **[Getting Started](../getting-started/README.md)** - Initial environment setup
- **[Payment System](../payment-system/README.md)** - Payment integration details
- **[Deployment](../deployment/README.md)** - Production configuration

## ⚠️ Security Notes

- **NEVER commit real API keys** to version control
- Use placeholder values in documentation examples
- Store sensitive configuration in environment variables only
- Review [Security Guidelines](../deployment/security.md) before production setup