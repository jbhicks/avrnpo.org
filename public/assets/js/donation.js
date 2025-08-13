// Donation form functionality for AVR website
// Handles donation form interactions and payment processing

(function() {
    'use strict';
    
    // Donation amount presets
    const presetAmounts = [25, 50, 100, 250, 500, 1000];
    
    // Initialize donation form
    function initializeDonationForm() {
        const donationForm = document.querySelector('#donation-form');
        if (!donationForm) return;
        
        // Handle preset amount buttons
        const presetButtons = donationForm.querySelectorAll('.amount-preset');
        const customAmountInput = donationForm.querySelector('#custom-amount');
        const selectedAmountInput = donationForm.querySelector('#selected-amount');
        
        presetButtons.forEach(button => {
            button.addEventListener('click', function(e) {
                e.preventDefault();
                
                // Clear other selections
                presetButtons.forEach(btn => btn.classList.remove('selected'));
                
                // Select this button
                this.classList.add('selected');
                
                // Set the amount
                const amount = this.getAttribute('data-amount');
                if (selectedAmountInput) {
                    selectedAmountInput.value = amount;
                }
                
                // Clear custom amount
                if (customAmountInput) {
                    customAmountInput.value = '';
                }
            });
        });
        
        // Handle custom amount input
        if (customAmountInput) {
            customAmountInput.addEventListener('input', function() {
                // Clear preset selections
                presetButtons.forEach(btn => btn.classList.remove('selected'));
                
                // Set the amount
                if (selectedAmountInput) {
                    selectedAmountInput.value = this.value;
                }
            });
        }
        
        // Handle recurring donation toggle
        const recurringToggle = donationForm.querySelector('#recurring');
        const recurringFrequency = donationForm.querySelector('#frequency-container');
        
        if (recurringToggle && recurringFrequency) {
            recurringToggle.addEventListener('change', function() {
                if (this.checked) {
                    recurringFrequency.style.display = 'block';
                    recurringFrequency.setAttribute('aria-hidden', 'false');
                } else {
                    recurringFrequency.style.display = 'none';
                    recurringFrequency.setAttribute('aria-hidden', 'true');
                }
            });
        }
    }
    
    // Validate donation form
    function validateDonationForm(form) {
        const selectedAmount = form.querySelector('#selected-amount');
        const customAmount = form.querySelector('#custom-amount');
        const email = form.querySelector('#email');
        const firstName = form.querySelector('#first_name');
        const lastName = form.querySelector('#last_name');
        
        let isValid = true;
        const errors = [];
        
        // Validate amount
        const amount = selectedAmount ? parseFloat(selectedAmount.value) : 0;
        if (!amount || amount < 1) {
            errors.push('Please select or enter a donation amount of at least $1');
            isValid = false;
        }
        
        // Validate required fields
        if (email && !email.value.trim()) {
            errors.push('Email address is required');
            isValid = false;
        }
        
        if (firstName && !firstName.value.trim()) {
            errors.push('First name is required');
            isValid = false;
        }
        
        if (lastName && !lastName.value.trim()) {
            errors.push('Last name is required');
            isValid = false;
        }
        
        // Validate email format
        if (email && email.value.trim()) {
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            if (!emailRegex.test(email.value.trim())) {
                errors.push('Please enter a valid email address');
                isValid = false;
            }
        }
        
        return { isValid, errors };
    }
    
    // Show validation errors
    function showErrors(errors) {
        const errorContainer = document.querySelector('#donation-errors');
        if (!errorContainer) return;
        
        errorContainer.innerHTML = '';
        
        if (errors.length > 0) {
            const errorList = document.createElement('ul');
            errors.forEach(error => {
                const errorItem = document.createElement('li');
                errorItem.textContent = error;
                errorList.appendChild(errorItem);
            });
            
            errorContainer.appendChild(errorList);
            errorContainer.style.display = 'block';
        } else {
            errorContainer.style.display = 'none';
        }
    }
    
    // Handle form submission
    function handleDonationSubmit(e) {
        e.preventDefault();
        
        const form = e.target;
        const validation = validateDonationForm(form);
        
        if (!validation.isValid) {
            showErrors(validation.errors);
            return false;
        }
        
        // Clear errors
        showErrors([]);
        
        // Show loading state
        const submitButton = form.querySelector('button[type="submit"]');
        if (submitButton) {
            submitButton.disabled = true;
            submitButton.textContent = 'Processing...';
        }
        
        // Submit the form (let Buffalo handle the actual submission)
        form.submit();
    }
    
    // Initialize when DOM is ready
    function initialize() {
        initializeDonationForm();
        
        // Attach form submission handler
        const donationForms = document.querySelectorAll('#donation-form, .donation-form');
        donationForms.forEach(form => {
            form.addEventListener('submit', handleDonationSubmit);
        });
    }
    
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initialize);
    } else {
        initialize();
    }
    
    // Make donation utilities available globally
    window.DonationManager = {
        validate: validateDonationForm,
        showErrors: showErrors,
        initialize: initialize
    };
})();
