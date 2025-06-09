// Simple test to validate donation.js syntax
const fs = require('fs');

try {
    const jsContent = fs.readFileSync('./public/js/donation.js', 'utf8');
    // Try to parse it as a Function to check syntax
    new Function(jsContent);
    console.log('✅ JavaScript syntax is valid');
} catch (error) {
    console.error('❌ JavaScript syntax error:', error.message);
    console.error('Line:', error.lineNumber || 'Unknown');
}
