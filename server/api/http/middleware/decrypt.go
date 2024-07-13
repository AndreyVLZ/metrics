package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	mycrypto "github.com/AndreyVLZ/metrics/pkg/crypto"
)

// Decrypt Расшифровывает req.Body приватным ключом.
func Decrypt(privateKey *rsa.PrivateKey, next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if privateKey == nil {
			next.ServeHTTP(rw, req)

			return
		}

		bodyByte, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)

			return
		}

		cipher, err := mycrypto.Decrypt(privateKey, bodyByte)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)

			return
		}

		req.Body = io.NopCloser(bytes.NewReader(cipher))

		next.ServeHTTP(rw, req)
	}
}
