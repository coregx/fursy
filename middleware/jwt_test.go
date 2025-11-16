// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coregx/fursy"
	"github.com/golang-jwt/jwt/v5"
)

// Test constants.
const (
	testSecret   = "test-secret-key"
	testIssuer   = "test-issuer"
	testAudience = "test-audience"
	testSubject  = "user123"
)

// Helper function to generate test token.
func generateTestToken(claims jwt.Claims, key interface{}, method string) string {
	token, err := JWTHelper{}.GenerateToken(claims, key, method)
	if err != nil {
		panic(err)
	}
	return token
}

// Helper function to generate expired token.
func generateExpiredToken(key interface{}, method string) string {
	claims := jwt.MapClaims{
		"sub": testSubject,
		"exp": time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	}
	return generateTestToken(claims, key, method)
}

// Helper function to generate not-yet-valid token.
func generateNotYetValidToken(key interface{}, method string) string {
	claims := jwt.MapClaims{
		"sub": testSubject,
		"nbf": time.Now().Add(1 * time.Hour).Unix(), // Valid 1 hour from now
		"exp": time.Now().Add(2 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	return generateTestToken(claims, key, method)
}

// Helper function to generate valid token.
//
//nolint:unparam // Test helper uses HS256 by convention, parameter kept for flexibility
func generateValidToken(key interface{}, method string) string {
	claims := jwt.MapClaims{
		"sub": testSubject,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	return generateTestToken(claims, key, method)
}

func TestJWT_BasicAuth_HS256(t *testing.T) {
	secret := []byte(testSecret)
	token := generateValidToken(secret, "HS256")

	router := fursy.New()
	router.Use(JWT(secret))

	router.GET("/protected", func(c *fursy.Context) error {
		claims := c.Get(JWTContextKey).(jwt.MapClaims)
		sub := claims["sub"].(string)
		return c.String(200, "Hello, "+sub)
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	expected := "Hello, " + testSubject
	if rec.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, rec.Body.String())
	}
}

func TestJWT_MissingToken(t *testing.T) {
	router := fursy.New()
	router.Use(JWT([]byte(testSecret)))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestJWT_MalformedToken(t *testing.T) {
	router := fursy.New()
	router.Use(JWT([]byte(testSecret)))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	secret := []byte(testSecret)
	token := generateExpiredToken(secret, "HS256")

	router := fursy.New()
	router.Use(JWT(secret))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected status 401 for expired token, got %d", rec.Code)
	}
}

func TestJWT_NotYetValidToken(t *testing.T) {
	secret := []byte(testSecret)
	token := generateNotYetValidToken(secret, "HS256")

	router := fursy.New()
	router.Use(JWT(secret))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected status 401 for not-yet-valid token, got %d", rec.Code)
	}
}

func TestJWT_WrongSigningKey(t *testing.T) {
	// Generate token with one key.
	token := generateValidToken([]byte("key1"), "HS256")

	// Validate with different key.
	router := fursy.New()
	router.Use(JWT([]byte("key2")))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected status 401 for wrong signing key, got %d", rec.Code)
	}
}

func TestJWT_NoneAlgorithm_Forbidden(t *testing.T) {
	// Test that "none" algorithm is explicitly forbidden.
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for 'none' algorithm, but did not panic")
		}
	}()

	_ = JWTWithConfig(JWTConfig{
		SigningKey:    []byte("secret"),
		SigningMethod: "none",
	})
}

func TestJWT_AlgorithmConfusion_Prevention(t *testing.T) {
	// Generate token with HS256.
	secret := []byte(testSecret)
	tokenHS256 := generateValidToken(secret, "HS256")

	// Try to validate with RS256 expectation.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey:    &privateKey.PublicKey,
		SigningMethod: "RS256",
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tokenHS256)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected status 401 for algorithm confusion, got %d", rec.Code)
	}
}

func TestJWT_RS256(t *testing.T) {
	// Generate RSA key pair.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	claims := jwt.MapClaims{
		"sub": testSubject,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}

	token := generateTestToken(claims, privateKey, "RS256")

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey:    &privateKey.PublicKey,
		SigningMethod: "RS256",
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		claims := c.Get(JWTContextKey).(jwt.MapClaims)
		sub := claims["sub"].(string)
		return c.String(200, "Hello, "+sub)
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestJWT_ES256(t *testing.T) {
	// Generate ECDSA key pair.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	claims := jwt.MapClaims{
		"sub": testSubject,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}

	token := generateTestToken(claims, privateKey, "ES256")

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey:    &privateKey.PublicKey,
		SigningMethod: "ES256",
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestJWT_ValidateIssuer(t *testing.T) {
	secret := []byte(testSecret)

	// Generate token with issuer.
	claims := jwt.MapClaims{
		"sub": testSubject,
		"iss": testIssuer,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	token := generateTestToken(claims, secret, "HS256")

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey:     secret,
		ValidateIssuer: testIssuer,
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200 for valid issuer, got %d", rec.Code)
	}
}

func TestJWT_ValidateIssuer_Invalid(t *testing.T) {
	secret := []byte(testSecret)

	// Generate token with wrong issuer.
	claims := jwt.MapClaims{
		"sub": testSubject,
		"iss": "wrong-issuer",
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	token := generateTestToken(claims, secret, "HS256")

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey:     secret,
		ValidateIssuer: testIssuer,
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected status 401 for invalid issuer, got %d", rec.Code)
	}
}

func TestJWT_ValidateAudience(t *testing.T) {
	secret := []byte(testSecret)

	// Generate token with audience.
	claims := jwt.MapClaims{
		"sub": testSubject,
		"aud": testAudience,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	token := generateTestToken(claims, secret, "HS256")

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey:       secret,
		ValidateAudience: testAudience,
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200 for valid audience, got %d", rec.Code)
	}
}

func TestJWT_ValidateAudience_Array(t *testing.T) {
	secret := []byte(testSecret)

	// Generate token with multiple audiences.
	claims := jwt.MapClaims{
		"sub": testSubject,
		"aud": []string{"aud1", testAudience, "aud3"},
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	token := generateTestToken(claims, secret, "HS256")

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey:       secret,
		ValidateAudience: testAudience,
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200 for valid audience in array, got %d", rec.Code)
	}
}

func TestJWT_Skipper(t *testing.T) {
	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey: []byte(testSecret),
		Skipper: func(c *fursy.Context) bool {
			// Skip /health endpoint.
			return c.Request.URL.Path == "/health"
		},
	}))

	router.GET("/health", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "Protected")
	})

	// Request to /health (should be skipped).
	req := httptest.NewRequest("GET", "/health", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200 for skipped endpoint, got %d", rec.Code)
	}

	// Request to /protected without token (should fail).
	req = httptest.NewRequest("GET", "/protected", http.NoBody)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected status 401 for protected endpoint without token, got %d", rec.Code)
	}
}

func TestJWT_TokenFromQuery(t *testing.T) {
	secret := []byte(testSecret)
	token := generateValidToken(secret, "HS256")

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey:  secret,
		TokenLookup: "query:token",
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected?token="+token, http.NoBody)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200 for token from query, got %d", rec.Code)
	}
}

func TestJWT_TokenFromCookie(t *testing.T) {
	secret := []byte(testSecret)
	token := generateValidToken(secret, "HS256")

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey:  secret,
		TokenLookup: "cookie:jwt",
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "jwt", Value: token})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200 for token from cookie, got %d", rec.Code)
	}
}

func TestJWT_CustomClaims(t *testing.T) {
	type CustomClaims struct {
		jwt.RegisteredClaims
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}

	secret := []byte(testSecret)

	// Generate token with custom claims.
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   testSubject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: "123",
		Role:   "admin",
	}

	token := generateTestToken(claims, secret, "HS256")

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey: secret,
		Claims: func() jwt.Claims {
			return &CustomClaims{}
		},
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		claims := c.Get(JWTContextKey).(*CustomClaims)
		return c.String(200, "Role: "+claims.Role)
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if rec.Body.String() != "Role: admin" {
		t.Errorf("expected body 'Role: admin', got %q", rec.Body.String())
	}
}

func TestJWT_CustomErrorHandler(t *testing.T) {
	customErrorCalled := false

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey: []byte(testSecret),
		ErrorHandler: func(c *fursy.Context, err error) error {
			customErrorCalled = true
			return c.String(http.StatusForbidden, "Custom error: "+err.Error())
		},
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !customErrorCalled {
		t.Error("expected custom error handler to be called")
	}

	if rec.Code != 403 {
		t.Errorf("expected status 403 from custom error handler, got %d", rec.Code)
	}
}

func TestJWT_SuccessHandler(t *testing.T) {
	secret := []byte(testSecret)
	token := generateValidToken(secret, "HS256")

	successHandlerCalled := false

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey: secret,
		SuccessHandler: func(c *fursy.Context, claims jwt.Claims) error {
			successHandlerCalled = true
			// Store user info in context.
			mapClaims := claims.(jwt.MapClaims)
			c.Set(UserContextKey, mapClaims["sub"])
			return nil
		},
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		user := c.GetString(UserContextKey)
		return c.String(200, "User: "+user)
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !successHandlerCalled {
		t.Error("expected success handler to be called")
	}

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if rec.Body.String() != "User: "+testSubject {
		t.Errorf("expected body 'User: %s', got %q", testSubject, rec.Body.String())
	}
}

func TestJWT_SuccessHandler_ReturnsError(t *testing.T) {
	secret := []byte(testSecret)
	token := generateValidToken(secret, "HS256")

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey: secret,
		SuccessHandler: func(_ *fursy.Context, _ jwt.Claims) error {
			// Return error from success handler.
			return errors.New("custom validation failed")
		},
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 500 {
		t.Errorf("expected status 500 when success handler returns error, got %d", rec.Code)
	}
}

func TestJWT_AllowedAlgorithms(t *testing.T) {
	secret := []byte(testSecret)
	tokenHS256 := generateValidToken(secret, "HS256")

	// Allow both HS256 and HS384.
	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey:        secret,
		SigningMethod:     "HS256",
		AllowedAlgorithms: []string{"HS256", "HS384"},
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tokenHS256)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200 for allowed algorithm, got %d", rec.Code)
	}
}

func TestJWT_AllowedAlgorithms_Forbidden_None(t *testing.T) {
	// Test that "none" algorithm cannot be added to AllowedAlgorithms.
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when 'none' is in AllowedAlgorithms, but did not panic")
		}
	}()

	_ = JWTWithConfig(JWTConfig{
		SigningKey:        []byte("secret"),
		SigningMethod:     "HS256",
		AllowedAlgorithms: []string{"HS256", "none"},
	})
}

func TestJWTHelper_GenerateToken(t *testing.T) {
	secret := []byte(testSecret)

	claims := jwt.MapClaims{
		"sub": testSubject,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}

	token, err := JWTHelper{}.GenerateToken(claims, secret, "HS256")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	// Verify token can be parsed.
	parsed, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		t.Fatalf("failed to parse generated token: %v", err)
	}

	if !parsed.Valid {
		t.Error("generated token is not valid")
	}
}

func TestJWTHelper_GenerateToken_ForbidsNone(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": testSubject,
	}

	_, err := JWTHelper{}.GenerateToken(claims, []byte("secret"), "none")
	if !errors.Is(err, ErrJWTNoneAlgo) {
		t.Errorf("expected ErrJWTNoneAlgo, got %v", err)
	}
}

func TestJWTHelper_GenerateAccessToken(t *testing.T) {
	secret := []byte(testSecret)

	token, err := JWTHelper{}.GenerateAccessToken(
		testSubject,
		testIssuer,
		[]string{testAudience},
		15*time.Minute,
		secret,
		"HS256",
	)

	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	// Parse and verify claims.
	parsed, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		t.Fatalf("failed to parse access token: %v", err)
	}

	if !parsed.Valid {
		t.Error("access token is not valid")
	}

	claims := parsed.Claims.(jwt.MapClaims)

	if claims["sub"] != testSubject {
		t.Errorf("expected sub=%s, got %v", testSubject, claims["sub"])
	}

	if claims["iss"] != testIssuer {
		t.Errorf("expected iss=%s, got %v", testIssuer, claims["iss"])
	}
}

func TestJWT_PanicOnNilSigningKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil signing key, but did not panic")
		}
	}()

	_ = JWTWithConfig(JWTConfig{
		SigningKey: nil,
	})
}

func TestJWT_PanicOnInvalidTokenLookup(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid TokenLookup format, but did not panic")
		}
	}()

	_ = JWTWithConfig(JWTConfig{
		SigningKey:  []byte("secret"),
		TokenLookup: "invalid-format",
	})
}

func TestJWT_TokenStorage(t *testing.T) {
	secret := []byte(testSecret)
	token := generateValidToken(secret, "HS256")

	router := fursy.New()
	router.Use(JWT(secret))

	router.GET("/protected", func(c *fursy.Context) error {
		// Check that both token and claims are stored.
		storedToken := c.GetString(JWTTokenContextKey)
		if storedToken != token {
			t.Errorf("expected token %q, got %q", token, storedToken)
		}

		claims := c.Get(JWTContextKey)
		if claims == nil {
			t.Error("expected claims to be stored in context")
		}

		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestJWT_TokenFromHeader_WithoutBearer(t *testing.T) {
	secret := []byte(testSecret)
	token := generateValidToken(secret, "HS256")

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey:  secret,
		TokenLookup: "header:X-API-Token",
		AuthScheme:  "",
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("X-API-Token", token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200 for custom header, got %d", rec.Code)
	}
}

func TestJWT_ValidateClaim_WithRegisteredClaims(t *testing.T) {
	type CustomClaims struct {
		jwt.RegisteredClaims
		UserID string `json:"user_id"`
	}

	secret := []byte(testSecret)

	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   testSubject,
			Issuer:    testIssuer,
			Audience:  []string{testAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: "123",
	}

	token := generateTestToken(claims, secret, "HS256")

	router := fursy.New()
	router.Use(JWTWithConfig(JWTConfig{
		SigningKey:       secret,
		ValidateIssuer:   testIssuer,
		ValidateAudience: testAudience,
		Claims: func() jwt.Claims {
			return &CustomClaims{}
		},
	}))

	router.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200 for valid registered claims, got %d", rec.Code)
	}
}

func TestJWTHelper_GenerateToken_UnsupportedMethod(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": testSubject,
	}

	_, err := JWTHelper{}.GenerateToken(claims, []byte("secret"), "UNKNOWN")
	if err == nil {
		t.Error("expected error for unsupported signing method")
	}
}

func TestJWTHelper_GenerateToken_AllMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
		key    interface{}
	}{
		{"HS256", "HS256", []byte(testSecret)},
		{"HS384", "HS384", []byte(testSecret)},
		{"HS512", "HS512", []byte(testSecret)},
	}

	claims := jwt.MapClaims{
		"sub": testSubject,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := JWTHelper{}.GenerateToken(claims, tt.key, tt.method)
			if err != nil {
				t.Fatalf("failed to generate token with %s: %v", tt.method, err)
			}

			if token == "" {
				t.Errorf("expected non-empty token for %s", tt.method)
			}
		})
	}
}

func TestJWTHelper_GenerateToken_RSA_Methods(t *testing.T) {
	// Generate RSA key for RS256/RS384/RS512 tests.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		method string
	}{
		{"RS256", "RS256"},
		{"RS384", "RS384"},
		{"RS512", "RS512"},
	}

	claims := jwt.MapClaims{
		"sub": testSubject,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := JWTHelper{}.GenerateToken(claims, privateKey, tt.method)
			if err != nil {
				t.Fatalf("failed to generate token with %s: %v", tt.method, err)
			}

			if token == "" {
				t.Errorf("expected non-empty token for %s", tt.method)
			}
		})
	}
}

func TestJWTHelper_GenerateToken_ECDSA_Methods(t *testing.T) {
	tests := []struct {
		name   string
		method string
		curve  elliptic.Curve
	}{
		{"ES256", "ES256", elliptic.P256()},
		{"ES384", "ES384", elliptic.P384()},
		{"ES512", "ES512", elliptic.P521()},
	}

	claims := jwt.MapClaims{
		"sub": testSubject,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKey, err := ecdsa.GenerateKey(tt.curve, rand.Reader)
			if err != nil {
				t.Fatal(err)
			}

			token, err := JWTHelper{}.GenerateToken(claims, privateKey, tt.method)
			if err != nil {
				t.Fatalf("failed to generate token with %s: %v", tt.method, err)
			}

			if token == "" {
				t.Errorf("expected non-empty token for %s", tt.method)
			}
		})
	}
}
