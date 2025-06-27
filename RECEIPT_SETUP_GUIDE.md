# Quick Receipt Setup Guide

## ✅ What We Have (Already Built)

The AVR NPO donation system has a **complete receipt system** ready to use:

- ✅ Professional email templates with AVR branding
- ✅ Tax-compliant donation receipts 
- ✅ Automatic sending after successful donations
- ✅ Error handling that doesn't break donation flow
- ✅ Both HTML and plain text versions

## 🚀 Quick Setup (Choose Your Method)

Google has **discontinued app passwords** for most use cases and now recommends more secure authentication methods. Here are the current options:

### ⭐ **Recommended: Gmail API with Service Account (Most Secure)**

⚠️ **IMPORTANT: Google Secure-by-Default Policies**

**Organizations created after May 3, 2024** automatically have security policies that **block service account key creation**. This affects the traditional Gmail API setup method.

**Key Security Constraints Applied:**
- `constraints/iam.disableServiceAccountKeyCreation` - Blocks creating service account JSON keys
- `constraints/iam.disableServiceAccountKeyUpload` - Blocks uploading external keys
- `constraints/iam.automaticIamGrantsForDefaultServiceAccounts` - Removes excessive default permissions

**Your Options:**
1. **🔐 Most Secure**: Use workload identity federation (complex setup)
2. **⚖️ Balanced**: Temporarily disable constraints to create keys, then re-enable
3. **🤝 Organizational**: Have your admin create keys for you
4. **🔄 Alternative**: Use OAuth2 with refresh tokens (simpler but requires user consent)

This guide covers **Option 2** (balanced approach) as it provides good security while remaining practical for email services.

#### Step 1: Create Google Cloud Project

1. **Go to Google Cloud Console**: https://console.cloud.google.com/
2. **Click "Select a project"** → **"New Project"**
3. **Enter project details**:
   - Project name: `AVR NPO Email Service`
   - Organization: (select your organization if applicable)
   - Location: (leave default or select appropriate)
4. **Click "CREATE"**
5. **Wait for project creation** (usually takes 30-60 seconds)

#### Step 2: Enable Gmail API

1. **In your new project**, go to **"APIs & Services"** → **"Library"**
2. **Search for "Gmail API"**
3. **Click on "Gmail API"** from the results
4. **Click "ENABLE"** button
5. **Wait for API to be enabled** (should be immediate)

#### Step 3: Create Service Account

1. **Go to "APIs & Services"** → **"Credentials"**
2. **Click "CREATE CREDENTIALS"** → **"Service account"**
3. **Fill in service account details**:
   - Service account name: `avr-email-service`
   - Service account ID: `avr-email-service` (auto-filled)
   - Description: `Service account for sending donation receipts`
4. **Click "CREATE AND CONTINUE"**
5. **Grant minimal roles** (security best practice):
   - Click "Select a role" → Search for "Service Account" → Select **"Service Account Token Creator"**
   - Click "ADD ANOTHER ROLE"
   - Search for "Gmail" → Select **"Gmail API User"** (if available, or skip this step)
   - **❌ AVOID "Project Editor"** - This role is overly permissive and violates security best practices
6. **Click "CONTINUE"**
7. **Skip "Grant users access"** → **Click "DONE"**

##### Understanding the Minimal Permissions Approach

**Why we avoid "Project Editor":**
- ✅ **"Service Account Token Creator"** - Allows the service account to generate its own access tokens for Gmail API
- ❌ **"Project Editor"** - Grants broad permissions to create/delete/modify almost all Google Cloud resources
- 🎯 **Security Principle**: Use minimal permissions needed for the specific task (sending emails)

**What permissions we actually need:**
1. **Gmail API access** - Enabled at the project level (Step 2)
2. **Token generation** - "Service Account Token Creator" role
3. **Domain authorization** - Domain-wide delegation (Step 5, if using organization email)

**If you encounter permission errors later:**
- The service account only needs to send emails via Gmail API
- Domain-wide delegation provides the "send as organization user" permission
- No additional Google Cloud resource permissions should be needed

#### Step 4: Create Service Account Key

⚠️ **CRITICAL: Handle Secure-by-Default Organization Policies**

**If your Google Cloud organization was created after May 3, 2024**, it automatically enforces secure-by-default policies that **block service account key creation**. You'll need to handle this first.

##### Check if Key Creation is Blocked

1. **Attempt to create a key** (next steps). If you get an error like:
   ```
   Permission denied: Service account key creation is disabled by organization policy
   ```
   
2. **Check your organization policies**:
   ```bash
   # Get your organization ID first
   gcloud organizations list
   
   # Check for the blocking policy
   gcloud resource-manager org-policies list --organization=YOUR_ORG_ID | grep -i serviceAccountKey
   ```

##### Option A: Use Workload Identity Federation (Most Secure - Recommended)

**Best Practice**: Instead of creating service account keys, use workload identity federation. However, this is complex for Gmail API. Skip to Option B for simpler setup.

##### Option B: Temporarily Disable the Policy (Less Secure)

**⚠️ Security Trade-off**: This reduces security but allows traditional service account keys.

1. **You must be an Organization Policy Administrator**. Check your permissions:
   ```bash
   gcloud organizations get-iam-policy YOUR_ORG_ID --flatten="bindings[].members" --filter="bindings.role:roles/orgpolicy.policyAdmin"
   ```

2. **If you don't have permissions**, ask your organization admin to either:
   - Grant you `roles/orgpolicy.policyAdmin` role
   - Temporarily disable the policy for you
   - Create the service account key for you

3. **Disable the service account key creation policy**:
   ```bash
   # Disable the constraint
   gcloud org-policies delete iam.disableServiceAccountKeyCreation --organization=YOUR_ORG_ID
   ```

4. **Create the service account key** (follow steps below)

5. **Re-enable the policy** (recommended after key creation):
   ```bash
   # Create a policy file to re-enable the constraint
   cat > disable-sa-keys.yaml << EOF
   name: organizations/YOUR_ORG_ID/policies/iam.disableServiceAccountKeyCreation
   spec:
     rules:
     - enforce: true
   EOF
   
   # Apply the policy
   gcloud org-policies set-policy disable-sa-keys.yaml
   ```

##### Option C: Ask Organization Admin

**Simplest approach**: Ask your Google Workspace or Cloud Organization administrator to:
1. Temporarily disable `constraints/iam.disableServiceAccountKeyCreation`
2. Create the service account and key for you
3. Re-enable the policy
4. Provide you with the JSON key file

##### Now Create the Service Account Key

1. **In the Credentials page**, find your service account under **"Service Accounts"**
2. **Click on the service account email** (e.g., `avr-email-service@your-project.iam.gserviceaccount.com`)
3. **Go to "KEYS" tab**
4. **Click "ADD KEY"** → **"Create new key"**
5. **Select "JSON"** format
6. **Click "CREATE"**
7. **Save the downloaded JSON file** securely:
   ```bash
   # Move to a secure location on your server
   mkdir -p /etc/avr-secrets/
   mv ~/Downloads/your-project-xxxxx.json /etc/avr-secrets/gmail-service-account.json
   chmod 600 /etc/avr-secrets/gmail-service-account.json
   ```

#### Step 5: Configure Domain-Wide Delegation (for @avrnpo.org emails)

⚠️ **Required only if sending from your organization domain (e.g., noreply@avrnpo.org)**

**Note**: The minimal roles we assigned above (`Service Account Token Creator`) are sufficient for Gmail API access. Domain-wide delegation provides the authorization to send emails on behalf of organization users.

1. **Copy the Service Account's Client ID**:
   - In the service account details page, copy the **"Unique ID"** (it's a long number)
   
2. **Go to Google Admin Console**: https://admin.google.com/
3. **Navigate to Security** → **Access and data control** → **API controls**
4. **Click "MANAGE DOMAIN WIDE DELEGATION"**
5. **Click "Add new"**
6. **Fill in the form**:
   - Client ID: `paste the Unique ID from step 1`
   - OAuth scopes: `https://www.googleapis.com/auth/gmail.send`
7. **Click "Authorize"**

#### Step 6: Add to `.env` file

```bash
# Gmail API Configuration (Recommended)
GOOGLE_SERVICE_ACCOUNT_FILE=/etc/avr-secrets/gmail-service-account.json
FROM_EMAIL=noreply@avrnpo.org
FROM_NAME=American Veterans Rebuilding

# Optional Tax Information
ORGANIZATION_EIN=your-ein-number
ORGANIZATION_ADDRESS=Your full mailing address
```

#### Step 7: Verify Setup

1. **Check file permissions**:
   ```bash
   ls -la /etc/avr-secrets/gmail-service-account.json
   # Should show: -rw------- (600 permissions)
   ```

2. **Verify JSON file contents**:
   ```bash
   # Should contain these fields:
   jq -r '.type, .project_id, .client_email' /etc/avr-secrets/gmail-service-account.json
   # Output should be:
   # service_account
   # your-project-id
   # avr-email-service@your-project.iam.gserviceaccount.com
   ```

### 🔐 **Alternative: OAuth2 with Refresh Tokens**

#### Step 1: Create OAuth2 Client Credentials

1. **Go to Google Cloud Console**: https://console.cloud.google.com/apis/credentials
2. **Click "CREATE CREDENTIALS"** → **"OAuth client ID"**
3. **Configure OAuth consent screen** (if first time):
   - Click "CONFIGURE CONSENT SCREEN"
   - Choose "External" (for testing) or "Internal" (for organization)
   - Fill required fields:
     - App name: `AVR NPO Email Service`
     - User support email: `your-email@avrnpo.org`
     - Developer contact: `your-email@avrnpo.org`
   - Click "SAVE AND CONTINUE" through all steps
4. **Create OAuth Client ID**:
   - Application type: **"Desktop application"**
   - Name: `AVR Email Desktop Client`
   - Click "CREATE"
5. **Download the JSON file** (client_secret_xxxxx.json)

#### Step 2: Generate Refresh Token

1. **Install Google OAuth2 CLI tool**:
   ```bash
   # Option 1: Use Google's OAuth2 Playground
   # Go to: https://developers.google.com/oauthplayground/
   
   # Option 2: Use a simple Go script (recommended)
   ```

2. **Create token generation script** (`get_refresh_token.go`):
   ```go
   package main
   
   import (
       "context"
       "encoding/json"
       "fmt"
       "log"
       "os"
       
       "golang.org/x/oauth2"
       "golang.org/x/oauth2/google"
   )
   
   func main() {
       // Read client credentials
       credentialsFile := "client_secret.json" // Path to your downloaded file
       b, err := os.ReadFile(credentialsFile)
       if err != nil {
           log.Fatalf("Unable to read client secret file: %v", err)
       }
       
       config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/gmail.send")
       if err != nil {
           log.Fatalf("Unable to parse client secret file to config: %v", err)
       }
       
       // Generate authorization URL
       authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
       fmt.Printf("Go to the following link in your browser:\n%v\n", authURL)
       fmt.Print("Enter the authorization code: ")
       
       var authCode string
       if _, err := fmt.Scan(&authCode); err != nil {
           log.Fatalf("Unable to read authorization code: %v", err)
       }
       
       // Exchange code for token
       tok, err := config.Exchange(context.TODO(), authCode)
       if err != nil {
           log.Fatalf("Unable to retrieve token from web: %v", err)
       }
       
       fmt.Printf("Refresh token: %s\n", tok.RefreshToken)
   }
   ```

3. **Run the script**:
   ```bash
   go mod init token-generator
   go get golang.org/x/oauth2/google
   go run get_refresh_token.go
   # Follow the prompts to get your refresh token
   ```

#### Step 3: Add to `.env` file

```bash
# OAuth2 Configuration
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REFRESH_TOKEN=your-refresh-token-from-step-2
FROM_EMAIL=your-gmail@gmail.com
FROM_NAME=American Veterans Rebuilding
```

### 📧 **Fallback: Modern SMTP with App-Specific Password**

⚠️ **Only works if 2FA is enabled and organization allows app passwords**

1. **Enable 2-Factor Authentication** on Google Account
2. **Generate App Password**: https://myaccount.google.com/apppasswords
3. **Add to `.env` file**:
   ```bash
   # SMTP with App Password (Legacy)
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your-gmail@gmail.com
   SMTP_PASSWORD=your-16-char-app-password
   FROM_EMAIL=your-gmail@gmail.com
   FROM_NAME=American Veterans Rebuilding
   ```

### Step 2: Update Email Service Implementation

Our current email service supports SMTP only. For Gmail API, we need to implement Gmail API support:

```bash
# Install Gmail API dependencies
go get google.golang.org/api/gmail/v1
go get golang.org/x/oauth2/google
go get google.golang.org/api/option
```

### Step 3: Test the Receipt System

1. **Restart Buffalo**: `make dev`
2. **Make a test donation** through the website
3. **Check logs** for receipt status:
   ```bash
   tail -f buffalo.log | grep receipt
   ```
4. **Verify email received** in donor's inbox

## 🔄 Implementation Options

### Option A: Service Account (Production Ready)
**Best for:** Automated server-to-server email sending
- ✅ No user interaction required
- ✅ Perfect for background receipt sending
- ✅ Scales well for production
- ⚠️ Requires Google Workspace admin setup for domain emails

### Option B: OAuth2 (User Consent)
**Best for:** Development and testing
- ✅ Works with personal Gmail accounts
- ✅ Secure token-based authentication
- ⚠️ Requires initial user consent flow
- ⚠️ Tokens need periodic refresh

### Option C: External Email Service (Simplest)
**Best for:** Quick setup without Google complexity
- ✅ SendGrid, Mailgun, Amazon SES
- ✅ Simple API key authentication
- ✅ Built for transactional emails
- 💰 Usually requires paid plan for production volume

## 📧 What the Receipt Looks Like

**Email Subject:** "Donation Receipt - American Veterans Rebuilding"

**Content Includes:**
- AVR logo and branding
- Donor name and donation amount
- Transaction ID and date
- Tax-deductible language
- Organization mission statement
- Professional formatting

## 🔍 Troubleshooting

### Check Receipt Status in Logs
```bash
# Success message:
grep "Donation receipt sent" buffalo.log

# Error message:
grep "Failed to send donation receipt" buffalo.log
```

### Common Issues
1. **"email service not configured"** - Add SMTP environment variables
2. **Authentication failed** - Check SMTP username/password
3. **Connection timeout** - Verify SMTP host and port

## 🚨 Important Notes

- **Google has changed authentication requirements** - App passwords are being deprecated
- **Modern OAuth2 or Service Accounts required** - See `GMAIL_IMPLEMENTATION_GUIDE.md` for details
- **Receipts are sent automatically** - no manual intervention needed
- **Donations work even if email fails** - robust error handling
- **Logs track all receipt attempts** - easy debugging
- **Professional templates included** - tax-compliant formatting

## ✅ Why This Is Better Than Helcim's Built-in Receipts

1. **Custom Branding** - Full AVR branding and messaging
2. **Tax Compliance** - Proper 501(c)(3) receipt language
3. **Modern Security** - OAuth2 and Service Account authentication
4. **Reliable Delivery** - Not dependent on Helcim's email system
5. **Complete Control** - Can customize templates and content
6. **Multiple Providers** - Gmail API, SMTP, or external services
7. **Redundant Sending** - Sent via both completion handler and webhooks

## 🧪 Testing Your Setup

### Test Email Sending

1. **Add test endpoint** (temporary):
   ```go
   // In actions/donations.go
   func (app *App) TestEmail(c buffalo.Context) error {
       emailService := services.NewEmailService()
       
       err := emailService.SendDonationReceipt(
           "test@example.com",
           "Test Donor",
           100.00,
           "TEST123",
           time.Now(),
       )
       
       if err != nil {
           return c.Render(200, r.String("Email failed: " + err.Error()))
       }
       
       return c.Render(200, r.String("Email sent successfully!"))
   }
   ```

2. **Add route** (temporary):
   ```go
   // In actions/app.go
   app.GET("/test-email", TestEmail)
   ```

3. **Test the endpoint**:
   ```bash
   curl http://localhost:3000/test-email
   ```

### Verify Receipt Content

1. **Check email formatting**:
   - Subject line appears correctly
   - Tax information is present
   - AVR branding displays properly
   - All donation details are accurate

2. **Test with different amounts**:
   - Small donations ($5-50)
   - Medium donations ($100-500)
   - Large donations ($1000+)

### Remove Test Code

After testing, remove the temporary test endpoint and route before deploying to production.

## 🚀 Production Deployment Checklist

### Security Verification

- [ ] **Service Account key is NOT in version control**
- [ ] **Environment variables are properly set on production server**
- [ ] **JSON key file has restricted permissions** (600)
- [ ] **FROM_EMAIL domain matches Service Account delegation**
- [ ] **All test endpoints removed from production code**

### Email Configuration

- [ ] **Gmail API enabled for production project**
- [ ] **Service Account has domain-wide delegation**
- [ ] **FROM_EMAIL is verified and authorized**
- [ ] **Email templates tested with real donation amounts**
- [ ] **Tax compliance information is accurate**

### Monitoring Setup

- [ ] **Email delivery logs are configured**
- [ ] **Error alerting is in place for failed emails**
- [ ] **Backup email method configured (if needed)**
- [ ] **Receipt delivery tracking implemented**

### Final Production Test

1. **Make a small test donation** ($1-5)
2. **Verify receipt email arrives**
3. **Check all email content for accuracy**
4. **Confirm email appears professional and branded**
5. **Test from multiple email providers** (Gmail, Outlook, etc.)

## 📞 Support and Troubleshooting

### Common Issues

### Common Issues

**"Service account key creation is disabled by organization policy"**:
- Your organization has secure-by-default policies enabled (post-May 2024)
- Follow the "Handle Secure-by-Default Organization Policies" section in Step 4
- You need Organization Policy Administrator role to disable constraints
- Alternative: Ask your admin to create the key for you

**"Permission denied" errors**:
- Verify Service Account has domain-wide delegation
- Check that FROM_EMAIL domain matches delegation scope
- Ensure Gmail API is enabled

**"Insufficient roles" or "Access denied"**:
- The minimal roles (Service Account Token Creator) should be sufficient
- Avoid adding "Project Editor" - it's a security risk
- Domain-wide delegation provides the email-sending authorization

**"Authentication failed"**:
- Verify JSON key file path and permissions
- Check environment variables are loaded correctly
- Confirm Service Account email is correct

**Emails not delivering**:
- Check spam folders
- Verify FROM_EMAIL is not blacklisted
- Review Gmail API quotas and limits
- Check email content for spam triggers

### Getting Help

1. **Check application logs** for detailed error messages
2. **Review Google Cloud Console** for API usage and errors
3. **Check organization policies** if you encounter permission issues:
   ```bash
   gcloud resource-manager org-policies list --organization=YOUR_ORG_ID
   ```
4. **Test with Gmail API Explorer**: https://developers.google.com/gmail/api/reference/rest/v1/users.messages/send
5. **Contact Google Cloud Support** for API-specific issues

### Security Best Practices Compliance

**Why this setup follows Google's 2025 security guidelines:**
- ✅ **Minimal IAM roles** - Only Service Account Token Creator, not Project Editor
- ✅ **Acknowledges secure-by-default policies** - Provides guidance for handling blocked key creation
- ✅ **Domain-wide delegation** - Proper organization email authorization
- ✅ **Environment variable configuration** - No credentials in code
- ✅ **Key file security** - Proper file permissions and exclusion from version control

**Organizations created before May 2024** may not have these constraints but should still follow these practices.

### Backup Options

If Gmail API fails, the system automatically falls back to SMTP if configured:

```bash
# Backup SMTP configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-gmail@gmail.com
SMTP_PASSWORD=your-app-password  # Only if still supported
```

**Note**: SMTP with app passwords is deprecated and may stop working at any time. Gmail API is the recommended solution.

## 🚀 Quick Start for Production

1. **Follow modern Gmail setup**: Read `GMAIL_IMPLEMENTATION_GUIDE.md`
2. **Update dependencies**: `go get` the required OAuth2 and Gmail API packages
3. **Replace email service**: Use the new `email_v2.go` implementation
4. **Configure authentication**: Service Account (recommended) or OAuth2
5. **Test thoroughly**: Verify receipt delivery and error handling

**The receipt system is production-ready with modern authentication!**
