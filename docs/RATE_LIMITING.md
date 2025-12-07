# Rate Limiting Documentation

## Table of Contents

- [Overview](#overview)
- [Why Rate Limiting?](#why-rate-limiting)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [How It Works](#how-it-works)
- [Response Format](#response-format)
- [Testing](#testing)
- [Advanced Configuration](#advanced-configuration)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

---

## Overview

The service implements **two-tier rate limiting** to protect against abuse and brute force attacks:

1. **General API Rate Limiting** - 100 requests/minute per IP (configurable)
2. **Authentication Rate Limiting** - 5 login attempts/minute per IP (stricter, prevents brute force)

### Key Features

- ✅ **Per-IP tracking** - Prevents abuse from single sources
- ✅ **Two-tier protection** - Different limits for different endpoint types
- ✅ **Configurable** - Adjust limits via YAML or environment variables
- ✅ **Clear error messages** - Users know when and why they're limited
- ✅ **Production-ready** - Memory-efficient sliding window implementation
- ✅ **Optional** - Can be disabled per environment

---

## Why Rate Limiting?

### Security Benefits

**1. Brute Force Prevention**
```
Without Rate Limiting:
  Attacker tries 10,000 passwords/second
  Time to crack 8-char password: minutes

With Rate Limiting (5 req/min):
  Attacker limited to 5 attempts/minute
  Time to crack 8-char password: years
```

**2. DoS Attack Mitigation**
```
Without Rate Limiting:
  100,000 requests from one IP
  Server: Overloaded, crashes

With Rate Limiting (100 req/min):
  Request 101+: Rejected
  Server: Stable, serves legitimate users
```

**3. Resource Protection**
```
Prevents:
  - API abuse
  - Credential stuffing
  - Account enumeration
  - Resource exhaustion
```

---

## Architecture

### Two-Tier System

```
┌──────────────────────────────────────────────────────────┐
│ Client Request                                           │
└────────────────┬─────────────────────────────────────────┘
                 ↓
         ┌───────┴────────┐
         │                │
    /api/v1/auth    /api/v1/user
         │                │
         ↓                ↓
  ┌──────────────┐  ┌──────────────┐
  │ Auth Rate    │  │ General      │
  │ Limiter      │  │ Rate Limiter │
  │ (5 req/min)  │  │ (100 req/min)│
  └──────┬───────┘  └──────┬───────┘
         │                  │
         └──────┬───────────┘
                ↓
        ┌───────────────┐
        │ Allow Request │
        └───────────────┘
```

### Request Flow

```
1. Request arrives at /api/v1/auth/login
   ↓
2. Check General Rate Limiter (100 req/min)
   ├─> If exceeded: Return 429 Too Many Requests
   └─> If OK: Continue
   ↓
3. Check Auth Rate Limiter (5 req/min)
   ├─> If exceeded: Return 429 Too Many Requests
   └─> If OK: Continue
   ↓
4. Process login request
```

---

## Configuration

### YAML Configuration

**File:** `config/application.yaml`

```yaml
middleware:
  rateLimit:
    # General API rate limiting
    enabled: true                    # Enable/disable rate limiting
    max: 100                         # Max requests per window
    expiration: 1m                   # Time window (1 minute)
    skipFailedReq: false             # Count failed requests
    skipSuccessReq: false            # Count successful requests
    limitReached: "Too many requests, please try again later."

    # Authentication endpoint rate limiting (stricter)
    authEnabled: true                # Enable/disable auth rate limiting
    authMax: 5                       # Max login attempts per window
    authExpiration: 1m               # Time window (1 minute)
```

### Environment Variables

Override configuration with environment variables:

```bash
# General rate limiting
APP_MIDDLEWARE_RATELIMIT_ENABLED=true
APP_MIDDLEWARE_RATELIMIT_MAX=100
APP_MIDDLEWARE_RATELIMIT_EXPIRATION=1m
APP_MIDDLEWARE_RATELIMIT_SKIPFAILEDREQ=false
APP_MIDDLEWARE_RATELIMIT_SKIPSUCCESSREQ=false
APP_MIDDLEWARE_RATELIMIT_LIMITREACHED="Too many requests, please try again later."

# Authentication rate limiting
APP_MIDDLEWARE_RATELIMIT_AUTHENABLED=true
APP_MIDDLEWARE_RATELIMIT_AUTHMAX=5
APP_MIDDLEWARE_RATELIMIT_AUTHEXPIRATION=1m
```

### Per-Environment Configuration

**Development** (`config/application.yaml`):
```yaml
middleware:
  rateLimit:
    enabled: false        # Disabled for local testing
    authEnabled: true     # Still protect auth endpoints
    authMax: 10           # More lenient for testing
```

**Staging** (`config/application-staging.yaml`):
```yaml
middleware:
  rateLimit:
    enabled: true
    max: 200              # Higher limit for testing
    authEnabled: true
    authMax: 10
```

**Production** (`config/application-production.yaml`):
```yaml
middleware:
  rateLimit:
    enabled: true
    max: 100              # Strict production limits
    authEnabled: true
    authMax: 5            # Very strict for auth
```

---

## How It Works

### Sliding Window Algorithm

The rate limiter uses Fiber's built-in memory-based sliding window:

```
Window: 1 minute (60 seconds)
Max: 5 requests

Timeline:
0s:  Request 1 ✅ (1/5)
10s: Request 2 ✅ (2/5)
20s: Request 3 ✅ (3/5)
30s: Request 4 ✅ (4/5)
40s: Request 5 ✅ (5/5)
50s: Request 6 ❌ 429 Too Many Requests

60s: Window resets
     Request 7 ✅ (1/5)
```

### IP-Based Tracking

Each IP address has its own counter:

```go
KeyGenerator: func(c *fiber.Ctx) string {
    return c.IP() // Track per IP
}
```

**Example:**
```
IP: 192.168.1.100
  └─> Counter: 3/100 requests

IP: 192.168.1.101
  └─> Counter: 87/100 requests

IP: 10.0.0.50
  └─> Counter: 5/5 (BLOCKED)
```

### Request Counting

**Default: Count All Requests**
```yaml
skipFailedReq: false   # Count 4xx/5xx responses
skipSuccessReq: false  # Count 2xx responses
```

**Alternative: Skip Failed Requests**
```yaml
skipFailedReq: true    # Don't count 4xx/5xx
skipSuccessReq: false  # Count 2xx
```

**Use case:** Allow unlimited retries for client errors (not recommended)

---

## Response Format

### Success Response (Under Limit)

**Request:**
```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username_email": "user@example.com", "password": "SecurePass123!"}'
```

**Response:** `200 OK`
```json
{
  "code": "SUCCESS",
  "data": {
    "user": { ... },
    "token": "eyJhbGci...",
    "refresh_token": "eyJhbGci..."
  }
}
```

### Rate Limit Exceeded

**Request 6 (after 5 attempts):**
```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username_email": "user@example.com", "password": "wrong"}'
```

**Response:** `429 Too Many Requests`
```json
{
  "error": "auth_rate_limit_exceeded",
  "message": "Too many authentication attempts. Please try again in 1m0s."
}
```

### Response Headers

Rate limit responses include helpful headers:

```
X-RateLimit-Limit: 5
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1638360000
Retry-After: 60
```

---

## Testing

### Manual Testing

**1. Test General Rate Limiting**

```bash
# Make 101 requests quickly
for i in {1..101}; do
  curl -w "\nRequest $i: %{http_code}\n" \
    http://localhost:8081/api/v1/user/profile \
    -H "Authorization: Bearer YOUR_TOKEN"
done
```

**Expected:**
- Requests 1-100: `200 OK`
- Request 101: `429 Too Many Requests`

**2. Test Auth Rate Limiting**

```bash
# Make 6 login attempts
for i in {1..6}; do
  curl -w "\nAttempt $i: %{http_code}\n" \
    -X POST http://localhost:8081/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username_email":"test@example.com","password":"wrong"}'
done
```

**Expected:**
- Attempts 1-5: `200 OK` or `400 Bad Request` (invalid password)
- Attempt 6: `429 Too Many Requests`

**3. Test Window Reset**

```bash
# Hit rate limit
for i in {1..6}; do curl -X POST http://localhost:8081/api/v1/auth/login; done

# Wait 60 seconds
sleep 60

# Try again - should work
curl -X POST http://localhost:8081/api/v1/auth/login
```

### Automated Testing

**Unit Test Example:**

```go
func TestAuthRateLimit(t *testing.T) {
    app := setupTestApp()

    // Make 5 requests (should all succeed)
    for i := 0; i < 5; i++ {
        req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
        resp, _ := app.Test(req)
        assert.Equal(t, 200, resp.StatusCode)
    }

    // 6th request should be rate limited
    req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
    resp, _ := app.Test(req)
    assert.Equal(t, 429, resp.StatusCode)
}
```

---

## Advanced Configuration

### Custom Rate Limits per Endpoint

```go
// In user_route.go
authGroup := r.Group("/auth")

// Custom rate limit for password reset (more lenient)
resetGroup := authGroup.Group("/reset-password")
resetGroup.Use(middleware.RateLimitFilter(config.RateLimitConfig{
    Enabled: true,
    Max: 3,              // Only 3 reset attempts per hour
    Expiration: 1 * time.Hour,
}))
POST(resetGroup, "/", authHandler.ResetPassword)
```

### Skip Rate Limiting for Specific IPs

```go
// Custom middleware to skip rate limiting
func SkipRateLimitForIPs(allowedIPs []string) fiber.Handler {
    allowed := make(map[string]bool)
    for _, ip := range allowedIPs {
        allowed[ip] = true
    }

    return func(c *fiber.Ctx) error {
        if allowed[c.IP()] {
            return c.Next() // Skip rate limiting
        }
        return rateLimitMiddleware(c)
    }
}

// Usage
app.Use(SkipRateLimitForIPs([]string{"127.0.0.1", "10.0.0.1"}))
```

### Dynamic Rate Limits

```go
// Rate limit based on user role
func DynamicRateLimit(c *fiber.Ctx) fiber.Handler {
    userRole := getUserRole(c)

    var max int
    switch userRole {
    case "premium":
        max = 1000  // Premium users get higher limits
    case "standard":
        max = 100
    default:
        max = 50    // Free users get lower limits
    }

    return limiter.New(limiter.Config{
        Max: max,
        Expiration: 1 * time.Minute,
    })
}
```

---

## Monitoring

### Logging

Rate limit violations are logged:

```
[WARN] Rate limit exceeded ip=192.168.1.100 path=/api/v1/auth/login method=POST
[WARN] Auth rate limit exceeded ip=10.0.0.50 path=/api/v1/auth/login method=POST
```

### Metrics

Track rate limiting in `/metrics`:

```prometheus
# Rate limit violations (custom metric)
rate_limit_exceeded_total{endpoint="/api/v1/auth/login"} 127
rate_limit_exceeded_total{endpoint="/api/v1/user/profile"} 15

# Per-IP tracking
rate_limit_exceeded_by_ip{ip="192.168.1.100"} 45
```

### Alerts

Set up alerts for suspicious activity:

```yaml
# Prometheus Alert
- alert: HighRateLimitViolations
  expr: rate(rate_limit_exceeded_total[5m]) > 100
  annotations:
    summary: "High rate limit violations detected"
    description: "{{ $value }} rate limit violations in the last 5 minutes"
```

---

## Troubleshooting

### Problem: Legitimate Users Getting Blocked

**Symptoms:**
- Users complaining they can't log in
- Many 429 errors in logs
- High rate limit violation count

**Possible Causes:**
1. Limits too strict
2. Multiple users behind same NAT/proxy
3. Mobile app retrying too aggressively

**Solutions:**

1. **Increase Limits**
```yaml
middleware:
  rateLimit:
    max: 200              # Double the limit
    authMax: 10           # More lenient
```

2. **Skip Failed Requests**
```yaml
middleware:
  rateLimit:
    skipFailedReq: true   # Don't count errors
```

3. **Use User-Based Tracking**
```go
// Instead of IP-based, use user ID
KeyGenerator: func(c *fiber.Ctx) string {
    userID := c.Locals("user_id")
    if userID != nil {
        return fmt.Sprintf("user:%v", userID)
    }
    return c.IP() // Fallback to IP for unauthenticated
}
```

### Problem: Rate Limiting Not Working

**Symptoms:**
- Can make unlimited requests
- No 429 errors

**Solutions:**

1. **Check if Enabled**
```yaml
middleware:
  rateLimit:
    enabled: true        # Must be true!
```

2. **Check Middleware Order**
```go
// Rate limiting must come BEFORE route handlers
app.Use(middleware.RateLimitFilter(...))
app.Post("/api/v1/auth/login", handler.Login)
```

3. **Check Route Group**
```go
// Make sure rate limiting is applied to the right group
authGroup.Use(middleware.AuthRateLimitFilter(...))
```

### Problem: Counter Not Resetting

**Symptoms:**
- Always blocked even after waiting
- Counter doesn't decrease

**Cause:** Incorrect expiration time

**Solution:**
```yaml
middleware:
  rateLimit:
    expiration: 1m       # Use time.Duration format (1m, 60s, etc.)
```

---

## Best Practices

### 1. Different Limits for Different Endpoints

```yaml
# Authentication: Very strict
authMax: 5
authExpiration: 1m

# Profile updates: Moderate
profileMax: 20
profileExpiration: 1m

# Read-only endpoints: Lenient
readMax: 200
readExpiration: 1m
```

### 2. Progressive Delays

Instead of hard blocking, increase delays:

```go
// After 3 failed attempts, add delay
if failedAttempts > 3 {
    time.Sleep(time.Duration(failedAttempts) * time.Second)
}
```

### 3. Clear Error Messages

```yaml
limitReached: "Too many requests. Please wait 60 seconds and try again."
```

**Include:**
- What happened (too many requests)
- When they can try again (60 seconds)
- How to avoid (wait)

### 4. Whitelist Internal Services

```go
// Don't rate limit health checks
if strings.HasPrefix(c.Path(), "/health") {
    return c.Next()
}

// Don't rate limit internal microservices
if c.Get("X-Internal-Service") == "true" {
    return c.Next()
}
```

### 5. Use Redis for Distributed Rate Limiting

For multi-server deployments:

```go
// Use Redis-based storage instead of in-memory
app.Use(limiter.New(limiter.Config{
    Max: 100,
    Expiration: 1 * time.Minute,
    Storage: redis.New(redis.Config{
        Host:     "localhost",
        Port:     6379,
        Database: 2,
    }),
}))
```

**Benefits:**
- Shared counter across all servers
- More accurate limiting
- Prevents bypass by hitting different servers

### 6. Monitor and Adjust

Track metrics and adjust limits:

```
Week 1: max=100, violations=50/day  → Too strict
Week 2: max=200, violations=5/day   → Good balance
Week 3: max=200, violations=500/day → Possible attack, investigate
```

---

## Summary

Rate limiting provides:

✅ **Brute force protection** - Limits password guessing attacks
✅ **DoS mitigation** - Prevents resource exhaustion
✅ **Resource protection** - Ensures fair API usage
✅ **Two-tier system** - Different limits for different needs
✅ **Configurable** - Adjust per environment
✅ **Observable** - Logging and metrics included

**Configuration:**
- General API: 100 requests/minute per IP
- Authentication: 5 attempts/minute per IP
- Fully configurable via YAML or environment variables

For questions or issues, see [Troubleshooting](#troubleshooting) or check logs.
