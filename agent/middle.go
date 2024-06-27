package agent

import (
	"fmt"
	"log/slog"
	"net/http"
)

// Middleware для логирования.
type loggingRoundTripper struct {
	next http.RoundTripper
	log  *slog.Logger
}

// Имплементация http.RoundTripper.
func (l loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	log := l.log.With(
		slog.String("http method", req.Method),
		slog.String("url", req.URL.String()),
	)

	resp, err := l.next.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	log.Info("res", "statusCode", resp.StatusCode)

	return resp, nil
}

/*
type gzipCompress struct {
	next   http.RoundTripper
	buf    *bytes.Buffer
	writer *gzip.Writer
}

func (g *gzipCompress) Reset() { g.writer.Reset(g.buf) }

func NewGzipCompress(next http.RoundTripper) *gzipCompress {
	var buf bytes.Buffer
	return &gzipCompress{
		next:   next,
		buf:    &buf,
		writer: gzip.NewWriter(&buf),
	}
}

// Имплементация http.RoundTripper.
func (g *gzipCompress) RoundTrip(req *http.Request) (*http.Response, error) {
	body := req.Body

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	defer body.Close()

	if _, err := g.writer.Write(data); err != nil {
		return nil, fmt.Errorf("gzip write: %w", err)
	}

	if err := g.writer.Close(); err != nil {
		return nil, fmt.Errorf("gzipWriter close: %w", err)
	}

	defer g.Reset()

	fmt.Printf("compressed [%d] to [%d] bytes\r\n", len(data), g.buf.Len())

	reqNew, err := http.NewRequestWithContext(req.Context(), req.Method, req.URL.String(), g.buf)
	if err != nil {
		return nil, fmt.Errorf("err build request: %w", err)
	}

	reqNew.Header = req.Header
	req.Header.Set("Content-Encoding", "gzip")

	return g.next.RoundTrip(reqNew)
}
*/
