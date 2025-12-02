package utils

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	DefaultShortCodeLength = 6
	charset                = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func GenerateShortCode(length int) (string, error) {
	if length <= 0 {
		length = DefaultShortCodeLength
	}

	result := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))

	for i := range result {
		num, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

func IsValidCustomSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 20 {
		return false
	}

	for _, char := range slug {
		if !isAlphanumeric(char) {
			return false
		}
	}

	return true
}

func isAlphanumeric(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}

func SanitizeCustomSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}
