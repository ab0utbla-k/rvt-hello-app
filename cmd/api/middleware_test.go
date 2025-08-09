package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoverPanic(t *testing.T) {
	t.Parallel()

	app := &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectPanic    bool
		expectStatus   int
		expectResponse string
	}{
		{
			name: "no panic",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			}),
			expectPanic:    false,
			expectStatus:   http.StatusOK,
			expectResponse: "ok",
		},
		{
			name: "panic with string",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			}),
			expectPanic:    true,
			expectStatus:   http.StatusInternalServerError,
			expectResponse: `{"error":"the server encountered a problem and could not process your request"}`,
		},
		{
			name: "panic with error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(fmt.Errorf("database connection failed"))
			}),
			expectPanic:    true,
			expectStatus:   http.StatusInternalServerError,
			expectResponse: `{"error":"the server encountered a problem and could not process your request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			middleware := app.recoverPanic(tt.handler)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test", nil)

			// Execute middleware
			middleware.ServeHTTP(w, r)

			assert.Equal(t, tt.expectStatus, w.Code)

			if tt.expectPanic {
				assert.Equal(t, "close", w.Header().Get("Connection"))
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				assert.Contains(t, w.Body.String(), tt.expectResponse)
			} else {
				assert.Equal(t, tt.expectResponse, w.Body.String())
			}
		})
	}
}
