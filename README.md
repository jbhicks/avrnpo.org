# My Go SaaS Template

A Buffalo-based SaaS application template with containerized PostgreSQL database.

## ✅ Setup Checklist

- [x] **Buffalo application generated** - Basic Buffalo app structure created
- [x] **PostgreSQL containerized** - Docker Compose setup with PostgreSQL 15
- [x] **Database configuration** - All databases created (development, test, production)
- [x] **Database migrations** - Schema up to date and working
- [x] **Application running** - Buffalo dev server successfully connecting to database
- [ ] Authentication system
- [ ] User management
- [ ] Billing/subscription features
- [ ] Email services
- [ ] Production deployment

## 🚀 Quick Start

### Prerequisites
- Go 1.19+
- Docker and Docker Compose
- Buffalo CLI

### Database Setup

This application uses PostgreSQL running in a Docker container. The setup is fully automated:

```console
# Start PostgreSQL container
docker-compose up -d

# Run database migrations (if needed)
buffalo pop migrate
```

### Starting the Application

```console
# Start the development server
buffalo dev
```

Visit [http://127.0.0.1:3000](http://127.0.0.1:3000) to see your application.

## 📊 Architecture

- **Backend**: Buffalo (Go web framework)
- **Database**: PostgreSQL 15 (containerized)
- **Frontend**: Plush templates with Bootstrap
- **Background Jobs**: Buffalo workers

## 🛠️ Development

### Database Management

The application includes a PostgreSQL container configured via `docker-compose.yml`:

- **Development DB**: `my_go_saas_template_development`
- **Test DB**: `my_go_saas_template_test`
- **Production DB**: `my_go_saas_template_production`

### Common Commands

```console
# Database operations
buffalo pop create -a          # Create databases
buffalo pop migrate            # Run migrations
buffalo pop generate migration # Create new migration

# Development
buffalo dev                    # Start dev server with hot reload
buffalo build                  # Build production binary
buffalo test                   # Run tests
```

## 🐛 Troubleshooting Log

### Issue 1: Database Connection Refused (Resolved ✅)
**Date**: May 24, 2025  
**Problem**: Buffalo app couldn't connect to PostgreSQL - "dial tcp 127.0.0.1:5432: connect: connection refused"  
**Root Cause**: PostgreSQL was not running  
**Solution**: 
1. Created `docker-compose.yml` with PostgreSQL 15 container
2. Added `init.sql` to create test and production databases
3. Started container with `docker-compose up -d`
4. Restarted Buffalo dev server to establish connection

**Files Modified**:
- Added `docker-compose.yml`
- Added `init.sql`

### Issue 2: Port Conflict During Restart (Resolved ✅)
**Date**: May 24, 2025  
**Problem**: "address already in use" when restarting Buffalo dev server  
**Solution**: Properly stopped old processes with `pkill -f "my-go-saas-template-build"`

## 📁 Project Structure

```
.
├── actions/          # Controllers and HTTP handlers
├── cmd/app/         # Application entry point
├── config/          # Buffalo configuration
├── migrations/      # Database migrations
├── models/          # Database models
├── templates/       # HTML templates
├── public/          # Static assets
├── docker-compose.yml  # PostgreSQL container setup
├── database.yml     # Database configuration
└── README.md        # This file
```

## 🔗 Resources

- [Buffalo Documentation](http://gobuffalo.io)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Compose Reference](https://docs.docker.com/compose/)

---

**Status**: ✅ Base infrastructure complete - Ready for SaaS feature development  
**Last Updated**: May 24, 2025
