package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	//"os"
	//"strings"
	"testing"
	//"time"

	"github.com/ab0utbla-k/rvt-hello-app/internal/data"
	"github.com/ab0utbla-k/rvt-hello-app/internal/testutils"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	app    *application
	router *httprouter.Router
	db     *sql.DB
}

func (suite *APITestSuite) SetupSuite() {
	suite.db = testutils.SetupTestDB(suite.T())

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	suite.app = &application{
		logger: logger,
		models: data.NewModels(suite.db),
	}

	suite.router = httprouter.New()
	suite.router.MethodNotAllowed = http.HandlerFunc(suite.app.methodNotAllowedResponse)
	suite.router.NotFound = http.HandlerFunc(suite.app.notFoundResponse)

	suite.router.HandlerFunc(http.MethodPut, "/hello/:username", suite.app.saveUserHandler)
	suite.router.HandlerFunc(http.MethodGet, "/hello/:username", suite.app.getBirthdayMessageHandler)
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}

func (suite *APITestSuite) SetupTest() {
	testutils.CleanupDB(suite.T(), suite.db)
}

func (suite *APITestSuite) makeRequest(method, path string, body any) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(suite.T(), err)
		reader = bytes.NewReader(jsonBody)
	}

	r := httptest.NewRequest(method, path, reader)
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)
	return w
}

func (suite *APITestSuite) TestSaveUser_Success() {
	tests := []struct {
		name        string
		username    string
		dateOfBirth string
		expectCode  int
	}{
		{
			name:        "valid user",
			username:    "john",
			dateOfBirth: "1990-01-01",
			expectCode:  http.StatusNoContent,
		},
		{
			name:        "another valid user",
			username:    "alice",
			dateOfBirth: "1995-12-25",
			expectCode:  http.StatusNoContent,
		},
		{
			name:        "mixed case username",
			username:    "JohnDoe",
			dateOfBirth: "1988-06-15",
			expectCode:  http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			payload := map[string]string{
				"dateOfBirth": tt.dateOfBirth,
			}

			w := suite.makeRequest(http.MethodPut, "/hello/"+tt.username, payload)
			assert.Equal(suite.T(), tt.expectCode, w.Code)

			// Verify user was saved
			user, err := suite.app.models.Users.Get(tt.username)
			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), tt.username, user.Username)
		})
	}
}

func (suite *APITestSuite) TestSaveUser_InvalidJSON() {
	tests := []struct {
		name        string
		username    string
		body        string
		expectError string
	}{
		{
			name:        "invalid date format",
			username:    "john",
			body:        `{"dateOfBirth": "01-01-1990"}`,
			expectError: "invalid date format",
		},
		{
			name:        "empty date of birth",
			username:    "john",
			body:        `{"dateOfBirth": ""}`,
			expectError: "invalid date format",
		},
		{
			name:        "missing date of birth",
			username:    "john",
			body:        `{}`,
			expectError: "invalid date format",
		},
		{
			name:        "malformed JSON",
			username:    "john",
			body:        `{"dateOfBirth": "1990-01-01"`,
			expectError: "badly-formed JSON",
		},
		{
			name:        "empty body",
			username:    "john",
			body:        "",
			expectError: "body must not be empty",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodPut, "/hello/"+tt.username, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)
			assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

			var response envelope
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(suite.T(), err)
			assert.Contains(suite.T(), response["error"], tt.expectError)
		})
	}
}

func (suite *APITestSuite) TestSaveUser_ValidationFailure() {
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	tests := []struct {
		name        string
		username    string
		dateOfBirth string
		expectField string
	}{
		{
			name:        "username with numbers",
			username:    "john123",
			dateOfBirth: "1990-01-01",
			expectField: "username",
		},
		{
			name:        "username with spaces",
			username:    "john doe",
			dateOfBirth: "1990-01-01",
			expectField: "username",
		},
		{
			name:        "future date of birth",
			username:    "john",
			dateOfBirth: tomorrow,
			expectField: "dateOfBirth",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			payload := map[string]string{
				"dateOfBirth": tt.dateOfBirth,
			}
			escaped := url.PathEscape(tt.username)
			w := suite.makeRequest(http.MethodPut, "/hello/"+escaped, payload)
			assert.Equal(suite.T(), http.StatusUnprocessableEntity, w.Code)

			var response envelope
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(suite.T(), err)

			errorMap, ok := response["error"].(map[string]any)
			require.True(suite.T(), ok)
			assert.Contains(suite.T(), errorMap, tt.expectField)
		})
	}
}

func (suite *APITestSuite) TestSaveUser_UpdateExisting() {
	const (
		username     = "updateuser"
		originalDate = "1990-01-01"
		updatedDate  = "1995-06-15"
	)

	// Create original user
	payload := map[string]string{"dateOfBirth": originalDate}
	w := suite.makeRequest(http.MethodPut, "/hello/"+username, payload)
	assert.Equal(suite.T(), http.StatusNoContent, w.Code)

	// Update user
	payload = map[string]string{"dateOfBirth": updatedDate}
	w = suite.makeRequest(http.MethodPut, "/hello/"+username, payload)
	assert.Equal(suite.T(), http.StatusNoContent, w.Code)

	// Verify update
	user, err := suite.app.models.Users.Get(username)
	require.NoError(suite.T(), err)
	expectedDate, err := time.Parse("2006-01-02", updatedDate)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), expectedDate.Equal(user.DateOfBirth))
}

func (suite *APITestSuite) TestGetBirthdayMessage_ExistingUser() {
	const username = "testuser"
	today := time.Now()
	tomorrow := today.AddDate(0, 0, 1)

	tests := []struct {
		name         string
		dateOfBirth  time.Time
		expectSubstr string
	}{
		{
			name:         "birthday today",
			dateOfBirth:  time.Date(1990, today.Month(), today.Day(), 0, 0, 0, 0, time.UTC),
			expectSubstr: "Happy birthday!",
		},
		{
			name:         "birthday tomorrow",
			dateOfBirth:  time.Date(1990, tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC),
			expectSubstr: "Your birthday is in 1 day",
		},
		{
			name:         "birthday in future",
			dateOfBirth:  time.Date(1990, 12, 25, 0, 0, 0, 0, time.UTC),
			expectSubstr: "Your birthday is in",
		},
		{
			name:         "birthday yesterday (next year calculation)",
			dateOfBirth:  time.Date(1990, today.Month(), today.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1),
			expectSubstr: "Your birthday is in 364", // or 365 depending on leap year
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			payload := map[string]string{
				"dateOfBirth": tt.dateOfBirth.Format("2006-01-02"),
			}

			w := suite.makeRequest(http.MethodPut, "/hello/"+username, payload)
			require.Equal(suite.T(), http.StatusNoContent, w.Code)

			// Get birthday message
			w = suite.makeRequest(http.MethodGet, "/hello/"+username, nil)
			assert.Equal(suite.T(), http.StatusOK, w.Code)

			var response envelope
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(suite.T(), err)

			message, ok := response["message"].(string)
			require.True(suite.T(), ok, "message must be a string")
			assert.Contains(suite.T(), message, username)
			assert.Contains(suite.T(), message, tt.expectSubstr)
		})
	}
}

func (suite *APITestSuite) TestGetBirthdayMessage_NonExistentUser() {
	w := suite.makeRequest(http.MethodGet, "/hello/nonexistent", nil)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response envelope
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "could not be found")
}

func (suite *APITestSuite) TestGetBirthdayMessage_ResponseFormat() {
	const username = "formattest"
	dateOfBirth := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)

	payload := map[string]string{
		"dateOfBirth": dateOfBirth.Format("2006-01-02"),
	}

	w := suite.makeRequest(http.MethodPut, "/hello/"+username, payload)
	require.Equal(suite.T(), http.StatusNoContent, w.Code)

	// Get birthday message
	w = suite.makeRequest(http.MethodGet, "/hello/"+username, nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var response envelope
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	message, ok := response["message"].(string)
	require.True(suite.T(), ok, "message must be a string")
	assert.NotEmpty(suite.T(), message)
	assert.True(suite.T(), strings.HasPrefix(message, "Hello, "+username+"!"))
	assert.Len(suite.T(), response, 1, "response should only contain 'message' field")
}

func (suite *APITestSuite) TestMethodNotAllowed() {
	payload := map[string]string{
		"dateOfBirth": "1990-01-01",
	}

	w := suite.makeRequest(http.MethodPut, "/hello/testuser", payload)
	require.Equal(suite.T(), http.StatusNoContent, w.Code)

	// Test unsupported method
	w = suite.makeRequest(http.MethodPost, "/hello/testuser", payload)
	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)

	var response envelope
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "method is not supported")
}

func (suite *APITestSuite) TestNotFoundRoute() {
	w := suite.makeRequest(http.MethodGet, "/nonexistent", nil)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response envelope
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "could not be found")
}

//func Test_ReadJSON_InvalidJSON_ReturnsError(t *testing.T) {
//	t.Parallel()
//
//	tests := []struct {
//		name        string
//		payload     string
//		expectError string
//	}{
//		{
//			name:        "SyntaxError",
//			payload:     `{"dateOfBirth": "1990-01-01"`,
//			expectError: "badly-formed JSON",
//		},
//		{
//			name:        "EmptyBody",
//			payload:     "",
//			expectError: "body must not be empty",
//		},
//		{
//			name:        "UnknownField",
//			payload:     `{"unknown": "value"}`,
//			expectError: "unknown key",
//		},
//		{
//			name:        "MultipleValues",
//			payload:     `{"dateOfBirth": "1990-01-01"}{"extra": "value"}`,
//			expectError: "single JSON value",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			t.Parallel()
//
//			app := &application{}
//			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.payload))
//			w := httptest.NewRecorder()
//
//			var input struct {
//				DateOfBirth string `json:"dateOfBirth"`
//			}
//
//			err := app.readJSON(w, req, &input)
//
//			require.Error(t, err)
//			assert.Contains(t, err.Error(), tt.expectError)
//		})
//	}
//}
