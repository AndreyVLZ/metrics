package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/AndreyVLZ/metrics/agent/client/grps"
	"github.com/AndreyVLZ/metrics/agent/client/http"
	"github.com/AndreyVLZ/metrics/agent/config"
	"github.com/AndreyVLZ/metrics/internal/model"
)

type preparer interface {
	Prepare(arr []model.Metric) (func(context.Context) error, error)
}

type Client struct {
	httpClient preparer
	grpcClient preparer
}

func New(cfg *config.Config, log *slog.Logger) *Client {
	return &Client{
		httpClient: http.NewClient(cfg, log),
		grpcClient: grps.NewClient(cfg, log),
	}
}

// Prepare Возвращает функцию для отправки данных.
// Сначало попытка отправки данных по HTTP, затем по GRPC.
// Функция отпраки возвращает объединеную ошибку от обоих попыток.
func (c *Client) Prepare(arr []model.Metric) (func(context.Context) error, error) {
	fnSendByHTTP, err := c.httpClient.Prepare(arr)
	if err != nil {
		return nil, fmt.Errorf("prepare data by http: %w", err)
	}

	fnSendByGRPC, err := c.grpcClient.Prepare(arr)
	if err != nil {
		return nil, fmt.Errorf("prepare data by gprc: %w", err)
	}

	fnSend := func(ctx context.Context) error {
		errs := make([]error, 2)

		// пробуем отправить данные по HTTP
		errs[0] = fnSendByHTTP(ctx)
		if errs[0] == nil {
			return nil
		}

		// пробуем отправить данные по GRPC
		errs[1] = fnSendByGRPC(ctx)
		if errs[1] == nil {
			return nil
		}

		return errors.Join(errs...)
	}

	return fnSend, nil
}
