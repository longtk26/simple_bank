package val

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString
)

// Validate String
func ValidateString(value string, min int, max int) error {
	n := len(value)
	if n < min || n > max {
		return fmt.Errorf("length must be between %d and %d", min, max)
	}

	return nil
}

// Validate Username
func ValidateUsername(username string) error {
	if err := ValidateString(username, 3, 50); err != nil {
		return err
	}

	if !isValidUsername(username) {
		return fmt.Errorf("username must contain only letters, numbers and underscores")
	}
	return nil
}

// Validate Fullname
func ValidateFullname(fullname string) error {
	return 	ValidateString(fullname, 3, 50)
}

// Validate Password
func ValidatePassword(password string) error {
	return ValidateString(password, 6, 50)
}

// Validate Email
func ValidateEmail(email string) error {
	if err := ValidateString(email, 3, 50); err != nil {
		return err
	}

	_, err := mail.ParseAddress(email)

	if err != nil {
		return fmt.Errorf("invalid email address")
	}
	return nil
}