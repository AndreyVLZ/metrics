package grps

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/AndreyVLZ/metrics/internal/model"
	pb "github.com/AndreyVLZ/metrics/internal/proto"
	"github.com/AndreyVLZ/metrics/server/config"
	"google.golang.org/grpc"
)

type storager interface {
	AddBatch(ctx context.Context, arr []model.Metric) error
}

type Server struct {
	server  *grpc.Server
	mServer *metricServer
	cfg     *config.Config
}

func NewServer(cfg *config.Config, store storager, log *slog.Logger) *Server {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(Logging(log)),
	}

	return &Server{
		cfg:    cfg,
		server: grpc.NewServer(opts...),
		mServer: &metricServer{
			store: store,
		},
	}
}

func (s *Server) Start() error {
	listen, err := net.Listen("tcp", s.cfg.AddrGRPC)
	if err != nil {
		return fmt.Errorf("net listen: %w", err)
	}

	pb.RegisterMetricsServer(s.server, s.mServer)

	if err := s.server.Serve(listen); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

func (s *Server) Stop(_ context.Context) error {
	s.server.Stop()

	return nil
}

type metricServer struct {
	pb.UnimplementedMetricsServer
	store storager
}

func (ms *metricServer) AddBatch(ctx context.Context, req *pb.AddBatchRequest) (*pb.AddBatchResponse, error) {
	var response pb.AddBatchResponse

	if err := ms.store.AddBatch(ctx, buildArray(req.GetArr())); err != nil {
		response.Error = fmt.Sprintf("grpc addBatch: %v", err)

		return &response, nil
	}

	return &response, nil
}

func buildInfo(protoInfo *pb.Info) model.Info {
	mInfo := model.Info{
		MName: protoInfo.GetName(),
		MType: model.Type(protoInfo.GetType()),
	}

	return mInfo
}

func buildMetric(protoMetric *pb.Metric) model.Metric {
	met := model.Metric{
		Info: buildInfo(protoMetric.GetInfo()),
	}

	if met.MType == model.TypeCountConst {
		delta := protoMetric.GetDelta()
		met.Value.Delta = &delta
	}

	if met.MType == model.TypeGaugeConst {
		val := protoMetric.GetValue()
		met.Value.Val = &val
	}

	return met
}

func buildArray(protoArr []*pb.Metric) []model.Metric {
	arr := make([]model.Metric, len(protoArr))

	for i := range protoArr {
		arr[i] = buildMetric(protoArr[i])
	}

	return arr
}
