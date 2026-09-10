// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import "net/http"

// GET registers a type-safe GET handler. Deprecated: use router.GET() instead.
func GET[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodGet, path, adaptGenericHandler(handler))
}

// POST registers a type-safe POST handler. Deprecated: use router.POST() instead.
func POST[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodPost, path, adaptGenericHandler(handler))
}

// PUT registers a type-safe PUT handler. Deprecated: use router.PUT() instead.
func PUT[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodPut, path, adaptGenericHandler(handler))
}

// DELETE registers a type-safe DELETE handler. Deprecated: use router.DELETE() instead.
func DELETE[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodDelete, path, adaptGenericHandler(handler))
}

// PATCH registers a type-safe PATCH handler. Deprecated: use router.PATCH() instead.
func PATCH[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodPatch, path, adaptGenericHandler(handler))
}

// HEAD registers a type-safe HEAD handler. Deprecated: use router.HEAD() instead.
func HEAD[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodHead, path, adaptGenericHandler(handler))
}

// OPTIONS registers a type-safe OPTIONS handler. Deprecated: use router.OPTIONS() instead.
func OPTIONS[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodOptions, path, adaptGenericHandler(handler))
}
