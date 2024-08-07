package middleware

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/AndreyVLZ/metrics/pkg/hash"
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

// Хеширование данных.
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
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)

				return
			}

			req.Body = io.NopCloser(bytes.NewReader(bodyByte))

			if isValid, err := hash.ValidMAC(sha, bodyByte, []byte(key)); err != nil || !isValid {
				http.Error(rw, "internal errpr", http.StatusInternalServerError)

				return
			}
		}

		hw := newHashWriter(rw)

		// передаём управление хендлеру
		next.ServeHTTP(hw, req)

		// Вычисляет хеш и передавает его в HTTP-заголовке
		if hw.buf.Len() > 0 && hw.status < 300 {
			sum, err := hash.SHA256(hw.buf.Bytes(), []byte(key))
			if err == nil {
				hw.Header().Set("HashSHA256", hex.EncodeToString(sum))
			}
		}

		rw.WriteHeader(hw.status)

		if _, err := hw.buf.WriteTo(rw); err != nil {
			http.Error(rw, fmt.Sprintf("buf write: %s", err), http.StatusInternalServerError)
		}
	}
}
