package validation

import (
	"net/mail"
	"strings"
)

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	return strings.EqualFold(addr.Address, email)
}

func IsBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}
