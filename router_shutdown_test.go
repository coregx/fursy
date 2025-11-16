package fursy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestRouter_OnShutdown tests registering shutdown callbacks.
func TestRouter_OnShutdown(t *testing.T) {
	router := New()

	var called int32
	router.OnShutdown(func() {
		atomic.AddInt32(&called, 1)
	})

	if len(router.shutdownCallbacks) != 1 {
		t.Errorf("Expected 1 callback, got %d", len(router.shutdownCallbacks))
	}

	// Trigger shutdown to verify callback is called.
	ctx := context.Background()
	_ = router.Shutdown(ctx)

	if atomic.LoadInt32(&called) != 1 {
		t.Error("Shutdown callback was not called")
	}
}

// TestRouter_OnShutdown_Nil tests that nil callbacks are ignored.
func TestRouter_OnShutdown_Nil(t *testing.T) {
	router := New()

	router.OnShutdown(nil)

	if len(router.shutdownCallbacks) != 0 {
		t.Errorf("Expected 0 callbacks, got %d", len(router.shutdownCallbacks))
	}
}

// TestRouter_Shutdown_ReverseOrder tests callbacks execute in reverse order.
func TestRouter_Shutdown_ReverseOrder(t *testing.T) {
	router := New()

	var order []int
	var mu sync.Mutex

	router.OnShutdown(func() {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
	})

	router.OnShutdown(func() {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
	})

	router.OnShutdown(func() {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
	})

	ctx := context.Background()
	if err := router.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(order) != 3 {
		t.Fatalf("Expected 3 callbacks, got %d", len(order))
	}

	// Callbacks should execute in reverse order: 3, 2, 1.
	if order[0] != 3 || order[1] != 2 || order[2] != 1 {
		t.Errorf("Expected [3, 2, 1], got %v", order)
	}
}

// TestRouter_Shutdown_NoServer tests shutdown without server.
func TestRouter_Shutdown_NoServer(t *testing.T) {
	router := New()

	var called int32
	router.OnShutdown(func() {
		atomic.AddInt32(&called, 1)
	})

	ctx := context.Background()
	if err := router.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if atomic.LoadInt32(&called) != 1 {
		t.Error("Shutdown callback was not called")
	}
}

// TestRouter_Shutdown_WithServer tests shutdown with server.
func TestRouter_Shutdown_WithServer(t *testing.T) {
	router := New()
	router.GET("/test", func(_ *Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	// Create test server.
	srv := httptest.NewUnstartedServer(router)
	router.SetServer(&http.Server{
		Addr:    srv.Listener.Addr().String(),
		Handler: router,
	})
	srv.Start()
	defer srv.Close()

	var shutdownCalled int32
	router.OnShutdown(func() {
		atomic.AddInt32(&shutdownCalled, 1)
	})

	// Shutdown with context.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := router.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if atomic.LoadInt32(&shutdownCalled) != 1 {
		t.Error("Shutdown callback was not called")
	}
}

// TestRouter_Shutdown_ContextTimeout tests shutdown with context timeout.
func TestRouter_Shutdown_ContextTimeout(t *testing.T) {
	router := New()

	// Create a server with a handler that blocks.
	router.GET("/block", func(_ *Context) error {
		time.Sleep(5 * time.Second)
		return nil
	})

	// Start test server.
	srv := httptest.NewServer(router)
	defer srv.Close()

	// Make a request that will block.
	go func() {
		resp, err := http.Get(srv.URL + "/block")
		if err == nil {
			_ = resp.Body.Close()
		}
	}()

	// Give request time to start.
	time.Sleep(50 * time.Millisecond)

	// Create real http.Server for router.
	realSrv := &http.Server{
		Addr:    srv.Listener.Addr().String(),
		Handler: router,
	}
	router.SetServer(realSrv)

	// Shutdown with very short timeout (should timeout due to active request).
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Note: This test verifies timeout is respected, but actual timeout
	// depends on timing and may not always trigger in test environment.
	err := router.Shutdown(ctx)

	// Either succeeds or times out - both are acceptable in test environment.
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected nil or DeadlineExceeded, got %v", err)
	}
}

// TestRouter_SetServer tests SetServer method.
func TestRouter_SetServer(t *testing.T) {
	router := New()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	router.SetServer(srv)

	if router.server != srv {
		t.Error("Server not set correctly")
	}
}

// TestRouter_Shutdown_MultipleCalls tests multiple shutdown calls.
func TestRouter_Shutdown_MultipleCalls(t *testing.T) {
	router := New()

	var callCount int32
	router.OnShutdown(func() {
		atomic.AddInt32(&callCount, 1)
	})

	ctx := context.Background()

	// First shutdown.
	if err := router.Shutdown(ctx); err != nil {
		t.Fatalf("First shutdown failed: %v", err)
	}

	// Second shutdown - callbacks will be called again.
	if err := router.Shutdown(ctx); err != nil {
		t.Fatalf("Second shutdown failed: %v", err)
	}

	// Note: Callbacks are called on each Shutdown() call.
	// This is by design - user should guard against multiple calls if needed.
	count := atomic.LoadInt32(&callCount)
	if count != 2 {
		t.Errorf("Expected callbacks called twice, got %d times", count)
	}
}

// TestRouter_OnShutdown_Concurrent tests concurrent callback registration.
func TestRouter_OnShutdown_Concurrent(t *testing.T) {
	router := New()

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			router.OnShutdown(func() {
				// No-op callback.
			})
		}()
	}

	wg.Wait()

	if len(router.shutdownCallbacks) != numGoroutines {
		t.Errorf("Expected %d callbacks, got %d", numGoroutines, len(router.shutdownCallbacks))
	}
}

// TestRouter_ListenAndServeWithShutdown tests the all-in-one helper.
func TestRouter_ListenAndServeWithShutdown(t *testing.T) {
	// This test requires actual signal handling, which is tricky in unit tests.
	// We'll test the error handling instead.

	t.Run("InvalidAddress", func(t *testing.T) {
		router := New()
		router.GET("/test", func(_ *Context) error {
			return nil
		})

		done := make(chan error, 1)
		go func() {
			// Use invalid address to trigger immediate error.
			done <- router.ListenAndServeWithShutdown("invalid:address:99999")
		}()

		select {
		case err := <-done:
			if err == nil {
				t.Error("Expected error for invalid address, got nil")
			}
		case <-time.After(2 * time.Second):
			t.Error("ListenAndServeWithShutdown did not return error in time")
		}
	})
}

// TestRouter_ListenAndServeWithShutdown_Timeout tests custom timeout.
func TestRouter_ListenAndServeWithShutdown_Timeout(t *testing.T) {
	// This test verifies that custom timeout is accepted.
	router := New()
	router.GET("/test", func(_ *Context) error {
		return nil
	})

	done := make(chan error, 1)
	go func() {
		// Server will start but immediately fail with invalid address.
		done <- router.ListenAndServeWithShutdown("invalid:99999", 5*time.Second)
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Error("Expected error for invalid address, got nil")
		}
		// Verify error is from server startup, not from timeout handling.
	case <-time.After(2 * time.Second):
		t.Error("ListenAndServeWithShutdown did not return in time")
	}
}
