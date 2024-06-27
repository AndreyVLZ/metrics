package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
	}
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (int, error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}

	return c.zr.Close()
}

// Сжатие данных.
func Gzip(h http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := rw

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := req.Header.Get("Accept-Encoding")

		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// устанавливаем заголовок
			rw.Header().Set("Content-Encoding", "gzip")
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(rw)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := req.Header.Get("Content-Encoding")

		sendsGzip := strings.Contains(contentEncoding, "gzip")
		fmt.Printf("senGzip[%v]\n", sendsGzip)
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			creq, err := newCompressReader(req.Body)
			if err != nil {
				// w.WriteHeader(http.StatusInternalServerError)
				// fmt.Printf("errGzip: %v\n", err)
				http.Error(rw, err.Error(), http.StatusInternalServerError)

				return
			}
			// меняем тело запроса на новое
			req.Body = creq
			defer creq.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, req)
	}
}
