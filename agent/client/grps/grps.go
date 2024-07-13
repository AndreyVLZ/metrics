package grps

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AndreyVLZ/metrics/agent/config"
	"github.com/AndreyVLZ/metrics/internal/model"
	pb "github.com/AndreyVLZ/metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	cfg  *config.Config
	mlog *slog.Logger
}

func NewClient(cfg *config.Config, log *slog.Logger) *GRPCClient {
	return &GRPCClient{
		cfg:  cfg,
		mlog: log,
	}
}

func (c *GRPCClient) Prepare(arr []model.Metric) (func(context.Context) error, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(Logging(c.mlog)),
	}

	fnSend := func(ctx context.Context) error {
		// устанавливаем соединение
		conn, err := grpc.NewClient(c.cfg.AddrGRPC, opts...)
		if err != nil {
			return fmt.Errorf("new grpc conn: %w", err)
		}

		defer conn.Close()

		mClient := pb.NewMetricsClient(conn)

		resp, err := mClient.AddBatch(ctx, &pb.AddBatchRequest{Arr: buildProtoArray(arr)})
		if err != nil {
			return fmt.Errorf("grpc addBatch: %w", err)
		}

		if resp.GetError() != "" {
			return errors.New(resp.GetError())
		}

		return nil
	}

	return fnSend, nil
}

// buildProtoType ...
func buildProtoType(mType string) pb.Info_Type {
	v, ok := pb.Info_Type_value[strings.ToUpper(mType)]
	if !ok {
		return 0
	}

	return pb.Info_Type(v)
}

// buildProtoArray ...
func buildProtoArray(arr []model.Metric) []*pb.Metric {
	pArr := make([]*pb.Metric, len(arr))

	for i := range arr {
		pArr[i] = &pb.Metric{
			Info: &pb.Info{
				Name: arr[i].MName,
				Type: buildProtoType(arr[i].MType.String()),
			},
		}

		if arr[i].Value.Val != nil {
			pArr[i].Value = *arr[i].Value.Val
		}

		if arr[i].Value.Delta != nil {
			pArr[i].Delta = *arr[i].Value.Delta
		}
	}

	return pArr
}
