# Technical Reference

API documentation, database schemas, and technical references for the AVR NPO donation system.

## 📋 Reference Documentation

### 🔌 API Documentation
- **[API Endpoints](./api-endpoints.md)** - Complete REST API reference *(planned)*
- **[Authentication](./authentication.md)** - API authentication and authorization *(planned)*

### 🗄️ Database Documentation  
- **[Database Schema](./database-schema.md)** - Current database structure *(planned)*
- **[Migration History](./migration-history.md)** - Database change log *(planned)*

### 📦 Development References
- **[Dependencies](./dependencies.md)** - Dependency management rules and guidelines
- **[Subscription API Reference](./subscription-api-reference.md)** - Complete subscription management API
- **[Changelog](./changelog.md)** - Major changes and version history *(planned)*

## 🎯 Current System Architecture

### Technology Stack
- **Backend**: Go 1.24+ with Buffalo v0.18.14 web framework
- **Database**: PostgreSQL 13+ with Pop ORM
- **Frontend**: Pico CSS + HTMX for progressive enhancement
- **Payment Processing**: Helcim API integration
- **Email**: SMTP for donation receipts and notifications

### Key Dependencies
- **Buffalo**: Web framework and development tools
- **Pop**: Database ORM and migration system  
- **Plush**: Server-side templating engine
- **Helcim**: Payment processing APIs
- **Pico CSS**: Semantic CSS framework
- **HTMX**: Progressive enhancement library

## 🔌 API Overview

### Public Endpoints
- **Donation processing** - Secure payment initialization
- **Webhook handlers** - Payment status updates from Helcim
- **Contact forms** - General inquiries and communication

### Authenticated Endpoints
- **User management** - Account creation and management
- **Subscription management** - View and cancel recurring donations
- **Admin functions** - Donation oversight and reporting

### Payment Integration
- **Helcim APIs** - Payment and Recurring APIs
- **Webhook endpoints** - Real-time payment notifications
- **Receipt system** - Email confirmation delivery

## 🗄️ Database Schema Overview

### Core Tables
- **users** - User accounts and authentication
- **donations** - Donation records and transaction tracking
- **subscriptions** - Recurring donation management
- **admin_users** - Administrative user access

### Data Relationships
- Users can have multiple donations (one-to-many)
- Users can have multiple subscriptions (one-to-many)  
- Donations link to users when logged in (optional foreign key)
- Subscriptions always require user accounts (required foreign key)

## 📋 Development Standards

### Code Organization
- **Actions** - HTTP handlers and business logic
- **Models** - Database entities and validation
- **Services** - External API integrations and complex operations
- **Templates** - Server-side HTML rendering

### Testing Standards
- **ActionSuite** - Integration tests for HTTP endpoints
- **Model tests** - Database interaction and validation testing
- **Service tests** - External API integration testing
- **Manual testing** - Payment flow and user experience verification

### Security Standards
- **Input validation** - All user input validated and sanitized
- **SQL injection prevention** - Parameterized queries via Pop ORM
- **CSRF protection** - Built into Buffalo framework
- **Session security** - Secure session management
- **Payment security** - PCI compliance via Helcim integration

## 🔧 Configuration Management

### Environment Variables
- **Database configuration** - Connection strings and credentials
- **Helcim integration** - API tokens and merchant settings
- **Email configuration** - SMTP settings for receipt delivery
- **Application settings** - Debug modes and feature flags

### Configuration Files
- **database.yml** - Database connection configuration
- **buffalo-app.toml** - Buffalo framework settings
- **.env** - Development environment variables (not in version control)

## 📊 Performance Considerations

### Database Optimization
- **Indexing strategy** - Proper indexes for query performance
- **Connection pooling** - Efficient database connection management
- **Query optimization** - Efficient ORM usage patterns

### Application Performance
- **Template caching** - Production template compilation
- **Asset optimization** - CSS/JS minification and caching
- **HTTP caching** - Appropriate cache headers
- **GZIP compression** - Response compression for bandwidth

## 🔍 Debugging and Troubleshooting

### Logging Strategy
- **Structured logging** - Consistent log format and levels
- **Request tracing** - Track requests through the application
- **Error logging** - Comprehensive error capture and context
- **Performance logging** - Track slow queries and operations

### Development Tools
- **Buffalo CLI** - Development server and code generation
- **Soda** - Database migration and management
- **Make** - Build automation and testing
- **Podman/Docker** - Containerized development environment

## 📚 External Documentation

### Framework Documentation
- **Buffalo Framework**: https://gobuffalo.io/documentation/
- **Pop ORM**: https://gobuffalo.io/documentation/database/
- **Plush Templates**: https://github.com/gobuffalo/plush

### Integration Documentation  
- **Helcim API**: https://devdocs.helcim.com/
- **Pico CSS**: https://picocss.com/
- **HTMX**: https://htmx.org/docs/

For detailed technical information, see the specific reference documents listed above.
