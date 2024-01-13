package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
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

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}
}

/*
type formatType uint

func (ft formatType) String() string {
	return []string{"gzip"}[ft]
}

const (
	GzipConst formatType = iota
)

const qValConst string = "q="

const defaultQVal int = 1

type compresInfo struct {
	format formatType
	q      int
}

type arrCompressInfo []compresInfo

func NewArrCompress(arr []string) arrCompressInfo {
	arrComp := make([]compresInfo, len(arr))
	for i := range arr {
		compInfo, err := newCompress(arr[i])
		if err != nil {
			arrComp = append(arrComp, compInfo)
		}
	}

	return arrComp
}

func (arr arrCompressInfo) Check(ft formatType) (compresInfo, bool) {
	for i := range arr {
		if arr[i].format == ft {
			return arr[i], true
		}
	}

	return compresInfo{}, false
}

func newCompress(val string) (compresInfo, error) {
	compres := compresInfo{q: defaultQVal}

	arr := strings.Split(val, ";")
	if err := compres.parse(arr); err != nil {
		return compresInfo{}, err
	}

	return compres, nil
}

func (ci *compresInfo) parse(arr []string) error {
	if err := ci.setFormat(arr[0]); err != nil {
		return err
	}

	if len(arr) >= 2 {
		ci.setQVal(arr[1])
	}

	return nil
}

func (ci *compresInfo) setFormat(formatStr string) error {
	switch formatStr {
	case GzipConst.String():
		ci.format = GzipConst
		return nil
	default:
		return errors.New("not support compress format")
	}
}

func (ci *compresInfo) setQVal(qValStr string) error {
	if qValStr[:2] != qValStr {
		return errors.New("qval not correct")
	}

	valFloat, err := strconv.ParseFloat(qValStr[2:], 64)
	if err != nil {
		return err
	}

	ci.q = int(valFloat)

	return err
}
*/
