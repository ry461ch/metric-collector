package fileworker

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/metricservice"
	"github.com/ry461ch/metric-collector/internal/storage"
)

type FileWorker struct {
	filePath string
	metricService *metricservice.MetricService
}

func New(filePath string, metricStorage storage.Storage) *FileWorker {
	return &FileWorker{filePath: filePath, metricService: metricservice.New(metricStorage)}
}

// Here we write all the data into one variable, because we store
// all data in memory, so we can assume that we have
// enough memory to duplicate our metric data
func (fw *FileWorker) ExportFromFile() error {
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

	fw.metricService.SaveMetrics(metricList)
	return nil
}

func (fw *FileWorker) ImportToFile() error {
	metricList := fw.metricService.ExtractMetrics()

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
