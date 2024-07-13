package client

import (
	"context"
	"errors"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
)

type fakeClient struct {
	prErr error
	sErr  error
}

func (fc *fakeClient) Prepare(arr []model.Metric) (func(context.Context) error, error) {
	if fc.prErr != nil {
		return nil, fc.prErr
	}

	return func(_ context.Context) error {
		if fc.sErr != nil {
			return fc.sErr
		}

		return nil
	}, nil
}

func TestPrepare(t *testing.T) {
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		client := Client{
			httpClient: &fakeClient{},
			grpcClient: &fakeClient{},
		}

		arr := []model.Metric{}

		fnSend, err := client.Prepare(arr)
		if err != nil {
			t.Errorf("err client perpare: %v\n", err)
		}

		if err := fnSend(ctx); err != nil {
			t.Errorf("err fnSend: %v\n", err)
		}
	})

	t.Run("err prepare data by http", func(t *testing.T) {
		exErr := errors.New("prepare data err")

		client := Client{
			httpClient: &fakeClient{prErr: exErr},
			grpcClient: &fakeClient{},
		}

		arr := []model.Metric{}

		if _, err := client.Prepare(arr); err == nil {
			t.Error("want err")
		}
	})

	t.Run("err prepare data by grpc", func(t *testing.T) {
		exErr := errors.New("prepare data err")

		client := Client{
			httpClient: &fakeClient{},
			grpcClient: &fakeClient{prErr: exErr},
		}

		arr := []model.Metric{}

		if _, err := client.Prepare(arr); err == nil {
			t.Error("want err")
		}
	})

	t.Run("ok after err send data http", func(t *testing.T) {
		exErr := errors.New("err send data")

		client := Client{
			httpClient: &fakeClient{sErr: exErr},
			grpcClient: &fakeClient{},
		}

		arr := []model.Metric{}

		fnSend, err := client.Prepare(arr)
		if err != nil {
			t.Errorf("err client perpare: %v\n", err)
		}

		if err := fnSend(ctx); err != nil {
			t.Errorf("err fnSend: %v\n", err)
		}
	})

	t.Run("err send data http and grpc", func(t *testing.T) {
		exErr := errors.New("err send data")

		client := Client{
			httpClient: &fakeClient{sErr: exErr},
			grpcClient: &fakeClient{sErr: exErr},
		}

		arr := []model.Metric{}

		fnSend, err := client.Prepare(arr)
		if err != nil {
			t.Errorf("err client perpare: %v\n", err)
		}

		if err := fnSend(ctx); err == nil {
			t.Error("want err")
		}
	})
}
