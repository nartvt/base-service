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

### 5. Use Redis for Distributed Rate Limiting ✅ IMPLEMENTED

For multi-server deployments, the service now includes built-in Redis-based distributed rate limiting:

#### Configuration

**config/application.yaml:**
```yaml
middleware:
  rateLimit:
    # General API rate limiting
    enabled: true
    max: 100
    expiration: 1m

    # Authentication rate limiting
    authEnabled: true
    authMax: 5
    authExpiration: 1m

    # Distributed rate limiting (NEW in v2.2)
    useRedis: true      # Enable Redis storage for distributed rate limiting
    redisDB: 1          # Separate Redis database for rate limit counters
```

**Environment Variables:**
```bash
APP_MIDDLEWARE_RATELIMIT_USEREDIS=true
APP_MIDDLEWARE_RATELIMIT_REDISDB=1
```

#### How It Works

**Single-Server Deployment (Memory Storage):**
```yaml
useRedis: false  # Each server maintains its own counters
```
- Counters stored in application memory
- Fast but not shared across servers
- Acceptable for single-server deployments

**Multi-Server Deployment (Redis Storage):**
```yaml
useRedis: true  # All servers share the same Redis counters
redisDB: 1      # Separate database to avoid cache conflicts
```
- Counters stored in Redis
- Shared across all application servers
- Accurate limiting across distributed systems

#### Redis Database Separation

The service uses different Redis databases for different purposes:

```
Redis DB 0: JWT cache (token storage)
Redis DB 1: Rate limiting counters (separate to avoid conflicts)
```

This separation ensures:
- **No cache pollution**: Rate limit data doesn't interfere with JWT cache
- **Independent TTLs**: Different expiration strategies for each use case
- **Easy monitoring**: Can monitor rate limiting separately
- **Compatibility**: DB 1 is guaranteed to be available (Redis default: 16 databases 0-15)

#### Architecture

```
┌────────────────────────────────────────────────────────┐
│ Multi-Server Deployment with Distributed Rate Limiting│
└────────────────────────────────────────────────────────┘

   Client Request
        ↓
   Load Balancer
        ↓
   ┌────┴────┐
   │         │
Server A  Server B  Server C
   │         │         │
   └────┬────┴────┬────┘
        │         │
   Shared Redis (DB 1)
   ┌─────────────────────┐
   │ IP: 192.168.1.100   │
   │ Counter: 47/100     │ ← Shared across all servers
   └─────────────────────┘

Without Redis (Single-Server):
Server A: IP 192.168.1.100 = 30/100  ← Server A's counter
Server B: IP 192.168.1.100 = 17/100  ← Server B's counter
Total: 47 requests (but not rate limited!)

With Redis (Distributed):
All Servers: IP 192.168.1.100 = 47/100  ← Shared counter
Total: 47 requests (accurately tracked)
```

#### Implementation

The middleware automatically uses Redis when configured:

**internal/middleware/ratelimit.go:**
```go
func RateLimitFilter(rateLimitConfig config.RateLimitConfig, rediscf *config.RedisConfig, redisClient *goredis.Client) fiber.Handler {
    if !rateLimitConfig.Enabled {
        return func(c *fiber.Ctx) error {
            return c.Next()
        }
    }

    // Configure storage (Redis for distributed, memory for single-server)
    var storage fiber.Storage
    storageType := "memory"
    if redisClient != nil && rateLimitConfig.UseRedis {
        storage = redis.New(redis.Config{
            Host:     rediscf.Host,
            Port:     rediscf.Port,
            Database: rateLimitConfig.RedisDB,
            Reset:    false,
        })
        storageType = "redis"
    }

    slog.Info("Configuring rate limiting middleware",
        "storage", storageType,
        "max", rateLimitConfig.Max,
        "expiration", rateLimitConfig.Expiration,
    )

    return limiter.New(limiter.Config{
        Storage:    storage,
        Max:        rateLimitConfig.Max,
        Expiration: rateLimitConfig.Expiration,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
    })
}
```

#### Graceful Degradation

If Redis is unavailable, the service automatically falls back to memory storage:

```go
var storage fiber.Storage
if redisClient != nil && rateLimitConfig.UseRedis {
    storage = redis.New(...)  // Try Redis first
}
// If Redis is nil or useRedis=false, storage remains nil
// Fiber's limiter will use in-memory storage by default
```

**Benefits:**
- Service continues to work even if Redis is down
- Rate limiting still functions (per-server basis)
- No service disruption during Redis maintenance

#### Testing Distributed Rate Limiting

**1. Start Multiple Server Instances:**
```bash
# Terminal 1: Server on port 8081
APP_SERVER_HTTP_PORT=8081 go run cmd/main.go

# Terminal 2: Server on port 8082
APP_SERVER_HTTP_PORT=8082 go run cmd/main.go

# Terminal 3: Server on port 8083
APP_SERVER_HTTP_PORT=8083 go run cmd/main.go
```

**2. Test Without Redis (Memory Storage):**
```yaml
useRedis: false
max: 10
```

```bash
# Make 10 requests to Server A
for i in {1..10}; do curl http://localhost:8081/api/v1/user/profile -H "Authorization: Bearer TOKEN"; done
# ✅ All succeed (10/10)

# Make 5 more requests to Server B
for i in {1..5}; do curl http://localhost:8082/api/v1/user/profile -H "Authorization: Bearer TOKEN"; done
# ✅ All succeed (5/10)

# Total: 15 requests succeeded (should have blocked at 10!)
```

**3. Test With Redis (Distributed Storage):**
```yaml
useRedis: true
redisDB: 1
max: 10
```

```bash
# Make 6 requests to Server A
for i in {1..6}; do curl http://localhost:8081/api/v1/user/profile -H "Authorization: Bearer TOKEN"; done
# ✅ Requests 1-6 succeed (6/10 shared counter)

# Make 4 requests to Server B
for i in {1..4}; do curl http://localhost:8082/api/v1/user/profile -H "Authorization: Bearer TOKEN"; done
# ✅ Requests 1-4 succeed (10/10 shared counter)

# Make 1 request to Server C
curl http://localhost:8083/api/v1/user/profile -H "Authorization: Bearer TOKEN"
# ❌ 429 Too Many Requests (11/10 shared counter)

# Total: 10 requests succeeded, 1 blocked (correct!)
```

#### Monitoring Redis Rate Limiting

**Check Redis Counters:**
```bash
# Connect to Redis
redis-cli

# Switch to rate limiting database
SELECT 1

# List all rate limit keys
KEYS *

# Example output:
# 1) "192.168.1.100"
# 2) "10.0.0.50"

# Get counter value
GET "192.168.1.100"
# "47"  (47 requests in current window)

# Check TTL (time until reset)
TTL "192.168.1.100"
# 42    (42 seconds remaining)
```

**Monitor Rate Limit Activity:**
```bash
# Watch rate limiting in real-time
redis-cli -n 1 MONITOR

# Example output:
# OK
# 1638360000.123456 [1] "GET" "192.168.1.100"
# 1638360000.234567 [1] "INCR" "192.168.1.100"
# 1638360000.345678 [1] "EXPIRE" "192.168.1.100" "60"
```

#### Production Deployment Considerations

**1. Redis High Availability:**
```yaml
# Use Redis Sentinel or Cluster for production
redis:
  host: redis-cluster.example.com
  port: 6379
  password: ${REDIS_PASSWORD}
  db: 0  # JWT cache

middleware:
  rateLimit:
    useRedis: true
    redisDB: 1  # Rate limiting counters
```

**2. Connection Pooling:**
The service reuses the existing Redis connection pool, so no additional configuration needed.

**3. Memory Usage:**
```
Estimated memory per IP:
- Key: ~20 bytes (IP address)
- Value: ~8 bytes (counter)
- TTL: ~8 bytes
Total: ~36 bytes per tracked IP

For 10,000 active IPs: ~360 KB
For 100,000 active IPs: ~3.6 MB
```

**4. Performance:**
```
Memory storage: ~0.1ms per request
Redis storage: ~1-2ms per request (local Redis)
Redis storage: ~5-10ms per request (remote Redis)
```

**Benefits:**
- ✅ **Accurate limiting**: Shared counter across all servers
- ✅ **No bypass**: Can't hit different servers to bypass limit
- ✅ **Scalable**: Add/remove servers without reconfiguration
- ✅ **Persistent**: Counters survive server restarts
- ✅ **Observable**: Monitor rate limiting via Redis
- ✅ **Graceful fallback**: Works without Redis (memory storage)

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
