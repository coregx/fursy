// Package main demonstrates all built-in and custom middleware in fursy.
//
// This example showcases:
//   - All 8 built-in middleware (Logger, Recovery, CORS, BasicAuth, JWT, RateLimit, CircuitBreaker, Secure)
//   - Custom middleware patterns (RequestID, Timing, Authentication)
//   - Middleware ordering best practices
//   - Group-level middleware
//   - Conditional middleware (Skipper pattern)
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/middleware"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Create a new fursy router.
	router := fursy.New()

	// =============================================
	// MIDDLEWARE ORDERING (CRITICAL!)
	// =============================================
	// Order matters! Apply middleware in this sequence:
	// 1. Logger - Log all requests (first to capture everything)
	// 2. Recovery - Recover from panics (second to catch all panics)
	// 3. RequestID - Add request ID for tracing
	// 4. Timing - Measure response time
	// 5. Secure - Set security headers
	// 6. CORS - Handle CORS (for API endpoints)
	//
	// Per-route/group middleware:
	// 7. RateLimit - Rate limiting (per API group)
	// 8. BasicAuth/JWT - Authentication (per protected group)
	// 9. CircuitBreaker - Fault tolerance (per service)

	// ===========================================
	// GLOBAL MIDDLEWARE (Apply to all routes)
	// ===========================================

	// 1. Logger - Structured logging with slog
	router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
		SkipPaths: []string{"/health"}, // Skip health check logging
	}))

	// 2. Recovery - Panic recovery with stack traces
	router.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
		Logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError,
		})),
		DisableStackTrace: false, // Enable stack traces for debugging
	}))

	// 3. RequestID - Custom middleware for request tracing
	router.Use(RequestIDMiddleware())

	// 4. Timing - Custom middleware for response time measurement
	router.Use(TimingMiddleware())

	// 5. Secure - OWASP 2025 security headers
	router.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XFrameOptions:         middleware.XFrameOptionsSameOrigin,
		ContentTypeNosniff:    middleware.ContentTypeNosniffValue,
		ReferrerPolicy:        middleware.ReferrerPolicyStrictOrigin,
		HSTSMaxAge:            31536000, // 1 year (only if using HTTPS)
		ContentSecurityPolicy: "default-src 'self'",
	}))

	// ===========================================
	// PUBLIC ROUTES (No authentication)
	// ===========================================

	// Home endpoint - No middleware
	router.GET("/", func(c *fursy.Context) error {
		return c.OK(map[string]string{
			"message":    "Welcome to Fursy Middleware Demo",
			"version":    "1.0.0",
			"request_id": c.GetString("request_id"),
		})
	})

	// Health check - No middleware (skipped by logger)
	router.GET("/health", func(c *fursy.Context) error {
		return c.OK(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Panic test - Demonstrates Recovery middleware
	router.GET("/panic", func(c *fursy.Context) error {
		panic("intentional panic for testing recovery middleware")
	})

	// ===========================================
	// API GROUP (Public API with CORS + RateLimit)
	// ===========================================

	api := router.Group("/api")

	// Apply CORS middleware to API group
	api.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     "http://localhost:3000,https://example.com",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization,X-Request-ID",
		ExposeHeaders:    "X-Request-ID,X-Response-Time",
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Apply rate limiting to API group (10 req/s, burst 20)
	api.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
		Rate:    10,   // 10 requests per second
		Burst:   20,   // Allow bursts up to 20
		Headers: true, // Set X-RateLimit-* headers
		KeyFunc: func(c *fursy.Context) string {
			// Rate limit by IP address
			return c.Request.RemoteAddr
		},
	}))

	// Public API endpoint
	api.GET("/public", func(c *fursy.Context) error {
		return c.OK(map[string]interface{}{
			"message":    "This is a public API endpoint",
			"cors":       "enabled",
			"rate_limit": "10 req/s",
			"request_id": c.GetString("request_id"),
		})
	})

	// ===========================================
	// PROTECTED GROUP (JWT Authentication)
	// ===========================================

	protected := api.Group("/protected")

	// JWT middleware - Validates JWT tokens
	jwtSecret := []byte("my-super-secret-key-change-in-production")
	protected.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:    jwtSecret,
		SigningMethod: "HS256",
		TokenLookup:   "header:Authorization",
		AuthScheme:    "Bearer",
		Skipper: func(c *fursy.Context) bool {
			// Skip JWT for OPTIONS requests (CORS preflight)
			return c.Request.Method == http.MethodOptions
		},
		ErrorHandler: func(c *fursy.Context, err error) error {
			return c.Problem(fursy.Problem{
				Type:   "https://example.com/errors/unauthorized",
				Title:  "Unauthorized",
				Status: http.StatusUnauthorized,
				Detail: "Invalid or missing JWT token: " + err.Error(),
			})
		},
	}))

	// Protected endpoint - List users
	protected.GET("/users", func(c *fursy.Context) error {
		// Access JWT claims
		claims := c.Get(middleware.JWTContextKey).(jwt.MapClaims)
		userID := claims["sub"].(string)

		return c.OK(map[string]interface{}{
			"message":      "List of users",
			"current_user": userID,
			"users": []map[string]string{
				{"id": "1", "name": "Alice"},
				{"id": "2", "name": "Bob"},
			},
		})
	})

	// Protected endpoint - Create user
	protected.POST("/users", func(c *fursy.Context) error {
		claims := c.Get(middleware.JWTContextKey).(jwt.MapClaims)
		userID := claims["sub"].(string)

		return c.Created(map[string]interface{}{
			"message": "User created successfully",
			"creator": userID,
			"user":    map[string]string{"id": "3", "name": "Charlie"},
		})
	})

	// ===========================================
	// BASIC AUTH GROUP
	// ===========================================

	basic := router.Group("/basic")

	// BasicAuth middleware with account validation
	accounts := map[string]string{
		"admin": "secret",
		"user":  "password123",
	}
	basic.Use(middleware.BasicAuth(middleware.BasicAuthAccounts(accounts)))

	// Basic auth protected endpoint
	basic.GET("/dashboard", func(c *fursy.Context) error {
		username := c.GetString(middleware.UserContextKey)
		return c.OK(map[string]string{
			"message":  "Welcome to the dashboard",
			"username": username,
			"auth":     "basic",
		})
	})

	// ===========================================
	// CIRCUIT BREAKER DEMO
	// ===========================================

	circuit := router.Group("/circuit")

	// Circuit breaker middleware (simulates flaky service)
	circuit.Use(middleware.CircuitBreakerWithConfig(middleware.CircuitBreakerConfig{
		ConsecutiveFailures: 3,                // Open after 3 consecutive failures
		Timeout:             10 * time.Second, // Stay open for 10 seconds
		MaxRequests:         2,                // Allow 2 requests in half-open state
		OnStateChange: func(from, to middleware.State) {
			slog.Info("Circuit breaker state change",
				"from", from.String(),
				"to", to.String(),
			)
		},
		ErrorHandler: func(c *fursy.Context) error {
			return c.Problem(fursy.Problem{
				Type:   "https://example.com/errors/service-unavailable",
				Title:  "Service Unavailable",
				Status: http.StatusServiceUnavailable,
				Detail: "Circuit breaker is open - service is temporarily unavailable",
			})
		},
	}))

	// Flaky endpoint for testing circuit breaker
	failureCount := 0
	circuit.GET("/flaky", func(c *fursy.Context) error {
		failureCount++

		// Fail first 3 requests to trigger circuit breaker
		if failureCount <= 3 {
			return c.Problem(fursy.Problem{
				Type:   "https://example.com/errors/internal",
				Title:  "Internal Server Error",
				Status: http.StatusInternalServerError,
				Detail: "Simulated service failure",
			})
		}

		// Succeed after 3 failures
		return c.OK(map[string]interface{}{
			"message":       "Service recovered",
			"failure_count": failureCount,
		})
	})

	// ===========================================
	// TOKEN GENERATION ENDPOINT (For testing JWT)
	// ===========================================

	router.POST("/auth/token", func(c *fursy.Context) error {
		// Simple token generation (in production, validate credentials!)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user123",                            // User ID
			"iss": "fursy-demo",                         // Issuer
			"exp": time.Now().Add(1 * time.Hour).Unix(), // Expires in 1 hour
			"iat": time.Now().Unix(),                    // Issued at
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			return c.Problem(fursy.Problem{
				Type:   "https://example.com/errors/token-generation",
				Title:  "Token Generation Failed",
				Status: http.StatusInternalServerError,
				Detail: err.Error(),
			})
		}

		return c.OK(map[string]string{
			"token": tokenString,
			"type":  "Bearer",
			"usage": "Add to Authorization header: Bearer <token>",
		})
	})

	// ===========================================
	// START SERVER
	// ===========================================

	slog.Info("Server starting", "port", 8080)
	slog.Info("Middleware Demo Endpoints:")
	slog.Info("  Public:")
	slog.Info("    GET  /                  - Home")
	slog.Info("    GET  /health            - Health check (no logging)")
	slog.Info("    GET  /panic             - Panic test (recovery demo)")
	slog.Info("")
	slog.Info("  API (CORS + RateLimit):")
	slog.Info("    GET  /api/public        - Public API")
	slog.Info("")
	slog.Info("  Protected (JWT):")
	slog.Info("    POST /auth/token        - Generate JWT token")
	slog.Info("    GET  /api/protected/users - List users (requires JWT)")
	slog.Info("    POST /api/protected/users - Create user (requires JWT)")
	slog.Info("")
	slog.Info("  Basic Auth:")
	slog.Info("    GET  /basic/dashboard   - Dashboard (user:password123 or admin:secret)")
	slog.Info("")
	slog.Info("  Circuit Breaker:")
	slog.Info("    GET  /circuit/flaky     - Flaky service (fails 3 times, then succeeds)")
	slog.Info("")
	slog.Info("Visit: http://localhost:8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
