/**
 * Simple test script to verify donation API endpoint
 * Run this in browser console on the donation page
 */

async function testDonationAPI() {
    console.log('Testing donation API...');
    
    const testData = {
        amount: 50,
        frequency: 'one-time',
        donor_name: 'Test Donor',
        donor_email: 'test@example.com',
        donor_phone: '555-1234',
        comments: 'Test donation'
    };
    
    try {
        const response = await fetch('/api/donations/initialize', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(testData)
        });
        
        const result = await response.json();
        console.log('API Response:', result);
        
        if (result.success) {
            console.log('✅ Donation API working! Checkout token:', result.checkout_token);
        } else {
            console.log('❌ API returned error:', result.error);
        }
        
    } catch (error) {
        console.error('❌ Network error:', error);
    }
}

// Auto-run test
testDonationAPI();
