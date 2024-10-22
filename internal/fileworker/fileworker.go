// Module for export and import metrics data to file
package fileworker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

// FileWorker для перемещения метрик между хранлкой и файлами
type FileWorker struct {
	filePath      string
	metricStorage Storage
}

// Init FileWorker instance
func New(filePath string, metricStorage Storage) *FileWorker {
	return &FileWorker{filePath: filePath, metricStorage: metricStorage}
}

// Export metrics from file to storage.
// Here we write all the data into one variable, because we store
// all data in memory, so we can assume that we have
// enough memory to duplicate our metric data
func (fw *FileWorker) ExportFromFile(ctx context.Context) error {
	if fw.metricStorage == nil {
		return errors.New("DB_NOT_INITIALIZED")
	}
	metricList := []metrics.Metric{}

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

	err = fw.metricStorage.SaveMetrics(ctx, metricList)
	if err != nil {
		return err
	}
	return nil
}

// Import metrics from storage to file.
func (fw *FileWorker) ImportToFile(ctx context.Context) error {
	if fw.metricStorage == nil {
		return errors.New("DB_NOT_INITIALIZED")
	}
	metricList, err := fw.metricStorage.ExtractMetrics(ctx)
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
