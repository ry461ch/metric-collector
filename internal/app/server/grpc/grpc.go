package metricsgrpc

import (
	"google.golang.org/grpc"
	"io"

	config "github.com/ry461ch/metric-collector/internal/config/server"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	pb "github.com/ry461ch/metric-collector/internal/proto"
	"github.com/ry461ch/metric-collector/pkg/logging"
)

func New(config *config.Config, metricStorage Storage, fileWorker FileWorker) *MetricsGRPCServer {
	return &MetricsGRPCServer{
		metricStorage: metricStorage,
		config:        config,
		fileWorker:    fileWorker,
	}
}

type MetricsGRPCServer struct {
	pb.UnimplementedMetricsServer

	config        *config.Config
	metricStorage Storage
	fileWorker    FileWorker
}

func (mgs *MetricsGRPCServer) convert(m *pb.Metric) *metrics.Metric {
	if m == nil {
		return nil
	}

	var res metrics.Metric

	res.ID = m.Id

	switch m.Type {
	case pb.Metric_counter:
		res.MType = "counter"
		res.Delta = &m.Delta
	case pb.Metric_gauge:
		res.MType = "gauge"
		res.Value = &m.Value
	default:
		return nil
	}
	return &res
}

func (mgs *MetricsGRPCServer) PostMetrics(srv grpc.ClientStreamingServer[pb.Metric, pb.EmptyObject]) error {
	ctx := srv.Context()
	metricList := []metrics.Metric{}
	for {
		metric, err := srv.Recv()
		if err != nil && err == io.EOF {
			break
		}

		if err != nil || metric == nil {
			logging.Logger.Errorf("Failed while receiving metric %s", err.Error())
			return err
		}

		metricModel := mgs.convert(metric)
		if metricModel == nil {
			logging.Logger.Errorf("Failed while handling metric %s", metric.Id)
		}

		logging.Logger.Infof("Got metric %s", metricModel.ID)
		metricList = append(metricList, *metricModel)
	}

	err := mgs.metricStorage.SaveMetrics(ctx, metricList)
	if err != nil {
		srv.SendAndClose(&pb.EmptyObject{})
		return err
	}

	srv.SendAndClose(&pb.EmptyObject{})
	return nil
}
