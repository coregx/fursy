// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware provides security headers middleware.
package middleware

import (
	"fmt"
	"strconv"

	"github.com/coregx/fursy"
)

// Security header constant values (OWASP recommended defaults).
const (
	// ContentTypeNosniffValue is the recommended X-Content-Type-Options header value.
	ContentTypeNosniffValue = "nosniff"

	// XFrameOptionsSameOrigin allows framing only from same origin.
	XFrameOptionsSameOrigin = "SAMEORIGIN"

	// ReferrerPolicyStrictOrigin is the recommended Referrer-Policy value.
	ReferrerPolicyStrictOrigin = "strict-origin-when-cross-origin"
)

// SecureConfig defines the configuration for the Secure middleware.
type SecureConfig struct {
	// Skipper defines a function to skip the middleware.
	// Default: nil (middleware always executes)
	Skipper func(c *fursy.Context) bool

	// XSSProtection provides protection against cross-site scripting attack (XSS)
	// by setting the `X-XSS-Protection` header.
	// NOTE: This header is DEPRECATED and may introduce vulnerabilities.
	// Modern browsers use CSP instead.
	// Default: "" (not set, per OWASP recommendation)
	XSSProtection string

	// ContentTypeNosniff prevents the browser from MIME-sniffing a response away
	// from the declared content-type by setting the `X-Content-Type-Options` header.
	// Default: ContentTypeNosniffValue (OWASP recommended)
	ContentTypeNosniff string

	// XFrameOptions can be used to indicate whether or not a browser should
	// be allowed to render a page in a <frame>, <iframe> or <object>.
	// Sites can use this to avoid clickjacking attacks.
	// Values: DENY, XFrameOptionsSameOrigin, ALLOW-FROM uri
	// Default: XFrameOptionsSameOrigin (OWASP recommended)
	XFrameOptions string

	// HSTSMaxAge sets the `Strict-Transport-Security` header to indicate how long
	// (in seconds) browsers should remember that this site is only to be accessed using HTTPS.
	// This reduces your exposure to some SSL-stripping man-in-the-middle (MITM) attacks.
	// Only set if using HTTPS.
	// Default: 0 (header not set)
	// Recommended: 31536000 (1 year) or 63072000 (2 years)
	HSTSMaxAge int

	// HSTSExcludeSubdomains excludes subdomains when setting HSTS.
	// Default: false (includeSubDomains directive is set)
	HSTSExcludeSubdomains bool

	// HSTSPreloadEnabled adds the "preload" directive to HSTS header.
	// This allows the domain to be included in the HSTS preload list
	// maintained by Chrome (and used by Firefox and Safari).
	// WARNING: Requires HSTSMaxAge >= 31536000 (1 year)
	// Default: false
	HSTSPreloadEnabled bool

	// ContentSecurityPolicy sets the `Content-Security-Policy` header providing
	// security against cross-site scripting (XSS), clickjacking and other code
	// injection attacks resulting from execution of malicious content in the
	// trusted web page context.
	// Default: "" (not set - application specific)
	// Example: "default-src 'self'"
	ContentSecurityPolicy string

	// CSPReportOnly uses the `Content-Security-Policy-Report-Only` header instead
	// of `Content-Security-Policy`. This allows iterative updates of the content
	// security policy by only reporting the violations that would have occurred.
	// Default: false
	CSPReportOnly bool

	// ReferrerPolicy sets the `Referrer-Policy` header providing control over
	// what referrer information, sent in the Referer header, should be included.
	// Values: no-referrer, no-referrer-when-downgrade, same-origin,
	//         origin, strict-origin, origin-when-cross-origin,
	//         strict-origin-when-cross-origin, unsafe-url
	// Default: "strict-origin-when-cross-origin" (OWASP recommended)
	ReferrerPolicy string

	// CrossOriginEmbedderPolicy sets the `Cross-Origin-Embedder-Policy` header
	// to prevent a document from loading any cross-origin resources that don't
	// explicitly grant the document permission.
	// Values: unsafe-none, require-corp, credentialless
	// Default: "" (not set)
	CrossOriginEmbedderPolicy string

	// CrossOriginOpenerPolicy sets the `Cross-Origin-Opener-Policy` header.
	// COOP isolates your document from potential attackers by ensuring that
	// a document opened in a new context doesn't share a browsing context group
	// with the original document.
	// Values: unsafe-none, same-origin-allow-popups, same-origin
	// Default: "" (not set)
	CrossOriginOpenerPolicy string

	// CrossOriginResourcePolicy sets the `Cross-Origin-Resource-Policy` header.
	// CORP allows you to control which origins can load your resources.
	// Values: same-site, same-origin, cross-origin
	// Default: "" (not set)
	CrossOriginResourcePolicy string

	// PermissionsPolicy sets the `Permissions-Policy` header (formerly Feature-Policy).
	// This header allows you to control which features and APIs can be used
	// in the browser.
	// Default: "" (not set)
	// Example: "geolocation=(self), microphone=()"
	PermissionsPolicy string
}

// Secure returns a middleware that sets security headers following OWASP recommendations.
//
// The middleware sets the following headers by default:
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options: SAMEORIGIN
//   - Referrer-Policy: strict-origin-when-cross-origin
//
// Optional headers (require configuration):
//   - Strict-Transport-Security (HSTS) - only for HTTPS
//   - Content-Security-Policy (CSP) - application-specific
//   - Cross-Origin-* headers - application-specific
//   - Permissions-Policy - application-specific
//
// NOT included (deprecated/harmful per OWASP 2025):
//   - X-XSS-Protection - deprecated, may introduce vulnerabilities
//   - Expect-CT - deprecated
//
// Based on:
//   - OWASP Secure Headers Project (2025)
//   - Echo Secure middleware
//   - Fiber Helmet middleware
//   - Express.js Helmet (45% attack reduction)
//
// Example (sensible defaults):
//
//	router := fursy.New()
//	router.Use(middleware.Secure())
//
// Example (with HSTS for HTTPS):
//
//	router.Use(middleware.SecureWithConfig(middleware.SecureConfig{
//	    HSTSMaxAge: 31536000, // 1 year
//	    HSTSPreloadEnabled: true,
//	}))
//
// Example (with CSP):
//
//	router.Use(middleware.SecureWithConfig(middleware.SecureConfig{
//	    ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'",
//	}))
//
// Example (strict security):
//
//	router.Use(middleware.SecureWithConfig(middleware.SecureConfig{
//	    XFrameOptions: "DENY",
//	    HSTSMaxAge: 63072000, // 2 years
//	    HSTSPreloadEnabled: true,
//	    ContentSecurityPolicy: "default-src 'self'",
//	    ReferrerPolicy: "no-referrer",
//	    CrossOriginEmbedderPolicy: "require-corp",
//	    CrossOriginOpenerPolicy: "same-origin",
//	    CrossOriginResourcePolicy: "same-origin",
//	}))
func Secure() fursy.HandlerFunc {
	return SecureWithConfig(SecureConfig{})
}

// SecureWithConfig returns a middleware with custom security configuration.
//
//nolint:gocognit,gocyclo,cyclop // Security headers middleware has natural complexity due to multiple config options.
func SecureWithConfig(config SecureConfig) fursy.HandlerFunc {
	// Set defaults (OWASP recommended).
	if config.ContentTypeNosniff == "" {
		config.ContentTypeNosniff = ContentTypeNosniffValue
	}

	if config.XFrameOptions == "" {
		config.XFrameOptions = XFrameOptionsSameOrigin
	}

	if config.ReferrerPolicy == "" {
		config.ReferrerPolicy = ReferrerPolicyStrictOrigin
	}

	// NOTE: XSSProtection defaults to empty (not set) per OWASP 2025 recommendation.
	// Modern browsers use CSP instead, and X-XSS-Protection can introduce vulnerabilities.

	return func(c *fursy.Context) error {
		// Skip if Skipper returns true.
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		// X-Content-Type-Options: nosniff
		// Prevents MIME-sniffing attacks.
		if config.ContentTypeNosniff != "" {
			c.SetHeader("X-Content-Type-Options", config.ContentTypeNosniff)
		}

		// X-Frame-Options: SAMEORIGIN | DENY | ALLOW-FROM uri
		// Prevents clickjacking attacks.
		if config.XFrameOptions != "" {
			c.SetHeader("X-Frame-Options", config.XFrameOptions)
		}

		// Strict-Transport-Security (HSTS)
		// Only set if HSTSMaxAge > 0 (requires HTTPS).
		if config.HSTSMaxAge > 0 {
			hsts := "max-age=" + strconv.Itoa(config.HSTSMaxAge)

			if !config.HSTSExcludeSubdomains {
				hsts += "; includeSubDomains"
			}

			if config.HSTSPreloadEnabled {
				hsts += "; preload"
			}

			c.SetHeader("Strict-Transport-Security", hsts)
		}

		// Content-Security-Policy (CSP)
		// Application-specific, only set if configured.
		if config.ContentSecurityPolicy != "" {
			if config.CSPReportOnly {
				c.SetHeader("Content-Security-Policy-Report-Only", config.ContentSecurityPolicy)
			} else {
				c.SetHeader("Content-Security-Policy", config.ContentSecurityPolicy)
			}
		}

		// Referrer-Policy
		// Controls referrer information sent in Referer header.
		if config.ReferrerPolicy != "" {
			c.SetHeader("Referrer-Policy", config.ReferrerPolicy)
		}

		// Cross-Origin-Embedder-Policy (COEP)
		if config.CrossOriginEmbedderPolicy != "" {
			c.SetHeader("Cross-Origin-Embedder-Policy", config.CrossOriginEmbedderPolicy)
		}

		// Cross-Origin-Opener-Policy (COOP)
		if config.CrossOriginOpenerPolicy != "" {
			c.SetHeader("Cross-Origin-Opener-Policy", config.CrossOriginOpenerPolicy)
		}

		// Cross-Origin-Resource-Policy (CORP)
		if config.CrossOriginResourcePolicy != "" {
			c.SetHeader("Cross-Origin-Resource-Policy", config.CrossOriginResourcePolicy)
		}

		// Permissions-Policy (formerly Feature-Policy)
		if config.PermissionsPolicy != "" {
			c.SetHeader("Permissions-Policy", config.PermissionsPolicy)
		}

		// X-XSS-Protection (DEPRECATED)
		// NOTE: Only set if explicitly configured.
		// OWASP 2025: This header is deprecated and may introduce vulnerabilities.
		// Modern browsers use CSP instead.
		if config.XSSProtection != "" {
			c.SetHeader("X-XSS-Protection", config.XSSProtection)
		}

		return c.Next()
	}
}

// SecureDefaults returns a SecureConfig with OWASP recommended defaults.
//
// Includes:
//   - X-Content-Type-Options: ContentTypeNosniffValue
//   - X-Frame-Options: XFrameOptionsSameOrigin
//   - Referrer-Policy: ReferrerPolicyStrictOrigin
//
// Does NOT include (require explicit configuration):
//   - HSTS (requires HTTPS)
//   - CSP (application-specific)
//   - Cross-Origin-* headers (application-specific)
func SecureDefaults() SecureConfig {
	return SecureConfig{
		ContentTypeNosniff: ContentTypeNosniffValue,
		XFrameOptions:      XFrameOptionsSameOrigin,
		ReferrerPolicy:     ReferrerPolicyStrictOrigin,
	}
}

// SecureStrict returns a SecureConfig with strict security settings.
// Suitable for production HTTPS applications with no legacy browser support.
//
// WARNING: This configuration may break functionality if not carefully tested:
//   - CSP may block inline scripts/styles
//   - CORP may block cross-origin resources
//   - Frame blocking may affect embedded content
//
// Example:
//
//	router.Use(middleware.SecureWithConfig(middleware.SecureStrict()))
func SecureStrict() SecureConfig {
	return SecureConfig{
		ContentTypeNosniff:        ContentTypeNosniffValue,
		XFrameOptions:             "DENY",
		ReferrerPolicy:            "no-referrer",
		HSTSMaxAge:                63072000, // 2 years
		HSTSPreloadEnabled:        true,
		ContentSecurityPolicy:     "default-src 'self'",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "same-origin",
	}
}

// SecureWithCSP returns a SecureConfig with custom CSP policy.
// Convenience function for common CSP configurations.
//
// Example:
//
//	router.Use(middleware.SecureWithConfig(middleware.SecureWithCSP(
//	    "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
//	)))
func SecureWithCSP(csp string) SecureConfig {
	config := SecureDefaults()
	config.ContentSecurityPolicy = csp
	return config
}

// SecureWithHSTS returns a SecureConfig with HSTS enabled.
// Convenience function for HTTPS applications.
//
// Example (1 year HSTS):
//
//	router.Use(middleware.SecureWithConfig(middleware.SecureWithHSTS(31536000, false)))
//
// Example (2 years with preload):
//
//	router.Use(middleware.SecureWithConfig(middleware.SecureWithHSTS(63072000, true)))
func SecureWithHSTS(maxAge int, preload bool) SecureConfig {
	config := SecureDefaults()
	config.HSTSMaxAge = maxAge
	config.HSTSPreloadEnabled = preload
	return config
}

// BuildHSTSHeader builds the HSTS header value from configuration.
// Useful for testing or custom implementations.
func BuildHSTSHeader(maxAge int, includeSubdomains, preload bool) string {
	hsts := fmt.Sprintf("max-age=%d", maxAge)

	if includeSubdomains {
		hsts += "; includeSubDomains"
	}

	if preload {
		hsts += "; preload"
	}

	return hsts
}
