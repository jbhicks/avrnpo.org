# Production Deployment Checklist

**Pre-merge checklist for Buffalo app deployment to Coolify with PostgreSQL**

Use this checklist before merging to `main` branch. Upon merge, a webhook will trigger automatic deployment to Coolify.

---

## 🔄 Migration from Go App to Buffalo + PostgreSQL

**IMPORTANT:** This deployment introduces significant infrastructure changes:
- **Framework:** Go app → Buffalo framework
- **Database:** None/SQLite → PostgreSQL (via Coolify service)
- **Build System:** Direct Go build → Nixpacks auto-detection
- **Migrations:** Manual → Automated via `soda migrate up`

---

## 📋 Pre-Merge Checklist

### 🧪 Code Quality & Testing
- [ ] **All tests pass locally**: `make test` completes successfully
- [ ] **Test coverage adequate**: Core functionality covered
- [ ] **No critical linting issues**: Code meets project standards
- [ ] **Dependencies up to date**: No critical security vulnerabilities
- [ ] **Git status clean**: No uncommitted changes

### 🗄️ Database Readiness
- [ ] **Migration files valid**: All `.fizz` files in `migrations/` directory
- [ ] **Migration status verified**: `soda migrate status` shows all applied
- [ ] **No schema.sql files**: Auto-generated files removed
- [ ] **Database models tested**: Models work with PostgreSQL syntax
- [ ] **Data seeding (if needed)**: Test data procedures documented

### 🔧 Buffalo Framework
- [ ] **Buffalo app builds**: `buffalo build` completes without errors
- [ ] **Configuration valid**: `config/buffalo-app.toml` properly configured
- [ ] **Static assets working**: CSS, JS, images accessible at `/assets/`
- [ ] **Templates render**: All `.plush.html` templates compile correctly
- [ ] **Routes functional**: All endpoints respond as expected

### 🔒 Environment & Secrets
- [ ] **Environment variables documented**: All required vars listed below
- [ ] **No secrets in code**: `.env` files not committed
- [ ] **Production config ready**: `GO_ENV=production` tested
- [ ] **SMTP credentials valid**: Email sending functional
- [ ] **Helcim API key current**: Payment processing ready

### 💳 Payment System
- [ ] **Helcim integration tested**: Both test and live modes
- [ ] **Receipt generation working**: Email receipts with proper branding
- [ ] **Subscription handling ready**: Monthly donations functional
- [ ] **Error handling complete**: Failed payments handled gracefully

### 🔗 Integration Testing
- [ ] **Email delivery tested**: Receipts and notifications working
- [ ] **Admin dashboard functional**: User management operational
- [ ] **Blog system working**: Post creation and display functional
- [ ] **Donation flow complete**: End-to-end donation process tested

---

## 🌐 Coolify Infrastructure Checklist

### 📊 Database Service (PostgreSQL)
- [ ] **PostgreSQL service provisioned**: Database service running in Coolify
- [ ] **Database linked to app**: `DATABASE_URL` environment variable set
- [ ] **Connection tested**: App can connect to database
- [ ] **Backup strategy confirmed**: Database backup procedures in place

### 🚀 Application Deployment
- [ ] **Nixpacks build pack selected**: Auto-detection configured for Go/Buffalo
- [ ] **Migration command configured**: `soda migrate up` set as pre-start command
- [ ] **Domain configured**: `avrnpo.org` pointing to Coolify app
- [ ] **SSL certificate ready**: HTTPS properly configured

### 🔧 Environment Variables Set

**Required Environment Variables in Coolify:**

| Variable | Purpose | Status |
|----------|---------|---------|
| `GO_ENV` | Set to `production` | [ ] |
| `DATABASE_URL` | PostgreSQL connection (auto-set by Coolify) | [ ] |
| `HELCIM_PRIVATE_API_KEY` | Payment processing | [ ] |
| `SMTP_HOST` | Email delivery | [ ] |
| `SMTP_PORT` | Email delivery | [ ] |
| `SMTP_USERNAME` | Email authentication | [ ] |
| `SMTP_PASSWORD` | Email authentication | [ ] |
| `FROM_EMAIL` | Sender email address | [ ] |
| `FROM_NAME` | Sender display name | [ ] |
| `ADMIN_EMAIL` | Initial admin user email | [ ] |
| `ADMIN_PASSWORD` | Initial admin user password | [ ] |
| `ADMIN_FIRST_NAME` | Initial admin first name | [ ] |
| `ADMIN_LAST_NAME` | Initial admin last name | [ ] |

**Optional but Recommended:**
| Variable | Purpose | Status |
|----------|---------|---------|
| `LOG_LEVEL` | Application logging level | [ ] |
| `LOG_FILE_PATH` | Log file location | [ ] |

---

## 🔍 Pre-Deployment Testing

### 🧪 Local Testing with Production Config
- [ ] **Production environment simulation**: Test with `GO_ENV=production`
- [ ] **PostgreSQL compatibility**: Test against real PostgreSQL instance
- [ ] **Migration rollback tested**: Ensure migrations can be safely reverted
- [ ] **Performance testing**: Application performs adequately under load

### 💰 Payment Integration Final Check
- [ ] **Test credit cards work**: Use Helcim test card numbers
- [ ] **Receipt emails deliver**: End-to-end email verification
- [ ] **Subscription creation**: Monthly donations process correctly
- [ ] **Error scenarios handled**: Invalid cards, network failures, etc.

---

## 🚨 Deployment Safety

### 🔄 Rollback Plan
- [ ] **Previous version identified**: Know how to revert if needed
- [ ] **Database backup current**: Fresh backup before deployment
- [ ] **Rollback procedure documented**: Clear steps to revert changes
- [ ] **Emergency contacts ready**: Team members available during deployment

### 📊 Monitoring Ready
- [ ] **Health check endpoint**: Application health monitoring configured
- [ ] **Log monitoring**: Application logs accessible in Coolify
- [ ] **Error alerting**: Critical errors trigger notifications
- [ ] **Performance baseline**: Know expected response times and throughput

---

## ✅ Final Verification

### 🔧 Technical Readiness
- [ ] All items in this checklist completed
- [ ] Code review approved by team member
- [ ] Breaking changes documented
- [ ] Feature flags configured (if applicable)

### 📢 Communication
- [ ] **Stakeholders notified**: Deployment timing communicated
- [ ] **Maintenance window scheduled**: If downtime expected
- [ ] **Documentation updated**: User-facing changes documented
- [ ] **Support team briefed**: Customer service aware of changes

---

## 🚀 Post-Merge Verification

**After merge triggers deployment, verify:**

- [ ] **Site loads**: https://avrnpo.org responds
- [ ] **Database connected**: No database connection errors in logs
- [ ] **Migrations applied**: Database schema up to date
- [ ] **Static assets serve**: CSS, JS, images load correctly
- [ ] **Email functions**: Test receipt generation and delivery
- [ ] **Admin access**: Dashboard accessible and functional
- [ ] **Payment processing**: Test donation flow works
- [ ] **Performance acceptable**: Site responds within acceptable time

---

## 🆘 Emergency Procedures

**If deployment fails:**

1. **Check Coolify logs**: Review build and runtime logs for errors
2. **Verify environment variables**: Ensure all required variables set
3. **Database connectivity**: Confirm PostgreSQL service running and linked
4. **Migration issues**: Check migration logs for schema conflicts
5. **Rollback if necessary**: Revert to previous working version
6. **Contact team**: Escalate to development team if issues persist

---

## 📞 Emergency Contacts

- **Development Team**: [Contact Information]
- **System Administrator**: [Contact Information]  
- **Database Administrator**: [Contact Information]

---

**Deployment Date**: ________________  
**Deployed By**: ____________________  
**Verification Completed By**: _______

**✅ APPROVED FOR MERGE TO MAIN**