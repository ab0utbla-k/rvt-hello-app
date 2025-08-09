package data

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_GetBirthdayMessage(t *testing.T) {
	t.Parallel()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	tomorrow := today.AddDate(0, 0, 1)
	yesterday := today.AddDate(0, 0, -1)

	tests := []struct {
		name        string
		username    string
		dateOfBirth time.Time
		expectMsg   string
	}{
		{
			name:        "birthday today",
			username:    "john",
			dateOfBirth: time.Date(1990, today.Month(), today.Day(), 0, 0, 0, 0, time.UTC),
			expectMsg:   "Hello, john! Happy birthday!",
		},
		{
			name:        "birthday tomorrow",
			username:    "alice",
			dateOfBirth: time.Date(1990, tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC),
			expectMsg:   "Hello, alice! Your birthday is in 1 day(s)",
		},
		{
			name:        "birthday yesterday (next year)",
			username:    "bob",
			dateOfBirth: time.Date(1990, yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC),
			expectMsg:   "Hello, bob! Your birthday is in 364 day(s)",
		},
		{
			name:        "birthday in 10 days",
			username:    "carol",
			dateOfBirth: time.Date(1990, today.Month(), today.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 10),
			expectMsg:   "Hello, carol! Your birthday is in 10 day(s)",
		},
		{
			name:        "birthday in different month",
			username:    "dave",
			dateOfBirth: time.Date(1990, 12, 25, 0, 0, 0, 0, time.UTC),
			expectMsg:   "Hello, dave! Your birthday is in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user := &User{
				Username:    tt.username,
				DateOfBirth: tt.dateOfBirth,
			}

			message := user.GetBirthdayMessage()

			if tt.name == "birthday in different month" {
				assert.Contains(t, message, tt.expectMsg)
				assert.Contains(t, message, "day(s)")
			} else {
				assert.Equal(t, tt.expectMsg, message)
			}
		})
	}
}

func TestUser_GetBirthdayMessage_LeapYear(t *testing.T) {
	t.Parallel()

	// Test leap year edge case - Feb 29th birthday
	user := &User{
		Username:    "leapyear",
		DateOfBirth: time.Date(2000, 2, 29, 0, 0, 0, 0, time.UTC),
	}

	message := user.GetBirthdayMessage()

	assert.Contains(t, message, "Hello, leapyear!")
	assert.Contains(t, message, "birthday")
}
