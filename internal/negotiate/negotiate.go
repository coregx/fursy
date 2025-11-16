// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package negotiate provides RFC 9110 compliant content negotiation.
//
// This package is internal and not part of the public API.
// External code should use Context.NegotiateFormat() and Context.Negotiate() methods.
package negotiate

import (
	"sort"
	"strconv"
	"strings"
)

// mediaRange represents a single media range from an Accept header.
//
// RFC 9110 Section 12.5.1:
//
//	Accept = #( media-range [ weight ] )
//	media-range = ( "*/*" / ( type "/*" ) / ( type "/" subtype ) ) *( OWS ";" OWS parameter )
//	weight = OWS ";" OWS "q=" qvalue
//	qvalue = ( "0" [ "." 0*3DIGIT ] ) / ( "1" [ "." 0*3("0") ] )
type mediaRange struct {
	Type       string            // e.g., "application", "text", "*"
	Subtype    string            // e.g., "json", "html", "*"
	Quality    float64           // q-value (0.0 to 1.0), default 1.0
	Parameters map[string]string // Additional parameters (not including q)
}

// parseAccept parses an Accept header value into a sorted list of media ranges.
//
// The returned media ranges are sorted by:
//  1. Quality value (highest first)
//  2. Specificity (explicit > wildcard: "text/html" > "text/*" > "*/*")
//  3. Parameter count (more parameters = higher priority)
//
// Example:
//
//	ranges := parseAccept("text/html, application/json;q=0.9, */*;q=0.1")
//	// Returns: [text/html (q=1.0), application/json (q=0.9), */* (q=0.1)]
func parseAccept(accept string) []mediaRange {
	if accept == "" {
		return []mediaRange{{Type: "*", Subtype: "*", Quality: 1.0}}
	}

	parts := strings.Split(accept, ",")
	ranges := make([]mediaRange, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		mr := parseMediaRange(part)
		ranges = append(ranges, mr)
	}

	// Sort by RFC 9110 precedence rules.
	sort.Slice(ranges, func(i, j int) bool {
		a, b := ranges[i], ranges[j]

		// 1. Quality (highest first).
		if a.Quality != b.Quality {
			return a.Quality > b.Quality
		}

		// 2. Specificity (explicit > wildcard).
		specA := specificity(a)
		specB := specificity(b)
		if specA != specB {
			return specA > specB
		}

		// 3. Parameter count (more = higher priority).
		return len(a.Parameters) > len(b.Parameters)
	})

	return ranges
}

// parseQuality parses a quality value string and clamps it to [0.0, 1.0].
//
// Per RFC 9110, quality values are in range 0.0 to 1.0.
func parseQuality(value string) float64 {
	q, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 1.0 // Default on parse error.
	}

	// Clamp to [0.0, 1.0] per RFC 9110.
	if q < 0 {
		return 0
	}
	if q > 1 {
		return 1
	}

	return q
}

// parseMediaRange parses a single media range with parameters.
//
// Example: "application/json;q=0.9;charset=utf-8".
func parseMediaRange(s string) mediaRange {
	mr := mediaRange{
		Quality:    1.0, // Default quality.
		Parameters: make(map[string]string),
	}

	// Split media type from parameters.
	parts := strings.Split(s, ";")
	if len(parts) == 0 {
		return mr
	}

	// Parse media type (e.g., "application/json").
	mediaType := strings.TrimSpace(parts[0])
	typeParts := strings.Split(mediaType, "/")
	if len(typeParts) == 2 {
		mr.Type = strings.TrimSpace(typeParts[0])
		mr.Subtype = strings.TrimSpace(typeParts[1])
	} else {
		mr.Type = "*"
		mr.Subtype = "*"
	}

	// Parse parameters (q, charset, etc.).
	for i := 1; i < len(parts); i++ {
		param := strings.TrimSpace(parts[i])
		kv := strings.SplitN(param, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		if key == "q" {
			mr.Quality = parseQuality(value)
		} else {
			// Store other parameters.
			mr.Parameters[key] = value
		}
	}

	return mr
}

// specificity returns a numeric specificity score for sorting.
//
// Scores:
//   - 2: Explicit type and subtype (e.g., "application/json")
//   - 1: Wildcard subtype (e.g., "text/*")
//   - 0: Full wildcard (e.g., "*/*")
func specificity(mr mediaRange) int {
	if mr.Type == "*" {
		return 0
	}
	if mr.Subtype == "*" {
		return 1
	}
	return 2
}

// ContentType selects the best content type from offered types
// based on the Accept header.
//
// Returns the selected content type, or an empty string if no match found.
//
// Example:
//
//	offered := []string{"application/json", "text/html", "text/xml"}
//	accept := "text/html, application/json;q=0.9, */*;q=0.1"
//	result := ContentType(accept, offered)
//	// Returns: "text/html" (q=1.0 implicit)
func ContentType(accept string, offered []string) string {
	if len(offered) == 0 {
		return ""
	}

	ranges := parseAccept(accept)

	// Try to match each media range in priority order.
	for _, mr := range ranges {
		for _, offer := range offered {
			if matchMediaRange(mr, offer) {
				return offer
			}
		}
	}

	return ""
}

// matchMediaRange checks if an offered content type matches a media range.
//
// Examples:
//   - "*/*" matches everything
//   - "text/*" matches "text/html", "text/plain", etc.
//   - "application/json" matches only "application/json"
func matchMediaRange(mr mediaRange, offered string) bool {
	parts := strings.Split(offered, "/")
	if len(parts) != 2 {
		return false
	}

	offerType := strings.TrimSpace(parts[0])
	offerSubtype := strings.TrimSpace(parts[1])

	// Wildcard matching.
	if mr.Type == "*" {
		return true
	}

	if mr.Type != offerType {
		return false
	}

	if mr.Subtype == "*" {
		return true
	}

	return mr.Subtype == offerSubtype
}
