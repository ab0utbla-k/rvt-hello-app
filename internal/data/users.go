package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ab0utbla-k/rvt-hello-app/internal/validator"
)

type User struct {
	Username    string    `json:"username"`
	DateOfBirth time.Time `json:"dateOfBirth"`
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(validator.Matches(user.Username, validator.UserRX), "username", "must contain only letters")
	v.Check(!user.DateOfBirth.IsZero(), "dateOfBirth", "must be provided")
	v.Check(user.DateOfBirth.Before(time.Now()), "dateOfBirth", "must be in the past")
}

type UserModel struct {
	DB *sql.DB
}

func (u UserModel) Insert(user *User) error {
	query := `
        INSERT INTO users (username, date_of_birth)
		VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE SET date_of_birth = EXCLUDED.date_of_birth`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := u.DB.ExecContext(ctx, query, user.Username, user.DateOfBirth)
	return err
}

func (u UserModel) Get(username string) (*User, error) {
	query := "SELECT username, date_of_birth FROM users WHERE username = $1"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	err := u.DB.QueryRowContext(ctx, query, username).Scan(
		&user.Username,
		&user.DateOfBirth,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u *User) GetBirthdayMessage() string {
	now := time.Now()
	location := now.Location()

	// Today's date at midnight in local timezone
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)

	thisYearBirthday := time.Date(today.Year(), u.DateOfBirth.Month(), u.DateOfBirth.Day(), 0, 0, 0, 0, time.UTC)

	if thisYearBirthday.Before(today) {
		thisYearBirthday = thisYearBirthday.AddDate(1, 0, 0)
	}

	daysUntilBirthday := int(thisYearBirthday.Sub(today).Hours() / 24)

	if daysUntilBirthday == 0 {
		return fmt.Sprintf("Hello, %s! Happy birthday!", u.Username)
	}

	return fmt.Sprintf("Hello, %s! Your birthday is in %d day(s)", u.Username, daysUntilBirthday)
}
