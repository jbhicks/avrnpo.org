// Simple donation form enhancements
document.addEventListener('DOMContentLoaded', function() {
  
  // Amount selection functionality
  window.selectAmount = function(amount) {
    // Update hidden field
    const amountField = document.getElementById('selected-amount');
    if (amountField) {
      amountField.value = amount;
    }
    
    // Update button states
    const buttons = document.querySelectorAll('.amount-btn');
    buttons.forEach(btn => {
      btn.classList.remove('active');
      btn.setAttribute('aria-pressed', 'false');
    });
    
    // Activate selected button
    const selectedBtn = document.querySelector(`[data-amount="${amount}"]`);
    if (selectedBtn) {
      selectedBtn.classList.add('active');
      selectedBtn.setAttribute('aria-pressed', 'true');
    }
    
    // Clear custom amount
    const customAmount = document.getElementById('custom_amount');
    if (customAmount && customAmount.value !== amount) {
      customAmount.value = '';
    }
    
    updateSubmitButton();
  };
  
  // Custom amount functionality
  window.selectCustomAmount = function(amount) {
    // Update hidden field
    const amountField = document.getElementById('selected-amount');
    if (amountField) {
      amountField.value = amount;
    }
    
    // Clear preset button selection
    const buttons = document.querySelectorAll('.amount-btn');
    buttons.forEach(btn => {
      btn.classList.remove('active');
      btn.setAttribute('aria-pressed', 'false');
    });
    
    updateSubmitButton();
  };
  
  // Donation type functionality
  window.updateDonationType = function(type) {
    updateSubmitButton();
  };
  
  // Update submit button text based on selection
  function updateSubmitButton() {
    const amountField = document.getElementById('selected-amount');
    const donationType = document.querySelector('input[name="donation_type"]:checked');
    const submitText = document.getElementById('submit-text');
    
    if (!submitText) return;
    
    const amount = amountField ? amountField.value : '';
    const type = donationType ? donationType.value : 'one-time';
    
    if (amount && parseFloat(amount) > 0) {
      const formattedAmount = '$' + parseFloat(amount).toFixed(2);
      if (type === 'monthly') {
        submitText.textContent = `Donate ${formattedAmount}/month`;
      } else {
        submitText.textContent = `Donate ${formattedAmount}`;
      }
    } else {
      submitText.textContent = 'Donate Now';
    }
  }
  
  // Initialize on page load
  updateSubmitButton();
});