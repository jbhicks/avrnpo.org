# Subscription Management API Quick Reference

*Generated: August 18, 2025*

## 🔐 Authentication Required

All subscription management endpoints require user authentication via session-based login.

## 📋 Available Routes

### User Interface Routes

| Method | Route | Handler | Description |
|--------|-------|---------|-------------|
| `GET` | `/account/subscriptions` | `SubscriptionsList` | List all user subscriptions |
| `GET` | `/account/subscriptions/{id}` | `SubscriptionDetails` | View subscription details |
| `POST` | `/account/subscriptions/{id}/cancel` | `CancelSubscription` | Cancel subscription |

### Backend Service Functions

| Function | Purpose | Helcim API |
|----------|---------|------------|
| `GetSubscription(id)` | Retrieve subscription details | `GET /v2/subscriptions/{id}` |
| `CancelSubscription(id)` | Cancel subscription | `DELETE /v2/subscriptions/{id}` |
| `UpdateSubscription(id, updates)` | Modify subscription | `PATCH /v2/subscriptions` |
| `ListSubscriptionsByCustomer(customerID)` | List customer subscriptions | `GET /v2/subscriptions?customerId={id}` |

## 🔄 User Flow

```
1. User logs in → /auth/new
2. View account → /account
3. Click "View My Subscriptions" → /account/subscriptions
4. Select subscription → /account/subscriptions/{id}
5. Cancel if needed → POST /account/subscriptions/{id}/cancel
```

## 💾 Database Schema

```sql
-- Added to donations table
user_id UUID NULL  -- Links donation to user account
subscription_id STRING NULL  -- Helcim subscription ID
customer_id STRING NULL  -- Helcim customer ID
payment_plan_id STRING NULL  -- Helcim payment plan ID
transaction_id STRING NULL  -- Helcim transaction ID
```

## 🚨 Security Features

- **Session authentication**: `SetCurrentUser(Authorize(handler))`
- **Ownership verification**: Users can only access their own subscriptions
- **CSRF protection**: Built into Buffalo framework
- **Audit logging**: All actions logged with user context
- **Error sanitization**: No sensitive data exposed in error messages

## 🧪 Testing Endpoints

### Manual Testing
```bash
# 1. Create account and log in
curl -X POST http://localhost:3000/users -d "email=test@example.com&password=password123&password_confirmation=password123"

# 2. Login (get session cookie)
curl -c cookies.txt -X POST http://localhost:3000/auth -d "email=test@example.com&password=password123"

# 3. View subscriptions (requires session)
curl -b cookies.txt http://localhost:3000/account/subscriptions

# 4. View subscription details
curl -b cookies.txt http://localhost:3000/account/subscriptions/{subscription-id}

# 5. Cancel subscription (requires confirmation)
curl -b cookies.txt -X POST http://localhost:3000/account/subscriptions/{subscription-id}/cancel
```

### Test Database Queries
```sql
-- View user donations with subscriptions
SELECT d.*, u.email FROM donations d 
JOIN users u ON d.user_id = u.id 
WHERE d.subscription_id IS NOT NULL;

-- Check subscription status
SELECT user_id, subscription_id, status, created_at 
FROM donations 
WHERE donation_type = 'recurring' 
ORDER BY created_at DESC;
```

## ⚠️ Error Handling

| Error | Cause | User Message |
|-------|-------|--------------|
| 401 Unauthorized | Not logged in | "You must be authorized to see that page" |
| 404 Not Found | Subscription doesn't exist or not owned | "Subscription not found" |
| 500 API Error | Helcim API connectivity | "Unable to cancel subscription. Please contact support." |

## 📞 Support Contact

When API systems are unavailable, users are directed to:
- **Email**: info@avrnpo.org
- **Subject**: Subscription Management Request

## 🔧 Configuration

### Environment Variables
```bash
HELCIM_PRIVATE_API_KEY=your_api_key_here
```

### Required Middleware
```go
app.GET("/account/subscriptions", SetCurrentUser(Authorize(SubscriptionsList)))
```

---

**Related Documentation:**
- [Full Implementation Guide](./subscription-management-implementation.md)
- [Helcim API Reference](./helcim-api-reference.md)
- [Security Guidelines](./SECURITY-GUIDELINES.md)
