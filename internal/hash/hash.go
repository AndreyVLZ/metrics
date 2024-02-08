package hash

import (
	"crypto/hmac"
	"crypto/sha256"
)

func SHA256(data []byte, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
