package metricfilehelper

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/ry461ch/metric-collector/internal/helpers/metricmodelshelper"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

// Here we write all the data into one variable, because we store
// all data in memory, so we can assume that we have
// enough memory to duplicate our metric data
func SaveToStorage(filePath string, mStorage storage) error {
	metricList := []metrics.Metrics{}

	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
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

	metricmodelshelper.SaveMetrics(metricList, mStorage)
	return nil
}

func SaveToFile(filePath string, mStorage storage) error {
	metricList := []metrics.Metrics{}
	for metricName, val := range mStorage.GetGaugeValues() {
		metricList = append(metricList, metrics.Metrics{
			ID:    metricName,
			MType: "gauge",
			Value: &val,
		})
	}
	for metricName, val := range mStorage.GetCounterValues() {
		metricList = append(metricList, metrics.Metrics{
			ID:    metricName,
			MType: "counter",
			Delta: &val,
		})
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
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
