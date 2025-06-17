/**
 * Donation System Frontend
 * Handles form interactions and basic amount selection
 */

// Prevent duplicate class declarations
if (typeof window.DonationSystem === 'undefined') {
    class DonationSystem {
        constructor() {
            this.currentAmount = null;
            this.donationType = 'one-time';
            this.isProcessing = false;
            
            this.init();
        }

    init() {
        this.bindEvents();
        console.log('Donation system initialized');
        
        // Auto-fill test data in development mode
        if (this.isDevelopmentMode()) {
            this.setupDevelopmentHelpers();
        }
    }

    isDevelopmentMode() {
        // Check if development notice is present
        return document.getElementById('dev-notice') !== null;
    }

    setupDevelopmentHelpers() {
        // Auto-select $25 amount for quick testing
        const firstAmountBtn = document.querySelector('.amount-btn[data-amount="25"]');
        if (firstAmountBtn) {
            setTimeout(() => {
                firstAmountBtn.click();
            }, 100);
        }
        
        console.log('Development helpers enabled - test data auto-fill available');
    }

    bindEvents() {
        // Amount button selection
        document.querySelectorAll('.amount-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.preventDefault();
                console.log('Amount button clicked:', btn.dataset.amount);
                this.selectAmount(btn.dataset.amount);
                this.updateAmountDisplay(btn);
            });
        });

        // Custom amount input
        const customAmountInput = document.getElementById('custom-amount');
        if (customAmountInput) {
            customAmountInput.addEventListener('input', (e) => {
                console.log('Custom amount entered:', e.target.value);
                this.selectCustomAmount(e.target.value);
                this.clearAmountButtons();
            });
        }

        // Donation frequency
        document.querySelectorAll('input[name="frequency"]').forEach(radio => {
            radio.addEventListener('change', (e) => {
                this.donationType = e.target.value;
                this.updateDonateButton();
            });
        });

        // Main donate button
        const donateBtn = document.querySelector('.donation-submit');
        if (donateBtn) {
            donateBtn.addEventListener('click', (e) => {
                e.preventDefault();
                this.processDonation();
            });
        }

        console.log('Event listeners bound');
    }

    selectAmount(amount) {
        this.currentAmount = parseFloat(amount);
        console.log('Amount selected:', this.currentAmount);
        
        // Clear custom amount input
        const customInput = document.getElementById('custom-amount');
        if (customInput) {
            customInput.value = '';
        }
        
        this.updateDonateButton();
    }

    selectCustomAmount(amount) {
        const numAmount = parseFloat(amount);
        this.currentAmount = numAmount > 0 ? numAmount : null;
        console.log('Custom amount selected:', this.currentAmount);
        this.updateDonateButton();
    }

    updateAmountDisplay(selectedBtn) {
        // Remove active class from all buttons
        document.querySelectorAll('.amount-btn').forEach(btn => {
            btn.classList.remove('selected');
        });
        
        // Add active class to selected button
        selectedBtn.classList.add('selected');
        console.log('Amount display updated, selected button highlighted');
    }

    clearAmountButtons() {
        document.querySelectorAll('.amount-btn').forEach(btn => {
            btn.classList.remove('selected');
        });
        console.log('Amount buttons cleared');
    }

    updateDonateButton() {
        const donateBtn = document.querySelector('.donation-submit');
        if (!donateBtn) return;

        if (this.currentAmount && this.currentAmount > 0) {
            const frequencyText = this.donationType === 'monthly' ? 'Monthly' : '';
            donateBtn.textContent = `${frequencyText} Donate $${this.currentAmount.toFixed(2)}`.trim();
            donateBtn.disabled = false;
            donateBtn.classList.remove('secondary');
            donateBtn.classList.add('contrast');
        } else {
            donateBtn.textContent = 'Select Amount';
            donateBtn.disabled = true;
            donateBtn.classList.remove('contrast');
            donateBtn.classList.add('secondary');
        }
        console.log('Donate button updated:', donateBtn.textContent);
    }    processDonation() {
        if (!this.currentAmount || this.currentAmount <= 0) {
            alert('Please select a donation amount');
            return;
        }

        if (this.isProcessing) {
            return;
        }

        // Basic form validation
        const requiredFields = ['donor-name', 'donor-email'];
        for (const fieldId of requiredFields) {
            const field = document.getElementById(fieldId);
            if (!field || !field.value.trim()) {
                alert(`Please fill out the ${fieldId.replace('-', ' ')} field`);
                if (field) field.focus();
                return;
            }
        }

        // Email validation
        const emailField = document.getElementById('donor-email');
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(emailField.value.trim())) {
            alert('Please enter a valid email address');
            emailField.focus();
            return;
        }

        console.log('Processing donation:', {
            amount: this.currentAmount,
            type: this.donationType
        });

        this.isProcessing = true;
        this.updateProcessingState(true);

        // Collect form data
        const formData = {
            amount: this.currentAmount.toString(),
            custom_amount: '',
            donation_type: this.donationType,
            donor_name: document.getElementById('donor-name').value.trim(),
            donor_email: document.getElementById('donor-email').value.trim(),
            donor_phone: document.getElementById('donor-phone').value.trim(),
            address_line1: document.getElementById('address-line1').value.trim(),
            city: document.getElementById('city').value.trim(),
            state: document.getElementById('state').value.trim(),
            zip: document.getElementById('zip').value.trim(),
            comments: document.getElementById('comments').value.trim()
        };

        // Initialize payment with Helcim
        this.initializePayment(formData);
    }

    updateProcessingState(isProcessing) {
        const donateBtn = document.querySelector('.donation-submit');
        if (donateBtn) {
            if (isProcessing) {
                donateBtn.textContent = 'Processing...';
                donateBtn.disabled = true;
            } else {
                this.updateDonateButton();
            }
        }
    }

    async initializePayment(formData) {
        try {
            console.log('Initializing payment with data:', formData);
            
            const response = await fetch('/api/donations/initialize', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(formData)
            });

            const result = await response.json();
            console.log('Initialize response:', result);

            if (!response.ok) {
                throw new Error(result.error || 'Failed to initialize payment');
            }

            if (result.success && result.checkoutToken) {
                // Load Helcim JS and show payment modal
                this.loadHelcimJS(() => {
                    this.showHelcimModal(result.checkoutToken, result.donationId);
                });
            } else {
                throw new Error('Invalid response from payment processor');
            }

        } catch (error) {
            console.error('Payment initialization error:', error);
            alert('Error initializing payment: ' + error.message);
            this.isProcessing = false;
            this.updateProcessingState(false);
        }
    }    loadHelcimJS(callback) {
        // Check if Helcim JS is already loaded
        if (window.HelcimPay) {
            callback();
            return;
        }
        
        // Load Helcim JS dynamically
        const script = document.createElement('script');
        script.src = '/js/helcim-pay.min.js';  // Use local copy
        script.onload = callback;
        script.onerror = () => {
            console.error('Failed to load Helcim JS');
            alert('Error loading payment processor. Please try again.');
            this.isProcessing = false;
            this.updateProcessingState(false);
        };
        document.head.appendChild(script);
    }

    showHelcimModal(checkoutToken, donationId) {
        try {
            console.log('Showing Helcim modal with token:', checkoutToken);
            
            const helcim = new window.HelcimPay();
            
            // Use callback-based API for local Helcim implementation
            helcim.startPayment({
                checkoutToken: checkoutToken,
                amount: this.amount,
                onSuccess: (response) => {
                    console.log('Payment successful:', response);
                    this.handlePaymentSuccess(response, donationId);
                },
                onError: (error) => {
                    console.error('Payment error:', error);
                    this.handlePaymentError(error);
                },
                onCancel: () => {
                    console.log('Payment cancelled');
                    this.handlePaymentCancelled();
                }
            });

        } catch (error) {
            console.error('Error showing Helcim modal:', error);
            alert('Error loading payment form. Please try again.');
            this.isProcessing = false;
            this.updateProcessingState(false);
        }
    }

    handlePaymentSuccess(response, donationId) {
        console.log('Payment completed successfully');
        
        // Update our backend with the transaction details
        fetch(`/api/donations/${donationId}/complete`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                transactionId: response.transactionId,
                status: 'APPROVED'
            })
        }).then(response => response.json())
        .then(result => {
            console.log('Payment completion recorded:', result);
            // Redirect to success page
            window.location.href = '/donate/success';
        })
        .catch(error => {
            console.error('Error recording payment completion:', error);
            // Still show success since payment went through
            window.location.href = '/donate/success';
        });
    }

    handlePaymentError(error) {
        console.error('Payment failed:', error);
        alert('Payment failed: ' + (error.message || 'Unknown error'));
        this.isProcessing = false;
        this.updateProcessingState(false);
    }

    handlePaymentCancelled() {
        console.log('Payment was cancelled by user');
        this.isProcessing = false;
        this.updateProcessingState(false);
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM loaded, checking for donation form...');
    // Only initialize on donation page
    if (document.querySelector('.donation-form')) {
        console.log('Donation form found, initializing system...');
        new DonationSystem();
    } else {
        console.log('No donation form found on this page');
    }
});

// Make DonationSystem available globally
window.DonationSystem = DonationSystem;

// Global helper functions for development mode
window.copyTestCard = function(cardNumber) {
    // Format the card number with spaces for display
    const formatted = cardNumber.replace(/(.{4})/g, '$1 ').trim();
    
    // Copy unformatted number to clipboard
    navigator.clipboard.writeText(cardNumber).then(() => {
        showToast(`Copied: ${formatted}`, 'success');
    }).catch(() => {
        // Fallback for older browsers
        alert(`Test card: ${formatted}\nCopy this number manually.`);
    });
};

window.fillTestData = function() {
    // Fill donor information with test data
    const fields = {
        'donor-name': 'John Test Donor',
        'donor-email': 'test@example.com',
        'donor-phone': '555-123-4567',
        'address-line1': '123 Test Street',
        'city': 'Test City',
        'state': 'LA',
        'zip': '70001'
    };
    
    Object.entries(fields).forEach(([id, value]) => {
        const field = document.getElementById(id);
        if (field) {
            field.value = value;
            // Trigger input event to update any listeners
            field.dispatchEvent(new Event('input', { bubbles: true }));
        }
    });
    
    showToast('Test donor information filled', 'info');
    console.log('Test donor data filled');
};

// Simple toast notification system
function showToast(message, type = 'info') {
    // Remove existing toasts
    document.querySelectorAll('.toast').forEach(toast => toast.remove());
    
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: var(--pico-card-background-color);
        border: 1px solid var(--pico-muted-border-color);
        border-radius: var(--pico-border-radius);
        padding: 0.75rem 1rem;
        box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
        z-index: 9999;
        max-width: 300px;
        font-size: 0.9rem;
        animation: slideIn 0.3s ease-out;
    `;
    
    // Type-specific styling
    if (type === 'success') {
        toast.style.borderColor = 'var(--pico-primary)';
        toast.style.color = 'var(--pico-primary)';
    } else if (type === 'error') {
        toast.style.borderColor = 'var(--pico-contrast)';
        toast.style.color = 'var(--pico-contrast)';
    } else {
        toast.style.borderColor = '#3b82f6';
        toast.style.color = '#3b82f6';
    }
    
    toast.textContent = message;
    document.body.appendChild(toast);
    
    // Auto-remove after 3 seconds
    setTimeout(() => {
        if (toast.parentNode) {
            toast.remove();
        }
    }, 3000);
}

} // Close the conditional check

console.log('Donation.js loaded');
