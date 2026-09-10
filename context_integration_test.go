// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy_test

// Integration tests for Context.
//
// SSE, WebSocket, and DB stub tests removed in v0.6.0.
// These methods were removed from core Context.
// Use plugin-level helpers instead:
//   - stream.SSEUpgrade(c, handler)
//   - stream.WebSocketUpgrade(c, handler, opts)
//   - database.GetDB(c)
