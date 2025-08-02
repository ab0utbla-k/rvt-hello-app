package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_ErrorResponse_ReturnsCorrectFormat verifies that errorResponse method creates
// consistent JSON error responses with proper status codes and handles different
// message types (string, map, number) correctly.
func Test_ErrorResponse_ReturnsCorrectFormat(t *testing.T) {
	t.Parallel()

	app := &application{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	tests := []struct {
		name         string
		status       int
		message      any
		expectStatus int
		expectError  any
	}{
		{
			name:         "StringMessage",
			status:       http.StatusBadRequest,
			message:      "invalid input",
			expectStatus: http.StatusBadRequest,
			expectError:  "invalid input",
		},
		{
			name:         "MapMessage",
			status:       http.StatusUnprocessableEntity,
			message:      map[string]string{"field": "error"},
			expectStatus: http.StatusUnprocessableEntity,
			expectError:  map[string]any{"field": "error"},
		},
		{
			name:         "IntMessage",
			status:       http.StatusInternalServerError,
			message:      500,
			expectStatus: http.StatusInternalServerError,
			expectError:  float64(500), // JSON unmarshals numbers as float64
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test", nil)

			app.errorResponse(w, r, tt.status, tt.message)
			assert.Equal(t, tt.expectStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response envelope
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectError, response["error"])
		})
	}
}

// Test_ServerErrorResponse_Returns500WithGenericMessage verifies that serverErrorResponse
// returns a 500 status with a generic error message, ensuring sensitive error details
// are not exposed to clients.
func Test_ServerErrorResponse_Returns500WithGenericMessage(t *testing.T) {
	t.Parallel()

	app := &application{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)

	sensitiveError := errors.New("database password is wrong")
	app.serverErrorResponse(w, r, sensitiveError)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response envelope
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotContains(t, response["error"], "password")
	assert.NotEmpty(t, response["error"])
}

// Test_FailedValidationResponse_Returns422 verifies that validation errors
// are properly structured in the JSON response, preserving field-specific
// error messages within the envelope format.
func Test_FailedValidationResponse_Returns422(t *testing.T) {
	t.Parallel()

	app := &application{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	tests := []struct {
		name             string
		validationErrors map[string]string
		expectedFields   map[string]string
	}{
		{
			name: "WithValidationErrors",
			validationErrors: map[string]string{
				"username":    "must contain only letters",
				"dateOfBirth": "must be a date before today",
			},
			expectedFields: map[string]string{
				"username":    "must contain only letters",
				"dateOfBirth": "must be a date before today",
			},
		},
		{
			name:             "EmptyValidationErrors",
			validationErrors: map[string]string{},
			expectedFields:   map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/test", nil)

			app.failedValidationResponse(w, r, tt.validationErrors)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response envelope
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// convert any to map
			errorMap, ok := response["error"].(map[string]any)
			require.True(t, ok)

			if len(tt.expectedFields) == 0 {
				assert.Empty(t, errorMap)
			} else {
				for field, expectedMsg := range tt.expectedFields {
					assert.Equal(t, expectedMsg, errorMap[field])
				}
			}
		})
	}
}

// Test_BadRequestResponse_ExposesErrorMessage verifies that badRequestResponse
// returns a 400 status with the actual error message, ensuring client-side
// debugging information is available to callers.
func Test_BadRequestResponse_ExposesErrorMessage(t *testing.T) {
	t.Parallel()

	app := &application{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/test", nil)

	validationErr := errors.New("validation failed: email format")

	app.badRequestResponse(w, r, validationErr)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response envelope
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "validation failed")
	assert.Contains(t, response["error"], "email format")
}
