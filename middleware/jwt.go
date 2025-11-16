// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware provides JWT authentication middleware.
package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coregx/fursy"
	"github.com/golang-jwt/jwt/v5"
)

// JWTContextKey is the key used to store JWT claims in the context.
const JWTContextKey = "jwt"

// JWTTokenContextKey is the key used to store the raw JWT token string in the context.
const JWTTokenContextKey = "jwt_token"

// jwtAlgoNone is the insecure "none" algorithm (forbidden for security).
const jwtAlgoNone = "none"

// Common JWT errors.
var (
	ErrJWTMissing     = errors.New("missing or malformed jwt")
	ErrJWTInvalid     = errors.New("invalid or expired jwt")
	ErrJWTAlgorithm   = errors.New("invalid jwt signing algorithm")
	ErrJWTNoneAlgo    = errors.New("jwt 'none' algorithm is forbidden")
	ErrJWTExpired     = errors.New("jwt token has expired")
	ErrJWTNotValidYet = errors.New("jwt token not valid yet")
)

// JWTConfig defines the configuration for the JWT middleware.
type JWTConfig struct {
	// SigningKey is the key used to validate JWT signatures.
	// For HS256: []byte("secret-key")
	// For RS256/ES256: *rsa.PublicKey, *ecdsa.PublicKey, or []byte(publicKeyPEM)
	// Required.
	SigningKey interface{}

	// SigningMethod is the expected JWT signing algorithm.
	// Supported: HS256, HS384, HS512, RS256, RS384, RS512, ES256, ES384, ES512
	// Default: "HS256"
	// IMPORTANT: This prevents algorithm confusion attacks.
	SigningMethod string

	// Skipper defines a function to skip the middleware.
	// Default: nil (middleware always executes)
	Skipper func(c *fursy.Context) bool

	// TokenLookup is a string in the form of "<source>:<name>" that specifies
	// where to extract the token from.
	// Supported sources:
	//   - "header:<name>" - Authorization header (default: "header:Authorization")
	//   - "query:<name>" - URL query parameter
	//   - "cookie:<name>" - Cookie
	// Default: "header:Authorization"
	TokenLookup string

	// AuthScheme is the authorization scheme (e.g., "Bearer").
	// Only used when TokenLookup is "header:Authorization".
	// Default: "Bearer"
	AuthScheme string

	// Claims is a function that returns a fresh claims object for parsing.
	// This allows custom claims structures.
	// Default: returns jwt.MapClaims
	Claims func() jwt.Claims

	// ValidateIssuer enables issuer (iss) validation.
	// If set, the JWT must contain a matching "iss" claim.
	// Default: "" (no validation)
	ValidateIssuer string

	// ValidateAudience enables audience (aud) validation.
	// If set, the JWT must contain a matching "aud" claim.
	// Default: "" (no validation)
	ValidateAudience string

	// ErrorHandler is called when JWT validation fails.
	// Default: returns 401 Unauthorized
	ErrorHandler func(c *fursy.Context, err error) error

	// SuccessHandler is called after successful JWT validation.
	// Can be used to extract user info from claims.
	// Default: nil
	SuccessHandler func(c *fursy.Context, claims jwt.Claims) error

	// AllowedAlgorithms is a list of allowed signing algorithms.
	// If empty, only the algorithm specified in SigningMethod is allowed.
	// This provides defense-in-depth against algorithm confusion attacks.
	// Default: nil (use SigningMethod only)
	AllowedAlgorithms []string
}

// JWT returns a middleware that provides JWT authentication.
//
// The middleware:
//   - Extracts JWT token from Authorization header (Bearer scheme)
//   - Validates signature using the provided key
//   - Validates standard claims (exp, nbf, iat)
//   - Prevents "none" algorithm attack
//   - Prevents algorithm confusion attack
//   - Stores validated claims in context
//
// Security features (2025 best practices):
//   - Explicitly forbids "none" algorithm
//   - Enforces expected signing algorithm
//   - Validates expiration (exp) and not-before (nbf) claims
//   - Supports issuer (iss) and audience (aud) validation
//   - Defense-in-depth with AllowedAlgorithms list
//
// Example with HS256:
//
//	secret := []byte("my-secret-key")
//	router := fursy.New()
//	router.Use(middleware.JWT(secret))
//
// Example with RS256:
//
//	publicKey, _ := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
//	config := middleware.JWTConfig{
//	    SigningKey: publicKey,
//	    SigningMethod: "RS256",
//	}
//	router.Use(middleware.JWTWithConfig(config))
//
// Access claims in handlers:
//
//	router.GET("/protected", func(c *fursy.Context) error {
//	    claims := c.Get(middleware.JWTContextKey).(jwt.MapClaims)
//	    userID := claims["sub"].(string)
//	    return c.String(200, "Hello, "+userID)
//	})
func JWT(signingKey interface{}) fursy.HandlerFunc {
	return JWTWithConfig(JWTConfig{
		SigningKey: signingKey,
	})
}

// JWTWithConfig returns a middleware with custom JWT configuration.
//
// Example with custom claims:
//
//	type CustomClaims struct {
//	    jwt.RegisteredClaims
//	    UserID string `json:"user_id"`
//	    Role   string `json:"role"`
//	}
//
//	config := middleware.JWTConfig{
//	    SigningKey: secret,
//	    SigningMethod: "HS256",
//	    Claims: func() jwt.Claims {
//	        return &CustomClaims{}
//	    },
//	    ValidateIssuer: "my-app",
//	    ValidateAudience: "my-users",
//	}
//	router.Use(middleware.JWTWithConfig(config))
//
//nolint:gocognit,gocyclo,cyclop // Complex JWT validation logic is necessary for security
func JWTWithConfig(config JWTConfig) fursy.HandlerFunc {
	// Validate config.
	if config.SigningKey == nil {
		panic("fursy/middleware: JWT signing key cannot be nil")
	}

	// Set defaults.
	if config.SigningMethod == "" {
		config.SigningMethod = "HS256"
	}

	if config.TokenLookup == "" {
		config.TokenLookup = "header:Authorization"
	}

	if config.AuthScheme == "" {
		config.AuthScheme = "Bearer"
	}

	if config.Claims == nil {
		config.Claims = func() jwt.Claims {
			return jwt.MapClaims{}
		}
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultJWTErrorHandler
	}

	// Build allowed algorithms map for efficient lookup.
	allowedAlgos := make(map[string]bool)
	if len(config.AllowedAlgorithms) > 0 {
		for _, algo := range config.AllowedAlgorithms {
			// Security: Explicitly forbid "none" algorithm.
			if strings.EqualFold(algo, jwtAlgoNone) {
				panic("fursy/middleware: JWT 'none' algorithm is forbidden")
			}
			allowedAlgos[algo] = true
		}
	} else {
		// If no AllowedAlgorithms, only allow SigningMethod.
		allowedAlgos[config.SigningMethod] = true
	}

	// Security: Explicitly forbid "none" in SigningMethod.
	if strings.EqualFold(config.SigningMethod, jwtAlgoNone) {
		panic("fursy/middleware: JWT 'none' algorithm is forbidden")
	}

	// Parse TokenLookup.
	parts := strings.Split(config.TokenLookup, ":")
	if len(parts) != 2 {
		panic("fursy/middleware: invalid TokenLookup format (expected '<source>:<name>')")
	}
	extractorSource := parts[0]
	extractorParam := parts[1]

	return func(c *fursy.Context) error {
		// Skip if Skipper returns true.
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		// Extract token.
		tokenString := extractToken(c, extractorSource, extractorParam, config.AuthScheme)
		if tokenString == "" {
			return config.ErrorHandler(c, ErrJWTMissing)
		}

		// Parse and validate token.
		claims := config.Claims()
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Security: Prevent "none" algorithm attack.
			if token.Method.Alg() == "none" {
				return nil, ErrJWTNoneAlgo
			}

			// Security: Prevent algorithm confusion attack.
			alg := token.Method.Alg()
			if !allowedAlgos[alg] {
				return nil, fmt.Errorf("%w: expected %s, got %s", ErrJWTAlgorithm, config.SigningMethod, alg)
			}

			return config.SigningKey, nil
		})

		if err != nil {
			// Check for specific errors.
			if errors.Is(err, jwt.ErrTokenExpired) {
				return config.ErrorHandler(c, ErrJWTExpired)
			}
			if errors.Is(err, jwt.ErrTokenNotValidYet) {
				return config.ErrorHandler(c, ErrJWTNotValidYet)
			}
			return config.ErrorHandler(c, err)
		}

		if !token.Valid {
			return config.ErrorHandler(c, ErrJWTInvalid)
		}

		// Validate issuer if configured.
		if config.ValidateIssuer != "" {
			if !validateClaim(claims, "iss", config.ValidateIssuer) {
				return config.ErrorHandler(c, errors.New("invalid jwt issuer"))
			}
		}

		// Validate audience if configured.
		if config.ValidateAudience != "" {
			if !validateClaim(claims, "aud", config.ValidateAudience) {
				return config.ErrorHandler(c, errors.New("invalid jwt audience"))
			}
		}

		// Store token and claims in context.
		c.Set(JWTTokenContextKey, tokenString)
		c.Set(JWTContextKey, claims)

		// Call success handler if configured.
		if config.SuccessHandler != nil {
			if err := config.SuccessHandler(c, claims); err != nil {
				return err
			}
		}

		return c.Next()
	}
}

// extractToken extracts the JWT token from the request based on the configured source.
func extractToken(c *fursy.Context, source, param, authScheme string) string {
	switch source {
	case "header":
		auth := c.Request.Header.Get(param)
		if auth == "" {
			return ""
		}

		// If Authorization header and authScheme is set, extract token from scheme.
		if param == "Authorization" && authScheme != "" {
			prefix := authScheme + " "
			if !strings.HasPrefix(auth, prefix) {
				return ""
			}
			return auth[len(prefix):]
		}

		return auth

	case "query":
		return c.Request.URL.Query().Get(param)

	case "cookie":
		cookie, err := c.Request.Cookie(param)
		if err != nil {
			return ""
		}
		return cookie.Value

	default:
		return ""
	}
}

// validateClaim validates a specific claim against an expected value.
//
//nolint:gocognit // Claim validation requires checking multiple formats
func validateClaim(claims jwt.Claims, key, expected string) bool {
	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		// For custom claims, use GetExpirationTime, GetIssuer, GetAudience methods.
		switch key {
		case "iss":
			iss, err := claims.GetIssuer()
			return err == nil && iss == expected
		case "aud":
			aud, err := claims.GetAudience()
			if err != nil {
				return false
			}
			for _, a := range aud {
				if a == expected {
					return true
				}
			}
			return false
		}
		return false
	}

	// For MapClaims.
	value, ok := mapClaims[key]
	if !ok {
		return false
	}

	// Handle string claims.
	if str, ok := value.(string); ok {
		return str == expected
	}

	// Handle audience (can be string or []string).
	if key == "aud" {
		if audiences, ok := value.([]interface{}); ok {
			for _, aud := range audiences {
				if audStr, ok := aud.(string); ok && audStr == expected {
					return true
				}
			}
		}
	}

	return false
}

// defaultJWTErrorHandler is the default error handler for JWT validation failures.
func defaultJWTErrorHandler(c *fursy.Context, err error) error {
	// Return 401 Unauthorized for all JWT errors.
	return c.String(http.StatusUnauthorized, "Unauthorized: "+err.Error())
}

// JWTHelper provides helper functions for working with JWT tokens.
type JWTHelper struct{}

// GenerateToken generates a new JWT token with the provided claims and signing key.
//
// Example:
//
//	claims := jwt.MapClaims{
//	    "sub": "user123",
//	    "exp": time.Now().Add(15 * time.Minute).Unix(),
//	    "iat": time.Now().Unix(),
//	}
//	token, err := middleware.JWTHelper{}.GenerateToken(claims, []byte("secret"), "HS256")
func (JWTHelper) GenerateToken(claims jwt.Claims, signingKey interface{}, method string) (string, error) {
	// Security: Forbid "none" algorithm.
	if strings.EqualFold(method, jwtAlgoNone) {
		return "", ErrJWTNoneAlgo
	}

	var signingMethod jwt.SigningMethod
	switch method {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "HS384":
		signingMethod = jwt.SigningMethodHS384
	case "HS512":
		signingMethod = jwt.SigningMethodHS512
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	case "RS384":
		signingMethod = jwt.SigningMethodRS384
	case "RS512":
		signingMethod = jwt.SigningMethodRS512
	case "ES256":
		signingMethod = jwt.SigningMethodES256
	case "ES384":
		signingMethod = jwt.SigningMethodES384
	case "ES512":
		signingMethod = jwt.SigningMethodES512
	default:
		return "", fmt.Errorf("unsupported signing method: %s", method)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	return token.SignedString(signingKey)
}

// GenerateAccessToken generates a short-lived access token with standard claims.
// Best practice: 15-30 minutes expiration.
//
// Example:
//
//	token, err := middleware.JWTHelper{}.GenerateAccessToken(
//	    "user123",
//	    "my-app",
//	    []string{"my-users"},
//	    15*time.Minute,
//	    []byte("secret"),
//	    "HS256",
//	)
func (h JWTHelper) GenerateAccessToken(
	subject string,
	issuer string,
	audience []string,
	expiration time.Duration,
	signingKey interface{},
	method string,
) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		Issuer:    issuer,
		Audience:  audience,
		ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	return h.GenerateToken(claims, signingKey, method)
}
