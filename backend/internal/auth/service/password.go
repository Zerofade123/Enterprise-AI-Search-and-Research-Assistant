package service

import (
	"regexp"
	"strings"

	platformErrors "github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	upperRe   = regexp.MustCompile(`[A-Z]`)
	lowerRe   = regexp.MustCompile(`[a-z]`)
	digitRe   = regexp.MustCompile(`[0-9]`)
	specialRe = regexp.MustCompile(`[^A-Za-z0-9]`)
)

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func ValidatePasswordStrength(password string) error {
	if len(password) < 12 {
		return platformErrors.Wrap("auth.ValidatePasswordStrength", platformErrors.CodeValidation, "password must be at least 12 characters", nil)
	}
	if !upperRe.MatchString(password) || !lowerRe.MatchString(password) || !digitRe.MatchString(password) || !specialRe.MatchString(password) {
		return platformErrors.Wrap("auth.ValidatePasswordStrength", platformErrors.CodeValidation, "password must contain upper, lower, digit, and special character", nil)
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
