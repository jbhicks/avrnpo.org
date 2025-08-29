// HTMX-first donation helpers — keep behavior minimal and idempotent
(function() {
  'use strict';

  function initDonationContent(root) {
    if (!root) return;
    try {
      const form = (root.querySelector) ? root.querySelector('#donation-form') : null;
      if (!form) return;

      // Small accessibility helper: ensure numeric amount input has decimal inputmode
      const amountInput = form.querySelector('#amount');
      if (amountInput) amountInput.setAttribute('inputmode', 'decimal');
    } catch (e) {
      // swallow errors — this is a tiny helper
    }
  }

  // Run on initial load and after HTMX swaps
  initDonationContent(document);
  if (typeof htmx !== 'undefined') {
    htmx.onLoad(function(elt) { initDonationContent(elt); });
  }
})();
