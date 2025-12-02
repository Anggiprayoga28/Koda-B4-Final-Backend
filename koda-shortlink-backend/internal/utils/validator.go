package utils

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidURL(urlString string) bool {
	u, err := url.Parse(urlString)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host != ""
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

func ValidateStruct(data interface{}) error {
	return nil
}

func NormalizeURL(urlString string) (string, error) {
	urlString = strings.TrimSpace(urlString)

	if !strings.HasPrefix(urlString, "http://") && !strings.HasPrefix(urlString, "https://") {
		urlString = "https://" + urlString
	}

	u, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}
