// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Version represents an API version.
//
// FURSY supports semantic versioning (MAJOR.MINOR.PATCH) but commonly
// uses only major versions for API versioning (v1, v2, etc.).
//
// Example:
//
//	v1 := fursy.Version{Major: 1}
//	v2 := fursy.Version{Major: 2, Minor: 1}
type Version struct {
	Major int
	Minor int
	Patch int
}

// String returns the version as a string (e.g., "v1", "v2.1", "v3.2.1").
func (v Version) String() string {
	if v.Minor == 0 && v.Patch == 0 {
		return fmt.Sprintf("v%d", v.Major)
	}
	if v.Patch == 0 {
		return fmt.Sprintf("v%d.%d", v.Major, v.Minor)
	}
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Equal checks if two versions are equal.
func (v Version) Equal(other Version) bool {
	return v.Major == other.Major && v.Minor == other.Minor && v.Patch == other.Patch
}

// GreaterThan checks if this version is greater than another.
func (v Version) GreaterThan(other Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

// LessThan checks if this version is less than another.
func (v Version) LessThan(other Version) bool {
	return !v.GreaterThan(other) && !v.Equal(other)
}

// ParseVersion parses a version string into a Version struct.
//
// Supported formats:
//   - "v1" -> Version{Major: 1}
//   - "v2.1" -> Version{Major: 2, Minor: 1}
//   - "v3.2.1" -> Version{Major: 3, Minor: 2, Patch: 1}
//   - "1" -> Version{Major: 1}
//   - "2.1.0" -> Version{Major: 2, Minor: 1}
//
// Returns zero Version and false if parsing fails.
func ParseVersion(s string) (Version, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")

	parts := strings.Split(s, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return Version{}, false
	}

	// Parse major version.
	major, err := strconv.Atoi(parts[0])
	if err != nil || major < 0 {
		return Version{}, false
	}

	v := Version{Major: major}

	// Parse minor version if present.
	if len(parts) >= 2 {
		minor, err := strconv.Atoi(parts[1])
		if err != nil || minor < 0 {
			return Version{}, false
		}
		v.Minor = minor
	}

	// Parse patch version if present.
	if len(parts) >= 3 {
		patch, err := strconv.Atoi(parts[2])
		if err != nil || patch < 0 {
			return Version{}, false
		}
		v.Patch = patch
	}

	return v, true
}

// ExtractVersionFromPath extracts version from URL path.
//
// Supports patterns:
//   - /v1/users -> v1
//   - /api/v2/posts -> v2
//   - /api/v2.1/posts -> v2.1
//
// Returns zero Version and false if no version found.
func ExtractVersionFromPath(path string) (Version, bool) {
	// Match /vX, /vX.Y, or /vX.Y.Z in path.
	re := regexp.MustCompile(`/v(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:/|$)`)
	matches := re.FindStringSubmatch(path)

	if len(matches) < 2 {
		return Version{}, false
	}

	v := Version{}

	// Parse major.
	if matches[1] != "" {
		major, _ := strconv.Atoi(matches[1])
		v.Major = major
	}

	// Parse minor if present.
	if len(matches) > 2 && matches[2] != "" {
		minor, _ := strconv.Atoi(matches[2])
		v.Minor = minor
	}

	// Parse patch if present.
	if len(matches) > 3 && matches[3] != "" {
		patch, _ := strconv.Atoi(matches[3])
		v.Patch = patch
	}

	return v, true
}

// DeprecationInfo contains information about API deprecation.
//
// RFC 8594 defines the Sunset HTTP header to communicate deprecation.
type DeprecationInfo struct {
	// Version that is deprecated.
	Version Version

	// SunsetDate is when this version will be removed (RFC 8594).
	SunsetDate *time.Time

	// Message is a human-readable deprecation message.
	Message string

	// Link to migration guide or new version docs.
	Link string
}

// SetDeprecationHeaders sets deprecation headers on the response.
//
// RFC 8594 Sunset Header:
//
//	Sunset: Sat, 31 Dec 2025 23:59:59 GMT
//
// Also sets:
//   - Deprecation: true (draft standard)
//   - Link: <url>; rel="sunset" (RFC 8594)
//   - Warning: "299 - \"message\"" (RFC 7234)
func (d *DeprecationInfo) SetDeprecationHeaders(c *Context) {
	// Set Deprecation header (draft standard).
	c.SetHeader("Deprecation", "true")

	// Set Sunset header (RFC 8594) if date is set.
	if d.SunsetDate != nil {
		c.SetHeader("Sunset", d.SunsetDate.Format(time.RFC1123))
	}

	// Set Link header if provided.
	if d.Link != "" {
		c.SetHeader("Link", fmt.Sprintf("<%s>; rel=\"sunset\"", d.Link))
	}

	// Set Warning header (RFC 7234) with message.
	if d.Message != "" {
		warning := fmt.Sprintf("299 - \"API version %s is deprecated. %s\"", d.Version, d.Message)
		c.SetHeader("Warning", warning)
	}
}

// APIVersion returns the current API version from the request.
//
// Version is extracted in this order:
//  1. Api-Version header
//  2. URL path (/v1/, /v2/, etc.)
//
// Returns zero Version if no version found.
//
// Example:
//
//	version := c.APIVersion()
//	if version.Major == 1 {
//	    // Handle v1 request
//	}
func (c *Context) APIVersion() Version {
	// Try header first (Api-Version: 1 or Api-Version: v2.1).
	if header := c.Request.Header.Get("Api-Version"); header != "" {
		if v, ok := ParseVersion(header); ok {
			return v
		}
	}

	// Try path extraction (/api/v1/users).
	if v, ok := ExtractVersionFromPath(c.Request.URL.Path); ok {
		return v
	}

	// No version found.
	return Version{}
}

// RequireVersion is a middleware that requires a specific API version.
//
// If the request version doesn't match, returns 404 Not Found or 400 Bad Request.
//
// Example:
//
//	v1 := router.Group("/api/v1")
//	v1.Use(fursy.RequireVersion(fursy.Version{Major: 1}))
//
//	v2 := router.Group("/api/v2")
//	v2.Use(fursy.RequireVersion(fursy.Version{Major: 2}))
func RequireVersion(required Version) HandlerFunc {
	return func(c *Context) error {
		version := c.APIVersion()

		// If no version found in request, fail.
		if version.Major == 0 {
			return c.Problem(BadRequest("API version required. Use URL path (/api/v1/) or Api-Version header."))
		}

		// Check major version match.
		if version.Major != required.Major {
			msg := fmt.Sprintf("API version mismatch. Required: %s, Got: %s", required, version)
			return c.Problem(BadRequest(msg))
		}

		return c.Next()
	}
}

// DeprecateVersion is a middleware that marks a version as deprecated.
//
// Adds deprecation headers to all responses for this version.
//
// Example:
//
//	v1 := router.Group("/api/v1")
//	v1.Use(fursy.DeprecateVersion(fursy.DeprecationInfo{
//	    Version: fursy.Version{Major: 1},
//	    SunsetDate: &sunsetDate,
//	    Message: "Please migrate to v2",
//	    Link: "https://api.example.com/docs/v2-migration",
//	}))
func DeprecateVersion(info DeprecationInfo) HandlerFunc {
	return func(c *Context) error {
		// Set deprecation headers.
		info.SetDeprecationHeaders(c)

		// Continue processing.
		return c.Next()
	}
}
