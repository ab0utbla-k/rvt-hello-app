package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     int
		data       envelope
		headers    http.Header
		wantStatus int
	}{
		{
			name:       "simple response",
			status:     http.StatusOK,
			data:       envelope{"message": "test"},
			headers:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "with custom headers",
			status:     http.StatusCreated,
			data:       envelope{"id": float64(123), "name": "test"},
			headers:    http.Header{"X-Custom": []string{"value"}},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "empty envelope",
			status:     http.StatusNoContent,
			data:       envelope{},
			headers:    nil,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "complex nested data",
			status:     http.StatusOK,
			data:       envelope{"user": map[string]interface{}{"name": "John", "age": float64(30)}},
			headers:    nil,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}
			w := httptest.NewRecorder()

			err := app.writeJSON(w, tt.status, tt.data, tt.headers)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			// Verify custom headers were set
			if tt.headers != nil {
				for key, values := range tt.headers {
					assert.Equal(t, values[0], w.Header().Get(key))
				}
			}

			// Verify JSON structure
			var response envelope
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tt.data, response)

			// Verify trailing newline
			assert.True(t, strings.HasSuffix(w.Body.String(), "\n"))
		})
	}
}
