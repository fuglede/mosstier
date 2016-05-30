package main

import (
	"crypto/rand"
	"encoding/base64"
)

// generatePassword generates a 25 byte long random password.
func generatePassword() (string, error) {
	b := make([]byte, 25)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// The slice should now contain random bytes instead of only zeroes.
	// Base64 it just to have something users can use.
	return base64.StdEncoding.EncodeToString(b), err
}
