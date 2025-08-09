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

func TestReadJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		payload     string
		expectError string
	}{
		{
			name:        "SyntaxError",
			payload:     `{"dateOfBirth": "1990-01-01"`,
			expectError: "badly-formed JSON",
		},
		{
			name:        "EmptyBody",
			payload:     "",
			expectError: "body must not be empty",
		},
		{
			name:        "UnknownField",
			payload:     `{"unknown": "value"}`,
			expectError: "unknown key",
		},
		{
			name:        "MultipleValues",
			payload:     `{"dateOfBirth": "1990-01-01"}{"extra": "value"}`,
			expectError: "single JSON value",
		},
		{
			name:        "UnexpectedEOF",
			payload:     `{"dateOfBirth": "199`,
			expectError: "badly-formed JSON",
		},
		{
			name:        "TypeMismatch",
			payload:     `{"dateOfBirth": 123}`,
			expectError: "incorrect JSON type",
		},
		{
			name:        "ValidJSON",
			payload:     `{"dateOfBirth": "1990-01-01"}`,
			expectError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.payload))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			var input struct {
				DateOfBirth string `json:"dateOfBirth"`
			}

			err := app.readJSON(w, r, &input)

			if tt.expectError == "" {
				require.NoError(t, err)
				assert.Equal(t, "1990-01-01", input.DateOfBirth)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			}
		})
	}
}

func TestReadJSON_MaxBytesLimit(t *testing.T) {
	t.Parallel()

	app := &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	// Create payload larger than 1MB limit
	largePayload := `{"dateOfBirth": "` + strings.Repeat("x", 1_048_577) + `"}`

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largePayload))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	var input struct {
		DateOfBirth string `json:"dateOfBirth"`
	}

	err := app.readJSON(w, r, &input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not be larger than")
}
