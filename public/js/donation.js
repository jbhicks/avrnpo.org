/**
 * Donation System Frontend
 * Handles HelcimPay.js integration for secure payment processing
 */

class DonationSystem {
    constructor() {
        this.currentAmount = null;
        this.donationType = 'one-time';
        this.isProcessing = false;
        
        this.init();
    }

    init() {
        this.bindEvents();        this.loadHelcimPayJS().catch(error => {
            console.error('Failed to initialize HelcimPay:', error);
            console.log('Donation system will still work with mock payment processing');
        });
    }

    bindEvents() {
        // Amount button selection
        document.querySelectorAll('.amount-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.preventDefault();
                this.selectAmount(btn.dataset.amount);
                this.updateAmountDisplay(btn);
            });
        });

        // Custom amount input
        const customAmountInput = document.getElementById('custom-amount');
        if (customAmountInput) {
            customAmountInput.addEventListener('input', (e) => {
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
    }

    selectAmount(amount) {
        this.currentAmount = parseFloat(amount);
        
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
        this.updateDonateButton();
    }

    updateAmountDisplay(selectedBtn) {
        // Remove active class from all buttons
        document.querySelectorAll('.amount-btn').forEach(btn => {
            btn.classList.remove('selected');
        });
        
        // Add active class to selected button
        selectedBtn.classList.add('selected');
    }

    clearAmountButtons() {
        document.querySelectorAll('.amount-btn').forEach(btn => {
            btn.classList.remove('selected');
        });
    }

    updateDonateButton() {
        const donateBtn = document.querySelector('.donation-submit');
        if (!donateBtn) return;

        if (this.currentAmount && this.currentAmount > 0) {
            const frequencyText = this.donationType === 'monthly' ? 'Monthly' : '';
            donateBtn.textContent = `${frequencyText} Donate $${this.currentAmount.toFixed(2)}`.trim();
            donateBtn.disabled = false;
        } else {
            donateBtn.textContent = 'Select Amount';
            donateBtn.disabled = true;
        }
    }

    async processDonation() {
        if (this.isProcessing || !this.currentAmount || this.currentAmount <= 0) {
            return;
        }

        this.isProcessing = true;
        this.showLoading();

        try {
            // Collect donation data
            const donationData = this.collectDonationData();
            
            // Validate required fields
            if (!this.validateDonationData(donationData)) {
                this.showError('Please fill in all required fields');
                return;
            }

            // Initialize donation with backend
            const response = await this.initializeDonation(donationData);
            
            if (response.success) {
                // Launch HelcimPay modal
                await this.launchHelcimPay(response.checkoutToken, response.donationId);
            } else {
                this.showError(response.error || 'Failed to initialize donation');
            }
        } catch (error) {
            console.error('Donation error:', error);
            this.showError('Payment system is currently unavailable. Please try again later.');
        } finally {
            this.isProcessing = false;
            this.hideLoading();
        }
    }

    collectDonationData() {
        const customAmount = document.getElementById('custom-amount')?.value || '';
        
        return {
            amount: this.currentAmount <= 100 ? this.currentAmount.toString() : 'custom',
            custom_amount: customAmount,
            donation_type: this.donationType,
            donor_name: document.getElementById('donor-name')?.value || '',
            donor_email: document.getElementById('donor-email')?.value || '',
            donor_phone: document.getElementById('donor-phone')?.value || '',
            address_line1: document.getElementById('address-line1')?.value || '',
            city: document.getElementById('city')?.value || '',            state: document.getElementById('state')?.value || '',
            zip: document.getElementById('zip')?.value || '',
            comments: document.getElementById('comments')?.value || ''
        };
    }

    validateDonationData(data) {
        // Check required fields
        if (!this.currentAmount || this.currentAmount <= 0) {
            this.showError('Please select a donation amount');
            return false;
        }
        
        if (!data.donor_name.trim()) {
            this.showError('Please enter your full name');
            this.focusField('donor-name');
            return false;
        }
        
        if (!data.donor_email.trim()) {
            this.showError('Please enter your email address');
            this.focusField('donor-email');
            return false;
        }
        
        // Basic email validation
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(data.donor_email.trim())) {
            this.showError('Please enter a valid email address');
            this.focusField('donor-email');
            return false;
        }
        
        return true;
    }

    async initializeDonation(donationData) {
        const response = await fetch('/api/donations/initialize', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(donationData)
        });        return await response.json();
    }

    async launchHelcimPay(checkoutToken, donationId) {
        // Ensure HelcimPay is loaded
        if (!window.HelcimPay) {
            console.log('HelcimPay not loaded, attempting to load...');
            try {
                await this.loadHelcimPayJS();
            } catch (error) {
                throw new Error('Failed to load HelcimPay.js: ' + error.message);
            }
        }

        if (!window.HelcimPay) {
            throw new Error('HelcimPay.js is not available');
        }

        return new Promise((resolve, reject) => {
            try {
                const helcim = new window.HelcimPay();
                  helcim.startPayment({
                    checkoutToken: checkoutToken,
                    amount: this.currentAmount,
                    onSuccess: (result) => {
                        this.handlePaymentSuccess(result, donationId);
                        resolve(result);
                    },
                    onError: (error) => {
                        this.handlePaymentError(error);
                        reject(error);
                    },
                    onCancel: () => {
                        this.handlePaymentCancel();
                        resolve({ cancelled: true });
                    }
                });            } catch (error) {
                console.error('Error initializing HelcimPay:', error);
                reject(error);
            }
        });
    }

    async handlePaymentSuccess(result, donationId) {
        try {
            // Complete donation on backend
            const response = await fetch(`/api/donations/${donationId}/complete`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    transactionId: result.transactionId,
                    status: 'APPROVED' // Use Helcim's standard status
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            
            if (data.success) {
                this.showSuccess('Thank you for your donation! You should receive a receipt email shortly.');
                this.resetForm();
                  // Redirect to success page after a short delay
                setTimeout(() => {
                    window.location.href = '/donate/success';
                }, 2000);
            } else {
                this.showError('Payment successful, but there was an issue processing your donation. Please contact us.');
            }
        } catch (error) {
            console.error('Error completing donation:', error);
            this.showError('Payment was processed, but there was an issue with our system. Please contact us to confirm your donation.');        }
    }

    handlePaymentError(error) {
        console.error('Payment error:', error);
        
        let errorMessage = 'Payment failed. Please try again or contact us for assistance.';
        
        // Handle specific error types
        if (error && error.message) {
            if (error.message.includes('declined')) {
                errorMessage = 'Your payment was declined. Please check your card details and try again.';
            } else if (error.message.includes('network') || error.message.includes('timeout')) {
                errorMessage = 'Network error occurred. Please check your connection and try again.';
            }
        }
        
        this.showError(errorMessage);
        
        // Redirect to failure page after a short delay
        setTimeout(() => {
            window.location.href = '/donation/failed';
        }, 3000);
    }

    handlePaymentCancel() {
        console.log('Payment cancelled by user');        this.showInfo('Payment cancelled. You can try again anytime.');
    }

    loadHelcimPayJS() {
        // Load HelcimPay.js if not already loaded
        if (window.HelcimPay) {
            console.log('HelcimPay.js already loaded');
            return Promise.resolve();
        }

        return new Promise((resolve, reject) => {            const script = document.createElement('script');
            // Use local minified library instead of CDN (following template philosophy)
            script.src = '/js/helcim-pay.min.js';
            script.async = true;            script.onload = () => {
                console.log('HelcimPay.js loaded successfully from local source');
                console.log('Development mode: Using real Helcim API with test cards');
                resolve();
            };            script.onerror = () => {
                console.error('Failed to load local HelcimPay.js library');
                reject(new Error('Could not load HelcimPay.js from local source'));
            };
            
            document.head.appendChild(script);
        });
    }    showLoading() {
        const donateBtn = document.querySelector('.donation-submit');
        if (donateBtn) {
            donateBtn.setAttribute('aria-busy', 'true');
            donateBtn.textContent = 'Processing...';
            donateBtn.disabled = true;
        }
    }

    hideLoading() {
        const donateBtn = document.querySelector('.donation-submit');
        if (donateBtn) {
            donateBtn.setAttribute('aria-busy', 'false');
            donateBtn.disabled = false;
            this.updateDonateButton();
        }
    }

    showSuccess(message) {
        this.showMessage(message, 'success');
    }

    showError(message) {
        this.showMessage(message, 'error');
    }

    showInfo(message) {
        this.showMessage(message, 'info');
    }

    showMessage(message, type = 'info') {
        // Remove existing messages
        const existingMessage = document.querySelector('.donation-message');
        if (existingMessage) {
            existingMessage.remove();
        }

        // Create message element
        const messageEl = document.createElement('div');
        messageEl.className = `donation-message ${type}`;
        messageEl.setAttribute('role', 'alert');
        messageEl.textContent = message;

        // Insert message at top of donation form
        const donationCard = document.querySelector('.donation-card');
        if (donationCard) {
            donationCard.insertBefore(messageEl, donationCard.firstChild);
        }

        // Auto-remove after 10 seconds for non-error messages
        if (type !== 'error') {
            setTimeout(() => {
                if (messageEl.parentNode) {
                    messageEl.remove();
                }
            }, 10000);
        }
    }

    resetForm() {
        // Reset amount selection
        this.currentAmount = null;
        this.clearAmountButtons();
        
        // Clear custom amount
        const customInput = document.getElementById('custom-amount');
        if (customInput) {
            customInput.value = '';
        }
        
        // Reset frequency to one-time
        const oneTimeRadio = document.querySelector('input[name="frequency"][value="one-time"]');
        if (oneTimeRadio) {
            oneTimeRadio.checked = true;
            this.donationType = 'one-time';
        }
          // Update button
        this.updateDonateButton();
    }
    
    focusField(fieldId) {
        const field = document.getElementById(fieldId);
        if (field) {
            field.focus();
            field.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Only initialize on donation page
    if (document.querySelector('.donation-form')) {
        new DonationSystem();
    }
});
