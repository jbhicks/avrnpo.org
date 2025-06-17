/**
 * Donation System Frontend
 * Handles form interactions and basic amount selection
 */

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
    }

    processDonation() {
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

        console.log('Processing donation:', {
            amount: this.currentAmount,
            type: this.donationType
        });

        // For now, just show an alert - Helcim integration can be added later
        alert(`Donation processing would start here:\nAmount: $${this.currentAmount}\nType: ${this.donationType}\n\nHelcim integration coming next!`);
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

console.log('Donation.js loaded');
