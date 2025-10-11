package middleware

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
var slugSafeChars = regexp.MustCompile(`[^a-z0-9-]`)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func ValidateEmail(email string) error {
	if email == "" {
		return &ValidationError{Field: "email", Message: "Email is required"}
	}
	if len(email) > 254 {
		return &ValidationError{Field: "email", Message: "Email too long"}
	}
	if !emailRegex.MatchString(email) {
		return &ValidationError{Field: "email", Message: "Invalid email format"}
	}
	return nil
}

func ValidateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: field + " is required"}
	}
	return nil
}

func ValidateLength(field, value string, min, max int) error {
	length := utf8.RuneCountInString(value)
	if length < min {
		return &ValidationError{Field: field, Message: field + " is too short"}
	}
	if length > max {
		return &ValidationError{Field: field, Message: field + " is too long"}
	}
	return nil
}

func ValidateAmount(amount float64) error {
	if amount <= 0 {
		return &ValidationError{Field: "amount", Message: "Amount must be greater than 0"}
	}
	if amount > 1000000 {
		return &ValidationError{Field: "amount", Message: "Amount too large"}
	}
	return nil
}

func SanitizeSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	slug = slugSafeChars.ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if slug == "" {
		slug = "untitled"
	}

	if len(slug) > 100 {
		slug = slug[:100]
	}

	return slug
}

func ValidateDonationType(donationType string) error {
	if donationType != "one-time" && donationType != "monthly" {
		return &ValidationError{Field: "donation_type", Message: "Invalid donation type"}
	}
	return nil
}
