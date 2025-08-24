# American Veterans Rebuilding (AVR NPO)

Official website for American Veterans Rebuilding, a 501(c)(3) non-profit organization dedicated to helping combat veterans rebuild their lives through housing projects, skills training, and community support programs.

## About AVR NPO

American Veterans Rebuilding is formed by Combat Veterans of the wars in Afghanistan and Iraq. We are soldiers who have lived through hell on earth and found a way to continue to dedicate our lives to the military's core values of Loyalty, Duty, Respect, Selfless Service, Honor, Integrity and Personal Courage.

## 🚀 Quick Start

### Prerequisites
- **Go 1.19+** - [Download Go](https://golang.org/dl/)
- **Podman** or **Docker** - [Install Podman](https://podman.io/getting-started/installation)
- **Buffalo CLI** - `go install github.com/gobuffalo/cli/cmd/buffalo@latest`

### Local Development

```console
# Clone the repository
git clone <repository-url>
cd avrnpo.org

# Complete setup (database + migrations + first run)
make setup

# Start development mode
make dev
```

After setup, visit [http://127.0.0.1:3000](http://127.0.0.1:3000) to see the website running locally.

### Development Commands

```console
# Start development server with hot reload
make dev

# Run tests
make test

# Reset database (development)
make db-reset

# Create admin user (promote first registered user)
make admin
```

## 🌟 Website Features

### Public Features
- **Mission & About** - Information about AVR's mission and impact
- **Team Profiles** - Meet the combat veterans who founded and run AVR
- **Project Showcase** - Housing and community development projects
- **Contact Information** - Ways to reach out and get involved
- **Donation System** - Secure donation processing with Helcim integration
  - ✅ One-time donations 
  - ✅ Monthly recurring subscriptions
  - ✅ User account linking and subscription management
  - ✅ Automated email receipts

### Content Management
- **Blog System** - News updates and success stories
- **Admin Dashboard** - Content management for authorized users
- **SEO Optimization** - Search engine friendly with meta tags
- **HTMX Navigation** - Fast, dynamic page loading without full refreshes

## 🛠️ Technology Stack

- **Backend**: Buffalo (Go web framework), PostgreSQL
- **Frontend**: HTMX, Pico.css (semantic CSS framework)
- **Payments**: Helcim Payment and Recurring APIs
- **Authentication**: Session-based with role management
- **Deployment**: Container-ready with Docker/Podman

## 📚 Documentation

For detailed development information, see the [comprehensive documentation](./docs/):

### Getting Started
- **[Quick Start Guide](./docs/getting-started/quick-start.md)** - Detailed setup instructions
- **[Development Workflow](./docs/getting-started/development-workflow.md)** - Daily development commands
- **[Development Guide](./docs/DEVELOPMENT_GUIDE.md)** - Complete framework documentation

### Core Systems
- **[Payment System](./docs/payment-system/README.md)** - Donation and subscription management
- **[Buffalo Framework](./docs/buffalo-framework/README.md)** - Web framework patterns and best practices
- **[Frontend Development](./docs/frontend/README.md)** - HTMX patterns and Pico.css styling

### Deployment & Production
- **[Deployment Guide](./docs/deployment/README.md)** - Production deployment procedures
- **[Security Guidelines](./docs/deployment/security.md)** - Security best practices

## 🎯 Project Status

### ✅ Production Ready
- **Donation System** - Complete Helcim integration with one-time and recurring donations
- **User Management** - Registration, authentication, and role-based access
- **Content Management** - Blog system with admin panel
- **Email System** - Automated receipts and contact form processing

### 🔄 Current Focus
- **User Experience** - Optimizing donation flows and user interfaces
- **Documentation** - Comprehensive developer and deployment guides
- **Testing** - Robust testing procedures for payment system reliability

## 🤝 Contributing

This is a private repository for AVR NPO. For development work:

1. **Review Documentation** - Start with [docs/getting-started/](./docs/getting-started/)
2. **Follow Conventions** - Check [docs/buffalo-framework/](./docs/buffalo-framework/) for patterns
3. **Test Changes** - Use `make test` to verify all functionality works
4. **Security First** - Follow [security guidelines](./docs/deployment/security.md)

## 📞 Contact

**For AVR NPO Programs:**
- Website: [avrnpo.org](https://avrnpo.org)
- Email: michael@avrnpo.org

**For Technical Issues:**
- Review documentation in [./docs/](./docs/)
- Check [troubleshooting guides](./docs/buffalo-framework/troubleshooting.md)

## 📝 License

This website code is built on open-source technologies. Content and imagery related to American Veterans Rebuilding is proprietary to the organization.

---

*Supporting combat veterans in rebuilding their lives and strengthening communities.*