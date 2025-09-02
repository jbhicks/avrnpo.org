# Dashlane Autofill Compatibility Guide

## Overview

This guide explains how the AVR donation form has been optimized for compatibility with Dashlane and other password managers to ensure smooth autofill functionality.

## Key Changes Made

### 1. **Form Field Naming Consistency**
- **Changed from hyphens to underscores** in field names
- **Added `name` attributes** that match `id` attributes
- **Ensured both `id` and `name` are present** on all form fields

**Before:**
```html
<input type="text" id="first-name" name="first-name" autocomplete="given-name">
```

**After:**
```html
<input type="text" id="first_name" name="first_name" autocomplete="billing given-name">
```

### 2. **Enhanced Autocomplete Attributes**

#### Billing Context for Donations
All address and contact fields now use the `billing` context to help password managers group related information:

```html
<input type="text" id="first_name" name="first_name" autocomplete="billing given-name">
<input type="text" id="last_name" name="last_name" autocomplete="billing family-name">
<input type="email" id="donor_email" name="donor_email" autocomplete="billing email">
<input type="tel" id="donor_phone" name="donor_phone" autocomplete="billing tel">
<input type="text" id="address_line1" name="address_line1" autocomplete="billing address-line1">
<input type="text" id="address_line2" name="address_line2" autocomplete="billing address-line2">
<input type="text" id="city" name="city" autocomplete="billing address-level2">
<input type="text" id="state" name="state" autocomplete="billing address-level1">
<input type="text" id="zip_code" name="zip_code" autocomplete="billing postal-code">
```

#### Transaction Amount
```html
<input type="number" id="custom_amount" name="custom_amount" autocomplete="transaction-amount">
```

#### Comments Field
```html
<textarea id="comments" name="comments" autocomplete="off">
```

### 3. **Required Field Attributes**
All required fields include both HTML5 validation and ARIA attributes:

```html
<input type="text" 
       id="first_name" 
       name="first_name" 
       autocomplete="billing given-name" 
       aria-required="true" 
       required>
```

### 4. **Form Structure Optimization**

#### Form Element
```html
<form id="donation-form" 
      method="post" 
      action="/api/donations/initialize" 
      autocomplete="on" 
      novalidate>
```

- **`autocomplete="on"`** - Explicitly enables autofill for the form
- **`novalidate`** - Allows custom validation while maintaining autofill

## Password Manager Compatibility Requirements

### Dashlane Requirements Met

1. **✅ Consistent field naming** - Uses underscores throughout
2. **✅ Both `id` and `name` attributes** - Required for field recognition
3. **✅ Proper `autocomplete` values** - Uses standard HTML5 autocomplete tokens
4. **✅ Semantic form structure** - Uses standard form elements with labels
5. **✅ Billing context grouping** - Groups related billing information

### Other Password Managers

The changes also improve compatibility with:
- **1Password** - Recognizes billing context and standard autocomplete values
- **LastPass** - Benefits from consistent naming and proper form structure
- **Bitwarden** - Uses autocomplete attributes for field identification
- **Browser built-in managers** - Chrome, Firefox, Safari, Edge

## Backend Changes Required

### Updated Struct Tags
The `DonationRequest` struct was updated to match new field names:

```go
type DonationRequest struct {
    FirstName    string `json:"first_name" form:"first_name"`
    LastName     string `json:"last_name" form:"last_name"`
    DonorEmail   string `json:"donor_email" form:"donor_email"`
    DonorPhone   string `json:"donor_phone" form:"donor_phone"`
    AddressLine1 string `json:"address_line1" form:"address_line1"`
    AddressLine2 string `json:"address_line2" form:"address_line2"`
    City         string `json:"city" form:"city"`
    State        string `json:"state" form:"state"`
    Zip          string `json:"zip" form:"zip_code"`
    Comments     string `json:"comments" form:"comments"`
}
```

### Updated Validation Error Keys
Error validation and template variables updated to match:

```go
errors.Add("first_name", "First name is required")
errors.Add("last_name", "Last name is required")
errors.Add("donor_email", "Email address is required")
// ... etc
```

## Testing Autofill Functionality

### Manual Testing Steps

1. **Install Dashlane** (or preferred password manager)
2. **Save billing information** in password manager
3. **Visit donation page**: `/donate`
4. **Click in first name field**
5. **Verify autofill suggestions appear**
6. **Select autofill option**
7. **Confirm all fields populate correctly**

### Expected Behavior

- **First name field click** should trigger autofill suggestions
- **All billing fields** should populate together when autofill is selected
- **Email and phone** should populate from contact information
- **Address fields** should populate as a complete set

## Troubleshooting

### If Autofill Doesn't Work

1. **Check browser settings** - Ensure autofill is enabled
2. **Verify saved data** - Ensure billing information is saved in password manager
3. **Clear browser cache** - Sometimes cached form data interferes
4. **Test in incognito mode** - Rules out extension conflicts

### Common Issues

- **Missing `name` attributes** - Password managers need both `id` and `name`
- **Inconsistent naming** - Mixing hyphens and underscores breaks recognition
- **Wrong autocomplete values** - Use standard HTML5 tokens only
- **Form structure issues** - Ensure proper `<form>` element with submit button

## Best Practices for Future Forms

### Field Naming
- **Always use underscores** instead of hyphens in field names
- **Match `id` and `name` attributes** exactly
- **Use descriptive, semantic names** (e.g., `first_name` not `fname`)

### Autocomplete Attributes
- **Use standard HTML5 tokens** from [MDN autocomplete documentation](https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes/autocomplete)
- **Group related fields** with context tokens (`billing`, `shipping`)
- **Set `autocomplete="off"`** only for sensitive or one-time fields

### Form Structure
- **Include proper labels** for all form fields
- **Use semantic HTML** elements (`<form>`, `<fieldset>`, `<legend>`)
- **Add `autocomplete="on"`** to form element unless specifically disabled
- **Include submit button** within form for password manager recognition

## Security Considerations

### Safe for Password Managers
- **No sensitive payment data** - Credit card fields handled by Helcim
- **Standard contact information** - Safe for autofill storage
- **Billing address only** - Appropriate for donation context

### Fields That Should NOT Autofill
- **One-time codes** - Use `autocomplete="off"`
- **Honeypot fields** - Hidden spam prevention fields
- **Dynamic tokens** - CSRF tokens, session IDs

This implementation follows security best practices while maximizing autofill compatibility.
