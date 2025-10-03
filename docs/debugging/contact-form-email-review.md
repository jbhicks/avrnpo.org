# Contact Form Email Configuration Review

## Date: October 2, 2025

## Issue
Contact form submissions in production are not being sent to the `CONTACT_EMAIL` environment variable.

## Current Implementation Analysis

### 1. Email Service Initialization (`services/email.go`)

**Correct Implementation:**
```go
// Get contact email from environment with fallback
contactEmail := os.Getenv("CONTACT_EMAIL")
if contactEmail == "" {
    contactEmail = "AmericanVeteransRebuilding@avrnpo.org" // Default fallback
}

svc := &EmailService{
    // ... other fields ...
    ContactEmail: contactEmail,
    // ...
}
```

✅ **This is correct** - it reads `CONTACT_EMAIL` from environment variables.

### 2. Contact Form Handler (`actions/pages.go`)

**Flow:**
1. Validates contact form input
2. Creates `ContactFormData` struct with form values
3. Creates new `EmailService` instance: `emailService := services.NewEmailService()`
4. Calls `emailService.SendContactNotification(contactData)`

✅ **This is correct** - it creates a fresh email service instance that reads env vars.

### 3. Email Sending (`services/email.go`)

**Implementation:**
```go
func (e *EmailService) SendContactNotification(contactData ContactFormData) error {
    // Send to configured contact email
    toEmail := e.ContactEmail
    fmt.Printf("[EMAIL_SERVICE] Contact notification recipient: %s\n", toEmail)
    
    // ... generate email content ...
    
    return e.sendEmail(toEmail, subject, htmlBody, textBody)
}
```

✅ **This is correct** - it uses `e.ContactEmail` which was set from the env var.

### 4. Logging

**The code includes extensive logging:**
- `[EMAIL_SERVICE] Contact notification recipient: %s` - Shows where email is being sent
- `[EMAIL_SMTP] Recipients: Primary=%s` - Shows SMTP recipient
- `[EMAIL_SMTP] SEND SUCCESS` or `[EMAIL_SMTP] SEND FAILED` - Shows send status

## Diagnostic Steps for Production

### Step 1: Verify Environment Variable is Set

**Check Coolify/Production Environment:**
```bash
# In production environment, verify:
echo $CONTACT_EMAIL
# Should output: AmericanVeteransRebuilding@avrnpo.org (or your desired email)
```

**In Coolify Dashboard:**
1. Go to your app's Environment Variables section
2. Look for `CONTACT_EMAIL`
3. Verify it's set to the correct email address
4. Verify there are no trailing spaces or quotes

### Step 2: Check Application Logs

Look for these log lines when a contact form is submitted:

```
[EMAIL_SERVICE] Starting contact notification for submission from <name> (<email>)
[EMAIL_SERVICE] Configuration validated successfully
[EMAIL_SERVICE] Contact notification recipient: <SHOULD SHOW YOUR EMAIL HERE>
[EMAIL_SERVICE] Generated subject: New Contact Form Submission: <subject>
[EMAIL_SMTP] Recipients: Primary=<SHOULD SHOW YOUR EMAIL HERE>
```

**What to look for:**
- Is "Contact notification recipient" showing the correct email?
- Is the email being sent successfully?
- Are there any error messages?

### Step 3: Check EMAIL_ENABLED Setting

The email service checks `EMAIL_ENABLED`:
```go
if os.Getenv("GO_ENV") == "production" {
    enabledStr = "true"  // Auto-enables in production
}
```

**Verify in production:**
- `EMAIL_ENABLED` should be `true` (or not set, it defaults to true in production)
- If it's `false`, emails won't be sent

### Step 4: Verify SMTP Configuration

**Required environment variables:**
- `SMTP_HOST` - e.g., `smtp.gmail.com`
- `SMTP_PORT` - e.g., `587`
- `SMTP_USERNAME` - Your SMTP username
- `SMTP_PASSWORD` - Your SMTP password or app-specific password
- `FROM_EMAIL` - The "from" address
- `FROM_NAME` - The display name

**Check for configuration errors:**
```
[EMAIL_SERVICE] Configuration check failed - missing SMTP environment variables
```

If you see this, one or more SMTP variables are missing.

## Common Issues and Solutions

### Issue 1: CONTACT_EMAIL Not Set in Production
**Symptom:** Emails go to default `AmericanVeteransRebuilding@avrnpo.org` instead of env var
**Solution:** 
1. Add `CONTACT_EMAIL=your-email@domain.com` to production environment
2. Restart the application
3. Test contact form submission

### Issue 2: CONTACT_EMAIL Has Spaces or Quotes
**Symptom:** Email address looks wrong in logs with quotes or spaces
**Solution:**
```bash
# Wrong:
CONTACT_EMAIL=" admin@example.com "
CONTACT_EMAIL="admin@example.com"

# Correct:
CONTACT_EMAIL=admin@example.com
```

### Issue 3: Multiple Email Service Instances
**Symptom:** Old email address being used despite env var change
**Solution:**
1. The app creates a NEW `EmailService` instance for each request
2. This means env var changes require app restart
3. In Coolify: Update env var → Restart container

### Issue 4: EMAIL_ENABLED is False
**Symptom:** Logs show `[EMAIL_DISABLED] Email sending disabled`
**Solution:**
```bash
# Either set explicitly:
EMAIL_ENABLED=true

# Or rely on GO_ENV (production auto-enables):
GO_ENV=production
```

## Recommended Fix for Production

### Option A: Verify and Set in Coolify

1. **Go to Coolify Dashboard → Your App → Environment**

2. **Add/Verify these variables:**
```
CONTACT_EMAIL=your-actual-email@avrnpo.org
EMAIL_ENABLED=true
GO_ENV=production
```

3. **Restart the container**

4. **Test by submitting contact form**

5. **Check logs for:**
```
[EMAIL_SERVICE] Contact notification recipient: your-actual-email@avrnpo.org
```

### Option B: Add Explicit Logging

Add a startup log to verify what email is configured:

**In `services/email.go` after creating the service:**
```go
svc := &EmailService{
    // ... fields ...
    ContactEmail: contactEmail,
    // ...
}

// Log what was configured on startup
fmt.Printf("[EMAIL_SERVICE] Initialized with ContactEmail: %s\n", contactEmail)

return svc
```

This will show in startup logs what email is actually configured.

## Testing Checklist

- [ ] Verify `CONTACT_EMAIL` is set in production environment
- [ ] Verify `EMAIL_ENABLED` is `true` (or GO_ENV is production)
- [ ] Restart application after env var changes
- [ ] Submit test contact form
- [ ] Check application logs for recipient email address
- [ ] Verify email arrives at expected inbox
- [ ] Check spam folder if email doesn't arrive

## Expected Log Output (Success)

```
[EMAIL_SERVICE] Starting contact notification for submission from John Doe (john@example.com)
[EMAIL_SERVICE] Configuration validated successfully
[EMAIL_SERVICE] Contact notification recipient: your-email@avrnpo.org
[EMAIL_SERVICE] Generated subject: New Contact Form Submission: Test Subject
[EMAIL_SERVICE] Generated email content - HTML: 1234 bytes, Text: 567 bytes
[EMAIL_SERVICE] Initiating email send for contact notification
[EMAIL_SMTP] Starting email send operation at 2025-10-02 21:00:00
[EMAIL_SMTP] Message composed - Size: 2000 bytes, To: your-email@avrnpo.org
[EMAIL_SMTP] Recipients: Primary=your-email@avrnpo.org, BCC=none, Total=1
[EMAIL_SMTP] Attempting SMTP connection to smtp.gmail.com:587
[EMAIL_SMTP] SEND SUCCESS - Send: 500ms, Total: 550ms, Size: 2000 bytes
```

## Code Location Reference

- **Email Service:** `/services/email.go`
  - Line 60-64: CONTACT_EMAIL reading
  - Line 153-187: SendContactNotification function
  
- **Contact Handler:** `/actions/pages.go`
  - Line 273-320: ContactHandler function
  - Line 309: emailService.SendContactNotification call

- **Environment Variables:** `.env` (development), Coolify dashboard (production)
  - CONTACT_EMAIL
  - EMAIL_ENABLED
  - SMTP_* variables
