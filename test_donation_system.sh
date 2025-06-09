#!/bin/bash

# Test script for donation system
echo "🧪 Testing AVR NPO Donation System"
echo "=================================="

# Test 1: Check if Buffalo is running
echo "1. Checking if Buffalo server is running..."
if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Buffalo server is running on port 3000"
else
    echo "❌ Buffalo server is not responding"
    exit 1
fi

# Test 2: Test donation page loads
echo "2. Testing donation page..."
if curl -s http://localhost:3000/donate | grep -q "donation-form"; then
    echo "✅ Donation page loads and contains donation form"
else
    echo "❌ Donation page not working properly"
fi

# Test 3: Test donation API endpoint
echo "3. Testing donation API endpoint..."
RESPONSE=$(curl -s -X POST http://localhost:3000/api/donations/initialize \
    -H "Content-Type: application/json" \
    -d '{
        "amount": 50.00,
        "frequency": "one-time",
        "donor_name": "Test Donor",
        "donor_email": "test@example.com"
    }')

if echo "$RESPONSE" | grep -q "checkoutToken"; then
    echo "✅ Donation API endpoint working"
else
    echo "❌ Donation API endpoint failed"
    echo "Response: $RESPONSE"
fi

# Test 4: Test success page
echo "4. Testing donation success page..."
if curl -s http://localhost:3000/donate/success | grep -q "Thank you"; then
    echo "✅ Success page loads correctly"
else
    echo "❌ Success page not working"
fi

# Test 5: Test failure page
echo "5. Testing donation failure page..."
if curl -s http://localhost:3000/donate/failed | grep -q "not completed"; then
    echo "✅ Failure page loads correctly"
else
    echo "❌ Failure page not working"
fi

echo ""
echo "🎉 Donation system test completed!"
echo "Visit http://localhost:3000/donate to test the full flow manually"
