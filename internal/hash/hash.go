package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func SHA256(data []byte, key []byte) ([]byte, error) {
	hash := hmac.New(sha256.New, key)

	_, err := hash.Write(data)
	if err != nil {
		return nil, fmt.Errorf("hash write: %w", err)
	}

	return hash.Sum(nil), nil
}

func ValidMAC(messageMACStr string, message, key []byte) (bool, error) {
	expectedMAC, err := SHA256(message, key)
	if err != nil {
		return false, fmt.Errorf("sha: %w", err)
	}

	messageMAC, err := hex.DecodeString(messageMACStr)
	if err != nil {
		return false, fmt.Errorf("hex decode: %w", err)
	}

	return hmac.Equal(messageMAC, expectedMAC), nil
}
