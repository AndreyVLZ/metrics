package middleware

import (
	"bytes"
	"crypto/hmac"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/AndreyVLZ/metrics/internal/hash"
)

type hashWriter struct {
	rw     http.ResponseWriter
	buf    *bytes.Buffer
	status int
}

func newHashWriter(rw http.ResponseWriter) *hashWriter {
	buf := bytes.NewBuffer([]byte{})
	return &hashWriter{
		rw:     rw,
		buf:    buf,
		status: http.StatusOK,
	}
}

func (hw *hashWriter) Header() http.Header {
	return hw.rw.Header()
}

func (hw *hashWriter) WriteHeader(statusCode int) {
	hw.status = statusCode
}

func (hw *hashWriter) Write(p []byte) (int, error) {
	return hw.buf.Write(p)
}

type myReaderCloser struct {
	body io.ReadCloser
}

func Hash(key string, next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Выходим если key не задан
		if key == "" {
			// передаём управление хендлеру
			next.ServeHTTP(rw, req)
			return
		}

		// Если установлен Header проверяем MAC
		sha := req.Header.Get("HashSHA256")
		if sha != "" {
			bodyByte, err := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewReader(bodyByte))
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			if isValid, err := validMAC(sha, bodyByte, []byte(key)); err != nil && !isValid {
				http.Error(rw, "internal errpr", http.StatusInternalServerError)
				return
			}
		}

		hw := newHashWriter(rw)
		// меняем оригинальный http.ResponseWriter на новый

		// передаём управление хендлеру
		next.ServeHTTP(hw, req)

		// Вычисляет хеш и и передавает его в HTTP-заголовке
		if hw.buf.Len() > 0 && hw.status < 300 {
			sum, err := hash.SHA256(hw.buf.Bytes(), []byte(key))
			if err == nil {
				hw.Header().Set("HashSHA256", hex.EncodeToString(sum))
			}
		}
		rw.WriteHeader(hw.status)
		hw.buf.WriteTo(rw)
	}
}

func validMAC(messageMACStr string, message, key []byte) (bool, error) {
	expectedMAC, err := hash.SHA256(message, key)
	if err != nil {
		return false, err
	}

	messageMAC, err := hex.DecodeString(messageMACStr)
	if err != nil {
		return false, err
	}

	return hmac.Equal(messageMAC, expectedMAC), nil
}
