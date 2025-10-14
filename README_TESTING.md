# E2E Testing Guide for AVR NPO

## Modern Go Browser Testing Options

### 1. **Chromedp** (Recommended)
- Pure Go implementation of Chrome DevTools Protocol
- No external dependencies (uses system Chrome/Chromium)
- Headless and headed modes
- Fast and lightweight
- **Best for:** CI/CD pipelines and local testing

### 2. **Rod**
- Go library built on Chrome DevTools Protocol
- Higher-level API than chromedp
- Better debugging support
- Automatic browser download
- **Best for:** Complex UI automation

### 3. **Playwright-Go**
- Go bindings for Microsoft Playwright
- Multi-browser support (Chrome, Firefox, Safari)
- Powerful selectors and auto-waiting
- Excellent for visual testing
- **Best for:** Cross-browser compatibility testing

## Setup Instructions

### Install Chromedp

```bash
go get github.com/chromedp/chromedp
```

No additional dependencies needed - uses your system Chrome/Chromium.

### Install Chrome (if not present)

**Ubuntu/Debian:**
```bash
wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
sudo sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list'
sudo apt update
sudo apt install google-chrome-stable
```

**macOS:**
```bash
brew install --cask google-chrome
```

## Running E2E Tests

### 1. Start the Application

```bash
# Terminal 1: Start the server
./avrnpo serve

# Or with custom port
PB_PORT=8090 ./avrnpo serve
```

### 2. Seed Test Data (Optional)

```bash
go run seed_test_data.go
```

### 3. Run E2E Tests

```bash
# Run all E2E tests
E2E_TESTS=1 go test -v -run TestE2E

# Run specific test
E2E_TESTS=1 go test -v -run TestE2E_FullUserJourney

# Run admin workflow tests (requires admin credentials)
PB_ADMIN_EMAIL=admin@example.com PB_ADMIN_PASSWORD=password123 E2E_TESTS=1 go test -v -run TestE2E_AdminWorkflow

# Run with visible browser (no headless)
E2E_TESTS=1 HEADLESS=0 go test -v -run TestE2E
```

### 4. Run API Tests Only (No Browser Required)

```bash
E2E_TESTS=1 go test -v -run TestE2E_APIEndpoints
```

## Test Coverage

### Public Pages
- ✓ Homepage with recent blog posts
- ✓ Blog listing page
- ✓ Individual blog post pages
- ✓ Donation page
- ✓ Contact form submission
- ✓ About page
- ✓ Team page
- ✓ Projects page

### Admin Functionality
- ✓ Admin login
- ✓ Post list view
- ✓ Create new post
- ✓ Edit existing post
- ✓ Delete post
- ✓ Publish/unpublish posts

### API Endpoints
- ✓ All public routes return 200
- ✓ Navigation state API
- ✓ PocketBase admin interface

## MCP Server Integration (Future Enhancement)

Model Context Protocol (MCP) servers could enhance testing in several ways:

### 1. **Database MCP Server**
```json
{
  "mcpServers": {
    "sqlite": {
      "command": "mcp-server-sqlite",
      "args": ["./pb_data/data.db"]
    }
  }
}
```
**Benefits:**
- Query test data directly during test development
- Verify database state without manual SQL
- AI-assisted test data generation

### 2. **Browser Automation MCP Server**
```json
{
  "mcpServers": {
    "puppeteer": {
      "command": "mcp-server-puppeteer",
      "args": ["--headless"]
    }
  }
}
```
**Benefits:**
- AI can help write complex browser automation
- Natural language to browser actions
- Screenshot and debugging assistance

### 3. **Filesystem MCP Server**
```json
{
  "mcpServers": {
    "filesystem": {
      "command": "mcp-server-filesystem",
      "args": ["/home/josh/avrnpo.org"]
    }
  }
}
```
**Benefits:**
- Read test results and logs
- Generate test reports
- Analyze code coverage

### How to Use MCP for Testing

1. **Install MCP servers:**
```bash
npm install -g @modelcontextprotocol/server-sqlite
npm install -g @modelcontextprotocol/server-puppeteer
```

2. **Configure Claude Desktop** (`~/Library/Application Support/Claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "sqlite": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sqlite", "./pb_data/data.db"]
    },
    "puppeteer": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-puppeteer"]
    }
  }
}
```

3. **Ask Claude to:**
- "Query the posts table and show me published posts"
- "Write a browser test that fills out the contact form"
- "Check if there are any donations in the database"

## CI/CD Integration

### GitHub Actions Example

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Install Chrome
        run: |
          wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
          sudo sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list'
          sudo apt update
          sudo apt install -y google-chrome-stable
      
      - name: Build application
        run: go build -o avrnpo ./main.go
      
      - name: Start server in background
        run: |
          ./avrnpo serve &
          sleep 5
        env:
          PB_ADMIN_EMAIL: admin@test.com
          PB_ADMIN_PASSWORD: testpass123
      
      - name: Run E2E tests
        run: E2E_TESTS=1 go test -v -run TestE2E
```

## Debugging Tips

### 1. Run with Visible Browser
```bash
HEADLESS=0 E2E_TESTS=1 go test -v -run TestE2E_FullUserJourney
```

### 2. Add Screenshots to Tests
```go
var buf []byte
chromedp.Run(ctx,
    chromedp.Navigate(testBaseURL),
    chromedp.CaptureScreenshot(&buf),
)
os.WriteFile("screenshot.png", buf, 0644)
```

### 3. Enable Verbose Logging
```go
opts := append(chromedp.DefaultExecAllocatorOptions[:],
    chromedp.Flag("enable-logging", true),
    chromedp.Flag("v", 1),
)
```

## Performance Testing

For load testing, use standard Go tools:

```bash
# Install hey (HTTP load generator)
go install github.com/rakyll/hey@latest

# Test homepage
hey -n 1000 -c 10 http://localhost:8090/

# Test API endpoint
hey -n 500 -c 5 http://localhost:8090/api/nav-state
```

## Manual Testing Checklist

- [ ] Homepage loads with 3 recent posts
- [ ] Blog listing shows all published posts
- [ ] Individual blog posts render correctly
- [ ] Contact form submission creates database record
- [ ] Contact form sends email notification
- [ ] Donation page loads Helcim widget
- [ ] One-time donation processes successfully
- [ ] Monthly donation creates subscription
- [ ] Admin login with correct credentials
- [ ] Admin can create new posts
- [ ] Admin can edit existing posts
- [ ] Admin can delete posts
- [ ] Admin can publish/unpublish posts
- [ ] Static files serve correctly
- [ ] Navigation works on all pages
- [ ] Mobile responsive design works
- [ ] Forms have CSRF protection (HTMX)
- [ ] Email receipts sent for donations
- [ ] PocketBase admin UI accessible at `/_/`

## Next Steps

1. **Add the chromedp dependency:**
   ```bash
   go get github.com/chromedp/chromedp
   ```

2. **Create the E2E test file** (see `e2e_test.go`)

3. **Set up CI/CD pipeline** with the GitHub Actions workflow above

4. **Consider MCP integration** for AI-assisted testing workflows
