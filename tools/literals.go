package tools

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
)

const (
	SocialProviders = "social-providers"
)

func ToNullString(s string) sql.NullString {
	return sql.NullString{Valid: true, String: s}
}

func ErrorNullString(err error) sql.NullString {
	if err == nil {
		return sql.NullString{Valid: true}
	}
	return sql.NullString{Valid: true, String: err.Error()}
}

func GenerateRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateRandomPassword(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "RP1." + hex.EncodeToString(bytes), nil
}
