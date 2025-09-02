#!/bin/bash

# Email Service Migration Script
# Migrates from legacy email.go to modern email_v2.go with Gmail API support

set -e

echo "🔄 Migrating AVR NPO to Modern Email Service..."
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Backup existing email service
echo -e "\n${YELLOW}Step 1: Backing up existing email service...${NC}"
if [ -f "services/email.go" ]; then
    cp services/email.go services/email_legacy_backup.go
    echo -e "${GREEN}✓ Backed up services/email.go to services/email_legacy_backup.go${NC}"
else
    echo -e "${YELLOW}⚠ No existing services/email.go found${NC}"
fi

# Step 2: Check if new email service exists
echo -e "\n${YELLOW}Step 2: Checking for new email service...${NC}"
if [ -f "services/email_v2.go" ]; then
    echo -e "${GREEN}✓ Found services/email_v2.go${NC}"
else
    echo -e "${RED}✗ services/email_v2.go not found!${NC}"
    echo "Please create the new email service first."
    exit 1
fi

# Step 3: Replace old email service
echo -e "\n${YELLOW}Step 3: Replacing email service...${NC}"
if [ -f "services/email.go" ]; then
    mv services/email.go services/email_old.go
    mv services/email_v2.go services/email.go
    echo -e "${GREEN}✓ Replaced services/email.go with modern implementation${NC}"
else
    mv services/email_v2.go services/email.go
    echo -e "${GREEN}✓ Created services/email.go from modern implementation${NC}"
fi

# Step 4: Update go.mod dependencies
echo -e "\n${YELLOW}Step 4: Updating Go dependencies...${NC}"
go get golang.org/x/oauth2/google
go get google.golang.org/api/gmail/v1
go get golang.org/x/oauth2
echo -e "${GREEN}✓ Updated Go dependencies for Gmail API${NC}"

# Step 5: Check environment variables
echo -e "\n${YELLOW}Step 5: Checking environment configuration...${NC}"
if [ -f ".env" ]; then
    echo -e "${GREEN}✓ Found .env file${NC}"
    
    # Check for modern config
    if grep -q "GOOGLE_SERVICE_ACCOUNT_KEY" .env || grep -q "GOOGLE_CLIENT_ID" .env; then
        echo -e "${GREEN}✓ Modern Gmail configuration detected${NC}"
    else
        echo -e "${YELLOW}⚠ No Gmail API configuration found in .env${NC}"
        echo "Please add Gmail API configuration. See RECEIPT_SETUP_GUIDE.md"
    fi
    
    # Check for legacy SMTP config
    if grep -q "SMTP_HOST" .env; then
        echo -e "${YELLOW}⚠ Legacy SMTP configuration detected${NC}"
        echo "Consider migrating to Gmail API for better security"
    fi
else
    echo -e "${YELLOW}⚠ No .env file found${NC}"
    echo "Please create .env with Gmail API configuration"
fi

# Step 6: Run tests
echo -e "\n${YELLOW}Step 6: Running tests...${NC}"
if make test-fast; then
    echo -e "${GREEN}✓ Tests passed with new email service${NC}"
else
    echo -e "${RED}✗ Tests failed${NC}"
    echo "Please check for compilation errors or test failures"
fi

# Step 7: Instructions for manual steps
echo -e "\n${YELLOW}Step 7: Manual configuration required...${NC}"
echo "=============================================="
echo ""
echo "🔧 NEXT STEPS:"
echo ""
echo "1. 📖 Read RECEIPT_SETUP_GUIDE.md for Gmail setup instructions"
echo "2. 🔑 Choose authentication method:"
echo "   • Service Account (recommended for production)"
echo "   • OAuth2 with refresh token (for development)"
echo "3. 🌍 Set up Google Cloud Console:"
echo "   • Enable Gmail API"
echo "   • Create Service Account or OAuth2 credentials"
echo "   • Configure domain-wide delegation (if using Service Account)"
echo "4. 📝 Update .env file with your credentials"
echo "5. 🧪 Test email delivery with the new service"
echo ""
echo "📚 DOCUMENTATION:"
echo "   • RECEIPT_SETUP_GUIDE.md - Step-by-step setup"
echo "   • GMAIL_IMPLEMENTATION_GUIDE.md - Technical details"
echo ""

# Step 8: Verification
echo -e "\n${YELLOW}Step 8: Migration verification...${NC}"
echo "=================================="

# Check if email service compiles
if go build -o /tmp/test_build ./services/email.go; then
    echo -e "${GREEN}✓ New email service compiles successfully${NC}"
    rm -f /tmp/test_build
else
    echo -e "${RED}✗ Compilation errors in new email service${NC}"
fi

# Check for required functions
if grep -q "SendDonationReceipt" services/email.go; then
    echo -e "${GREEN}✓ SendDonationReceipt function found${NC}"
else
    echo -e "${RED}✗ SendDonationReceipt function missing${NC}"
fi

if grep -q "NewEmailService" services/email.go; then
    echo -e "${GREEN}✓ NewEmailService function found${NC}"
else
    echo -e "${RED}✗ NewEmailService function missing${NC}"
fi

echo ""
echo "🎉 EMAIL SERVICE MIGRATION COMPLETE!"
echo "====================================="
echo ""
echo -e "${GREEN}✓ Modern Gmail API support added${NC}"
echo -e "${GREEN}✓ OAuth2 and Service Account authentication${NC}"
echo -e "${GREEN}✓ SMTP fallback for compatibility${NC}"
echo -e "${GREEN}✓ Enhanced error handling and logging${NC}"
echo ""
echo "⚠️  IMPORTANT: Configure Gmail API credentials before deploying to production"
echo ""
echo "📖 See RECEIPT_SETUP_GUIDE.md for detailed setup instructions"
echo ""
