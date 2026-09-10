// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import "net/http"

// Deprecated: Use router.GET[Req, Res]() instead. Will be removed in v1.0.0.
func GET[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodGet, path, adaptGenericHandler(handler))
}

// Deprecated: Use router.POST[Req, Res]() instead. Will be removed in v1.0.0.
func POST[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodPost, path, adaptGenericHandler(handler))
}

// Deprecated: Use router.PUT[Req, Res]() instead. Will be removed in v1.0.0.
func PUT[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodPut, path, adaptGenericHandler(handler))
}

// Deprecated: Use router.DELETE[Req, Res]() instead. Will be removed in v1.0.0.
func DELETE[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodDelete, path, adaptGenericHandler(handler))
}

// Deprecated: Use router.PATCH[Req, Res]() instead. Will be removed in v1.0.0.
func PATCH[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodPatch, path, adaptGenericHandler(handler))
}

// Deprecated: Use router.HEAD[Req, Res]() instead. Will be removed in v1.0.0.
func HEAD[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodHead, path, adaptGenericHandler(handler))
}

// Deprecated: Use router.OPTIONS[Req, Res]() instead. Will be removed in v1.0.0.
func OPTIONS[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodOptions, path, adaptGenericHandler(handler))
}
