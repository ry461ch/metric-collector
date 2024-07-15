package fileworker

import (
	"bytes"
	"encoding/json"
	"os"
	"context"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/metricservice"
)

type FileWorker struct {
	filePath string
	metricService *metricservice.MetricService
}

func New(filePath string, metricService *metricservice.MetricService) *FileWorker {
	return &FileWorker{filePath: filePath, metricService: metricService}
}

// Here we write all the data into one variable, because we store
// all data in memory, so we can assume that we have
// enough memory to duplicate our metric data
func (fw *FileWorker) ExportFromFile(ctx context.Context) error {
	metricList := []metrics.Metrics{}

	file, err := os.OpenFile(fw.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(file)
	if err != nil {
		return err
	}

	data := buf.Bytes()
	if len(data) == 0 {
		return nil
	}

	err = json.Unmarshal(data, &metricList)
	if err != nil {
		return err
	}

	err = fw.metricService.SaveMetrics(ctx, metricList)
	if err != nil {
		return err
	}
	return nil
}

func (fw *FileWorker) ImportToFile(ctx context.Context) error {
	metricList, err := fw.metricService.ExtractMetrics(ctx)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(fw.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Here we write all the data into one variable, because we store
	// all data in memory, so we can assume that we have
	// enough memory to duplicate our metric data
	data, err := json.Marshal(metricList)
	if err != nil {
		return err
	}
	file.Write(data)

	return nil
}
