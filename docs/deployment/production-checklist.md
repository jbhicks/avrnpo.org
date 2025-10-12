# Production Deployment Checklist

**Pre-deployment checklist for PocketBase app deployment to Coolify**

Use this checklist before deploying to production. The application uses PocketBase with SQLite, Templ templates, and HTMX for frontend interactions.

---

## 📋 Pre-Deployment Checklist

### 🧪 Code Quality & Testing
- [ ] **All tests pass locally**: `go test ./...` completes successfully
- [ ] **E2E tests pass**: `E2E_TESTS=1 go test -v -run E2E` completes successfully
- [ ] **Test coverage adequate**: Core functionality covered
- [ ] **No critical linting issues**: `golangci-lint run` passes
- [ ] **Dependencies up to date**: No critical security vulnerabilities
- [ ] **Git status clean**: No uncommitted changes

### 🗄️ Database & Migrations
- [ ] **Migration files valid**: All `.go` files in `pb_migrations/` directory
- [ ] **Migrations tested**: Applied successfully in development
- [ ] **Collections verified**: All required collections defined
- [ ] **Indexes optimized**: Database queries perform well
- [ ] **Initial data ready**: Seed data procedures documented

### 🎨 Frontend & Templates
- [ ] **Templ templates compiled**: `templ generate` completes without errors
- [ ] **Static assets ready**: CSS, JS, images in `pb_public/assets/`
- [ ] **HTMX interactions tested**: All dynamic features working
- [ ] **Pico CSS theme working**: Dark/light mode toggle functional
- [ ] **Forms validated**: All forms include CSRF protection
- [ ] **Responsive design verified**: Mobile and desktop layouts tested

### 🔒 Environment & Secrets
- [ ] **Environment variables documented**: All required vars listed below
- [ ] **No secrets in code**: `.env` files not committed
- [ ] **SMTP credentials valid**: Email sending functional
- [ ] **Helcim API key current**: Payment processing ready
- [ ] **Admin credentials secure**: Strong password for admin account

### 💳 Payment System
- [ ] **Helcim integration tested**: Both test and live modes
- [ ] **Receipt generation working**: Email receipts with proper branding
- [ ] **Subscription handling ready**: Monthly donations functional
- [ ] **Error handling complete**: Failed payments handled gracefully
- [ ] **Webhook endpoints secure**: CSRF exemptions properly configured

### 🔗 Integration Testing
- [ ] **Email delivery tested**: Receipts and notifications working
- [ ] **Admin dashboard functional**: PocketBase admin at `/_/` operational
- [ ] **Blog system working**: Post creation and display functional
- [ ] **Donation flow complete**: End-to-end donation process tested
- [ ] **Contact form working**: Email delivery and rate limiting functional

---

## 🌐 Coolify Infrastructure Checklist

### 🚀 Application Deployment
- [ ] **Build command configured**: `make build` or `go build -o avrnpo`
- [ ] **Start command configured**: `./avrnpo serve --http=0.0.0.0:8090`
- [ ] **Port mapping**: Port 8090 exposed and mapped
- [ ] **Domain configured**: `avrnpo.org` pointing to Coolify app
- [ ] **SSL certificate ready**: HTTPS properly configured

### 💾 Persistent Storage
- [ ] **Volume mounted**: `pb_data/` directory persisted across deployments
- [ ] **SQLite database persistent**: Database file not lost on restart
- [ ] **Uploads directory persistent**: User-uploaded files preserved
- [ ] **Backup strategy confirmed**: Database backup procedures in place

### 🔧 Environment Variables Set

**Required Environment Variables in Coolify:**

| Variable | Purpose | Status |
|----------|---------|---------|
| `PB_ADMIN_EMAIL` | Admin user email | [ ] |
| `PB_ADMIN_PASSWORD` | Admin user password (secure!) | [ ] |
| `HELCIM_API_TOKEN` | Payment processing | [ ] |
| `HELCIM_TEST_MODE` | Set to `false` for production | [ ] |
| `SMTP_HOST` | Email delivery | [ ] |
| `SMTP_PORT` | Email delivery | [ ] |
| `SMTP_USERNAME` | Email authentication | [ ] |
| `SMTP_PASSWORD` | Email authentication | [ ] |
| `FROM_EMAIL` | Sender email address | [ ] |
| `FROM_NAME` | Sender display name | [ ] |
| `EMAIL_ENABLED` | Set to `true` for production | [ ] |
| `CONTACT_EMAIL` | Email for contact form submissions | [ ] |

**Optional but Recommended:**
| Variable | Purpose | Status |
|----------|---------|---------|
| `LOG_LEVEL` | Application logging level | [ ] |
| `PB_ENCRYPTION_KEY` | Encryption key for sensitive data | [ ] |

---

## 🔍 Pre-Deployment Testing

### 🧪 Local Testing with Production Config
- [ ] **Production environment simulation**: Test with production-like settings
- [ ] **Database performance**: SQLite performs adequately under load
- [ ] **Migration testing**: All migrations apply cleanly
- [ ] **Concurrent requests**: Application handles multiple simultaneous users

### 💰 Payment Integration Final Check
- [ ] **Test credit cards work**: Use Helcim test card numbers
- [ ] **Receipt emails deliver**: End-to-end email verification
- [ ] **Subscription creation**: Monthly donations process correctly
- [ ] **Error scenarios handled**: Invalid cards, network failures, etc.
- [ ] **Switch to live mode**: `HELCIM_TEST_MODE=false` verified

### 🔐 Security Verification
- [ ] **CSRF protection active**: All forms protected
- [ ] **Rate limiting configured**: Login and contact forms rate-limited
- [ ] **Input validation**: All user input sanitized
- [ ] **XSS prevention**: Content sanitization working
- [ ] **SQL injection safe**: Parameterized queries used
- [ ] **Admin UI secured**: Only accessible with authentication

---

## 🚨 Deployment Safety

### 🔄 Rollback Plan
- [ ] **Previous version identified**: Know how to revert if needed
- [ ] **Database backup current**: Fresh backup before deployment
- [ ] **Rollback procedure documented**: Clear steps to revert changes
- [ ] **Emergency contacts ready**: Team members available during deployment

### 📊 Monitoring Ready
- [ ] **Health check endpoint**: `/api/health` configured
- [ ] **Log monitoring**: Application logs accessible in Coolify
- [ ] **Error alerting**: Critical errors trigger notifications
- [ ] **Performance baseline**: Know expected response times

---

## ✅ Final Verification

### 🔧 Technical Readiness
- [ ] All items in this checklist completed
- [ ] Code review approved
- [ ] Breaking changes documented
- [ ] Build tested in production mode

### 📢 Communication
- [ ] **Stakeholders notified**: Deployment timing communicated
- [ ] **Maintenance window scheduled**: If downtime expected
- [ ] **Documentation updated**: User-facing changes documented
- [ ] **Support team briefed**: Team aware of changes

---

## 🚀 Post-Deployment Verification

**After deployment completes, verify:**

- [ ] **Site loads**: https://avrnpo.org responds
- [ ] **Database working**: No database errors in logs
- [ ] **PocketBase admin accessible**: `https://avrnpo.org/_/` loads
- [ ] **Static assets serve**: CSS, JS, images load correctly
- [ ] **Email functions**: Test contact form and receipt delivery
- [ ] **Admin access**: Dashboard accessible and functional
- [ ] **Payment processing**: Test donation flow works
- [ ] **Performance acceptable**: Site responds within acceptable time
- [ ] **SSL certificate valid**: HTTPS working correctly
- [ ] **Blog posts visible**: Content displays correctly
- [ ] **Theme toggle works**: Dark/light mode switching functional

---

## 🆘 Emergency Procedures

**If deployment fails:**

1. **Check Coolify logs**: Review build and runtime logs for errors
2. **Verify environment variables**: Ensure all required variables set
3. **Database connectivity**: Confirm `pb_data/` volume mounted correctly
4. **Port configuration**: Verify port 8090 is exposed
5. **Migration issues**: Check migration logs for errors
6. **Rollback if necessary**: Revert to previous working version
7. **Contact team**: Escalate to development team if issues persist

### Common Issues

**Database locked errors:**
- Ensure only one instance of PocketBase is running
- Check that `pb_data/` is properly mounted as persistent volume

**Migration failures:**
- Review migration logs in Coolify
- Check that all migration files are included in build

**Static assets not loading:**
- Verify `pb_public/` directory is included in deployment
- Check that assets are accessible at `/assets/*` paths

**Email not sending:**
- Verify SMTP credentials are correct
- Check that `EMAIL_ENABLED=true`
- Review email service logs

---

## 📞 Emergency Contacts

- **Development Team**: [Contact Information]
- **System Administrator**: [Contact Information]  
- **Coolify Support**: [Contact Information]

---

## 🔧 PocketBase-Specific Deployment Notes

### Database Backup Before Deployment
```bash
# Backup SQLite database
cp pb_data/data.db pb_data/data.db.backup-$(date +%Y%m%d-%H%M%S)
```

### Testing Migration Rollback
```bash
# PocketBase automatically manages migrations
# No manual rollback needed - restore from backup if necessary
```

### Verifying Admin User Creation
```bash
# Check admin user exists
./avrnpo admin:check
# Or check directly in PocketBase admin UI at /_/
```

---

**Deployment Date**: ________________  
**Deployed By**: ____________________  
**Verification Completed By**: _______

**✅ APPROVED FOR DEPLOYMENT**
