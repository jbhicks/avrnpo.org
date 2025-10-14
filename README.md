# American Veterans Rebuilding (AVR NPO)

<!-- Deployment trigger: 2025-09-02 -->

Official website for American Veterans Rebuilding, a 501(c)(3) non-profit organization dedicated to helping combat veterans rebuild their lives through housing projects, skills training, and community support programs.

## About AVR NPO

American Veterans Rebuilding is formed by Combat Veterans of the wars in Afghanistan and Iraq. We are soldiers who have lived through hell on earth and found a way to continue to dedicate our lives to the military's core values of Loyalty, Duty, Respect, Selfless Service, Honor, Integrity and Personal Courage.

## 🚀 Quick Start

### Prerequisites
- **Go 1.23+** - [Download Go](https://golang.org/dl/)
- **Templ CLI** - `go install github.com/a-h/templ/cmd/templ@latest`

### Local Development

```console
# Clone the repository
git clone <repository-url>
cd avrnpo.org

# Copy environment template
cp .env.example .env

# Edit .env with your settings (SMTP, Helcim, admin credentials)
# Important: Set PB_ADMIN_EMAIL and PB_ADMIN_PASSWORD

# Build and start development server
make dev
```

After setup, visit [http://127.0.0.1:8090](http://127.0.0.1:8090) to see the website running locally.

Admin dashboard available at [http://127.0.0.1:8090/_/](http://127.0.0.1:8090/_/)

### Development Commands

```console
# Start development server with hot reload
make dev

# Run unit tests
go test ./...

# Run E2E tests
E2E_TESTS=1 go test -v -run E2E

# Regenerate Templ templates
templ generate

# Build production binary
go build
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

- **Backend**: Go 1.23.0 + PocketBase v0.22+ (embedded SQLite database)
- **Templates**: Templ v0.3.943 (type-safe Go templates)
- **Frontend**: HTMX, Pico CSS v2 (semantic CSS framework)
- **Payments**: Helcim Payment and Recurring APIs
- **Authentication**: PocketBase auth with role management
- **Deployment**: Container-ready with Docker

## 💳 Helcim Integration Quickstart

The donation system uses Helcim's official payment integration. For development and testing:

### Official Integration Pattern
- **Script URL**: `https://secure.helcim.app/helcim-pay/services/start.js`
- **API Endpoint**: `POST /v2/helcim-pay/initialize`
- **Modal Function**: `appendHelcimPayIframe(checkoutToken)`
- **Events**: PostMessage events (SUCCESS, ABORTED, HIDE)

### Development Setup
```bash
# Validate Helcim URLs in codebase
./scripts/validate-helcim-urls.sh

# Test donation flow
make dev
# Visit: http://127.0.0.1:3000/donate
```

### Key Files
- **Templates**: `templates/*.templ` (Templ templates)
- **Backend**: `main.go`, `services/helcim.go`
- **Docs**: `docs/payment-system/`

### Testing
- Use official Helcim test cards: `4124939999999990` (CVV: 100)
- Check browser console for Helcim script loading
- Monitor server logs for payment processing

See [Payment System Documentation](./docs/payment-system/) for complete details.

## 📚 Documentation

For detailed development information, see the [comprehensive documentation](./docs/):

### Getting Started
- **[Quick Start Guide](./docs/getting-started/quick-start.md)** - Detailed setup instructions
- **[Development Workflow](./docs/getting-started/development-workflow.md)** - Daily development commands
- **[Development Guide](./docs/DEVELOPMENT_GUIDE.md)** - Complete framework documentation

### Core Systems
- **[Payment System](./docs/payment-system/README.md)** - Donation and subscription management
- **[Frontend Development](./docs/frontend/README.md)** - HTMX patterns and Pico CSS styling
- **[PocketBase](https://pocketbase.io/docs/)** - Database and admin UI

### Deployment & Production
- **[Coolify Deployment](./docs/deployment/coolify-pocketbase-migration.md)** - PocketBase deployment guide
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
2. **Follow Conventions** - Check [AGENTS.md](./AGENTS.md) for patterns
3. **Test Changes** - Use `go test ./...` to verify all functionality works
4. **Security First** - Follow [security guidelines](./docs/deployment/security.md)

## 📞 Contact

**For AVR NPO Programs:**
- Website: [avrnpo.org](https://avrnpo.org)
- Email: michael@avrnpo.org

**For Technical Issues:**
- Review documentation in [./docs/](./docs/)
- Check [troubleshooting guides](./docs/deployment/README.md)

## 📝 License

This website code is built on open-source technologies. Content and imagery related to American Veterans Rebuilding is proprietary to the organization.

---

*Supporting combat veterans in rebuilding their lives and strengthening communities.*