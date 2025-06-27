# Modern Gmail Integration Implementation Guide

## 🚨 Important: Google Authentication Changes

Google has **significantly changed email authentication requirements**:

1. **App Passwords**: Being phased out for most applications
2. **Less Secure Apps**: Completely deprecated as of May 2022
3. **OAuth2 Required**: Now the standard for programmatic access
4. **Service Accounts**: Recommended for server-to-server applications

## 🎯 Implementation Strategy

### Phase 1: Update Dependencies (Required)

Add these to your `go.mod`:

```bash
go get golang.org/x/oauth2/google
go get google.golang.org/api/gmail/v1
go get google.golang.org/api/option
```

### Phase 2: Choose Your Authentication Method

#### Method A: Service Account (Recommended for Production)

**Best for:** Production servers, automated systems, organization emails

1. **Create Service Account**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create new project or select existing
   - Enable Gmail API
   - Create Service Account
   - Download JSON key file

2. **Configure Domain-Wide Delegation** (for @avrnpo.org emails):
   - In Google Admin Console: Security → API Controls → Domain-wide Delegation
   - Add Service Account Client ID
   - Grant Gmail scopes: `https://www.googleapis.com/auth/gmail.send`

3. **Environment Variables**:
   ```bash
   GOOGLE_SERVICE_ACCOUNT_FILE=/path/to/service-account-key.json
   FROM_EMAIL=noreply@avrnpo.org
   FROM_NAME=American Veterans Rebuilding
   ```

#### Method B: OAuth2 with Refresh Token (Good for Development)

**Best for:** Development, testing, personal Gmail accounts

1. **Create OAuth2 Credentials**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
   - Create OAuth 2.0 Client ID (Desktop Application)
   - Download `client_secret.json`

2. **Generate Refresh Token** (one-time setup):
   ```bash
   # Use Google OAuth2 Playground or custom script
   # Scopes: https://www.googleapis.com/auth/gmail.send
   ```

3. **Environment Variables**:
   ```bash
   GOOGLE_CLIENT_ID=your-client-id.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your-client-secret
   GOOGLE_REFRESH_TOKEN=your-refresh-token
   FROM_EMAIL=your-gmail@gmail.com
   FROM_NAME=American Veterans Rebuilding
   ```

### Phase 3: Replace Email Service

1. **Backup Current Implementation**:
   ```bash
   cp services/email.go services/email_backup.go
   ```

2. **Replace with Modern Implementation**:
   ```bash
   cp services/email_v2.go services/email.go
   ```

3. **Update Imports** (if needed):
   ```go
   import (
       "golang.org/x/oauth2"
       "golang.org/x/oauth2/google"
       "google.golang.org/api/gmail/v1"
       "google.golang.org/api/option"
   )
   ```

### Phase 4: Testing

1. **Test Configuration**:
   ```bash
   # Check if service is properly configured
   make dev
   # Look for email service initialization logs
   ```

2. **Test Receipt Sending**:
   ```bash
   # Make a test donation
   # Check Buffalo logs for email status:
   tail -f buffalo.log | grep -i email
   ```

## 🔧 Troubleshooting Common Issues

### Service Account Issues

**Problem**: "insufficient permissions" or "domain-wide delegation required"
**Solution**: 
- Ensure domain-wide delegation is configured in Google Admin Console
- Service account must be granted Gmail send permissions
- FROM_EMAIL must match a domain that the service account can impersonate

### OAuth2 Issues

**Problem**: "invalid_grant" or "token expired"
**Solution**:
- Refresh tokens can expire if not used for 6 months
- Regenerate refresh token using OAuth2 flow
- Ensure scopes match what was originally granted

### General Issues

**Problem**: "API not enabled"
**Solution**: Enable Gmail API in Google Cloud Console

**Problem**: "quota exceeded"  
**Solution**: Gmail API has daily sending limits (check Cloud Console quotas)

## 🚀 Quick Start for Development

For immediate testing with personal Gmail:

1. **Enable 2FA** on your Google account
2. **Generate App Password** (if still available): https://myaccount.google.com/apppasswords
3. **Use Legacy SMTP** temporarily:
   ```bash
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your-gmail@gmail.com
   SMTP_PASSWORD=your-16-char-app-password
   FROM_EMAIL=your-gmail@gmail.com
   FROM_NAME=American Veterans Rebuilding
   ```

⚠️ **Note**: App passwords may be disabled on your account. OAuth2 is the recommended approach.

## 📋 Production Checklist

- [ ] Google Cloud Project created
- [ ] Gmail API enabled
- [ ] Service Account created and configured
- [ ] Domain-wide delegation set up (for org emails)
- [ ] JSON key file stored securely
- [ ] Environment variables configured
- [ ] Email service updated and tested
- [ ] Receipt sending verified with test donation
- [ ] Error handling and logging confirmed

## 🔐 Security Best Practices

1. **Service Account Keys**: Store JSON files securely, never commit to git
2. **Environment Variables**: Use proper secret management in production
3. **Principle of Least Privilege**: Only grant necessary Gmail scopes
4. **Key Rotation**: Regularly rotate service account keys
5. **Monitoring**: Monitor API usage and quota consumption

## 🆘 Need Help?

- **Google Cloud Support**: For service account and API issues
- **Gmail API Documentation**: https://developers.google.com/gmail/api
- **OAuth2 Troubleshooting**: https://developers.google.com/identity/protocols/oauth2

The new implementation maintains backward compatibility with SMTP while adding modern Gmail API support.
