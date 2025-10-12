@echo off
REM Start Brave with remote debugging enabled for Chrome DevTools Protocol
REM This allows WSL2 tools to connect to the browser

echo Starting Brave with remote debugging on port 9222...
echo.
echo You can now connect from WSL2 using Chrome DevTools MCP
echo Navigate to: http://localhost:8090
echo.

"C:\Users\joshu\AppData\Local\BraveSoftware\Brave-Browser\Application\brave.exe" --remote-debugging-port=9222 --remote-debugging-address=0.0.0.0 "http://localhost:8090"
