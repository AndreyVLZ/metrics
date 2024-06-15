package agent

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
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
		log.Error("logger", "error", err)

		return nil, fmt.Errorf("%w", err)
	}

	log.Info("res", "statusCode", resp.StatusCode)

	return resp, nil
}

type retryRoundTripper struct {
	next           http.RoundTripper
	maxRetries     int
	delayIncrement time.Duration
	log            *slog.Logger
	fnBuildReq     func(context.Context, string, io.Reader) (*http.Request, error)
}

func (rr retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		res *http.Response
		err error
	)

	delay := rr.delayIncrement

	for attempts := 1; attempts <= rr.maxRetries; attempts++ {
		res, err = rr.next.RoundTrip(req)
		if err == nil && res.StatusCode < http.StatusInternalServerError {
			break
		}

		rr.log.Debug("retry", "попытка", attempts, "timeout", delay, "err", err)

		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(delay):
			delay += rr.delayIncrement

			req, err = rr.fnBuildReq(req.Context(), req.URL.String(), req.Body)
			if err != nil {
				return nil, fmt.Errorf("retry build request: %w", err)
			}
		}
	}

	return res, err
}
