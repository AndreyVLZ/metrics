package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/AndreyVLZ/metrics/agent/config"
	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/AndreyVLZ/metrics/pkg/crypto"
	"github.com/AndreyVLZ/metrics/pkg/hash"
)

const urlFormat = "http://%s/updates/" // Эндпоинт для отправки метрик.

type HTTPClient struct {
	client    *http.Client
	cfg       *config.Config
	urlToSend string
}

func NewClient(cfg *config.Config, log *slog.Logger) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Transport: &loggingRoundTripper{
				log:  log,
				next: http.DefaultTransport,
			},
		},
		cfg:       cfg,
		urlToSend: fmt.Sprintf(urlFormat, cfg.Addr),
	}
}

// Prepare Подготавливает данные для отправки из [arr].
// Возвращает функцию отправки и ошибку.
func (c *HTTPClient) Prepare(arr []model.Metric) (func(context.Context) error, error) {
	var header http.Header = make(map[string][]string)

	data, err := json.Marshal(model.BuildArrMetricJSON(arr))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	// хeшируем данные
	sum, err := hashed(c.cfg.Key, data)
	if err != nil {
		return nil, fmt.Errorf("req hashed: %w", err)
	}

	header.Set("HashSHA256", hex.EncodeToString(sum))

	// сжимаем данные
	dataCompress, err := gzipCompres(data)
	if err != nil {
		return nil, fmt.Errorf("req compress: %w", err)
	}

	header.Set("Content-Encoding", "gzip")

	// шифруем данные
	dataEncrypt, err := encrypt(c.cfg.PublicKey, dataCompress)
	if err != nil {
		return nil, fmt.Errorf("req crypto: %w", err)
	}

	fnSend := func(ctx context.Context) error {
		// собираем запрос
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.urlToSend, bytes.NewReader(dataEncrypt))
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}

		header.Set("Content-Type", "application/json")

		clientIP, err := getIP()
		if err != nil {
			return err
		}

		header.Set("X-Real-IP", clientIP.String())

		req.Header = header

		return c.do(req)
	}

	return fnSend, nil
}

// do Выполняет запрос.
func (c *HTTPClient) do(req *http.Request) error {
	// выполняем запрос
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("err to send request: %w", err)
	}

	// закрывает тело ответа
	if err = resp.Body.Close(); err != nil {
		return fmt.Errorf("err body close: %w", err)
	}

	return nil
}

// hashed Возвращает хеш.
func hashed(key, data []byte) ([]byte, error) {
	if len(key) == 0 {
		return data, nil
	}

	sum, err := hash.SHA256(data, key)
	if err != nil {
		return nil, fmt.Errorf("hash: %w", err)
	}

	return sum, nil
}

// gzipCompres Сжимает данные.
func gzipCompres(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if _, err := gzipWriter.Write(data); err != nil {
		return nil, fmt.Errorf("gzip write: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}

	return buf.Bytes(), nil
}

// encrypt Шифрует данные публичным ключом.
func encrypt(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	if publicKey == nil {
		return data, nil
	}

	cipher, err := crypto.Encrypt(publicKey, data)
	if err != nil {
		return nil, fmt.Errorf("encrypt len[%d] :%w", len(data), err)
	}

	return cipher, nil
}

func getIP() (net.IP, error) {
	addrs, err := net.LookupHost("localhost")
	if err != nil {
		return nil, fmt.Errorf("lookupHost: %w", err)
	}

	if len(addrs) == 0 {
		return nil, errors.New("no find addrs")
	}

	netIP := net.ParseIP(addrs[0])
	if netIP == nil {
		return nil, fmt.Errorf("parse IP [%s]: %w", addrs[0], err)
	}

	return netIP, nil
}

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
