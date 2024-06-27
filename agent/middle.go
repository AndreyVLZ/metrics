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
