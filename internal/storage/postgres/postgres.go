package pgstorage

import (
	"context"
	"database/sql"
	"strings"
	"errors"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

type PGStorage struct {
	db	*sql.DB
}

func getDDL() string {
	return `
		CREATE SCHEMA IF NOT EXISTS content;
		CREATE TABLE IF NOT EXISTS content.gauge_metrics (
			name VARCHAR(255) PRIMARY KEY,
			value DOUBLE PRECISION NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS gauge_metrics_created_at_idx ON content.gauge_metrics(created_at);
		CREATE INDEX IF NOT EXISTS gauge_metrics_updated_at_idx ON content.gauge_metrics(updated_at);


		CREATE TABLE IF NOT EXISTS content.counter_metrics (
			name VARCHAR(255) PRIMARY KEY,
			delta BIGINT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS counter_metrics_created_at_idx ON content.counter_metrics(created_at);
		CREATE INDEX IF NOT EXISTS counter_metrics_updated_at_idx ON content.counter_metrics(updated_at);
	`
}

func NewPGStorage(ctx context.Context, DBDsn string) (*PGStorage, error) {
	db, err := sql.Open("pgx", DBDsn)

	if err != nil {
		return nil, err
	}

	requests := strings.Split(getDDL(), ";")

	for _, request := range requests {
		if request != "" {
			_, err := db.ExecContext(ctx, request)
			if err != nil {
				return nil, err
			}
		}
	}
	return &PGStorage{db: db}, nil
}

func (pg *PGStorage) Close() {
	defer pg.db.Close()
}

func (pg *PGStorage) Ping(ctx context.Context) bool {
	if err := pg.db.PingContext(ctx); err != nil {
		return false
    }
	return true
}


func (pg *PGStorage) SaveMetrics(ctx context.Context, metricList []metrics.Metric) error {
	gaugeMetrics := map[string]float64{}
	counterMetrics := map[string]int64{}

	// prepare arrays
	for _, metric := range metricList {
		if metric.ID == "" {
			return errors.New("INVALID_METRIC")
		}
		if metric.MType == "" {
			return errors.New("INVALID_METRIC")
		}

		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				return errors.New("INVALID_METRIC")
			}
			gaugeMetrics[metric.ID] = *metric.Value
		case "counter":
			if metric.Delta == nil {
				return errors.New("INVALID_METRIC")
			}
			counterMetrics[metric.ID] = *metric.Delta
		default:
			return errors.New("INVALID_METRIC")
		}
	}

	// begin trx
	tx, err := pg.db.Begin()
    if err != nil {
        return err
    }

	// insert gauge values
	gaugeQuery := `INSERT INTO content.gauge_metrics (name, value) 
			  VALUES ($1, $2)
			  ON CONFLICT (name) DO UPDATE
			  SET value = $2, updated_at = CURRENT_TIMESTAMP;`
	stmt, err := tx.PrepareContext(ctx, gaugeQuery)
	if err != nil {
        return err
    }
	for key, val := range gaugeMetrics {
		_, err := stmt.ExecContext(ctx, key, val)
		if err != nil {
			return err
		}
	}

	// insert counter values
	counterQuery := `INSERT INTO content.counter_metrics (name, delta) 
			  VALUES ($1, $2)
			  ON CONFLICT (name) DO UPDATE
			  SET delta = counter_metrics.delta + $2, updated_at = CURRENT_TIMESTAMP;`
	stmt, err = tx.PrepareContext(ctx, counterQuery)
	if err != nil {
		return err
	}
	for key, val := range counterMetrics {
		_, err := stmt.ExecContext(ctx, key, val)
		if err != nil {
			return err
		}
	}

	// commit trx
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (pg *PGStorage) ExtractMetrics(ctx context.Context) ([]metrics.Metric, error) {
	metricList := []metrics.Metric{}
	
	// get gauge metrics
	getGaugeQuery := "SELECT name, value FROM content.gauge_metrics"
	rows, err := pg.db.QueryContext(ctx, getGaugeQuery)
	if err != nil {
        return nil, err
    }
	defer rows.Close()

	for rows.Next() {
        var key string
		var val float64
        err = rows.Scan(key, val)
        if err != nil {
            return nil, err
        }

        metricList = append(metricList, metrics.Metric{
			ID:    key,
			MType: "gauge",
			Value: &val,
		})
    }

    err = rows.Err()
    if err != nil {
        return nil, err
    }

	// get counter metrics
	getCounterQuery := "SELECT value FROM content.counter_metrics"
	rows, err = pg.db.QueryContext(ctx, getCounterQuery)
	if err != nil {
        return nil, err
    }
	defer rows.Close()

	for rows.Next() {
        var key string
		var val int64
        err = rows.Scan(key, val)
        if err != nil {
            return nil, err
        }

        metricList = append(metricList, metrics.Metric{
			ID:    key,
			MType: "counter",
			Delta: &val,
		})
    }

    err = rows.Err()
    if err != nil {
        return nil, err
    }

	return metricList, nil
}

func (pg *PGStorage) GetMetric(ctx context.Context, metric *metrics.Metric) error {
	switch (metric.MType) {
	case "gauge":
		query := "SELECT value FROM content.gauge_metrics WHERE name = $1"
		row := pg.db.QueryRowContext(ctx, query, metric.ID)
		var value sql.NullFloat64
		err := row.Scan(&value)
		if err != nil {
			return err
		}
		if !value.Valid {
			return errors.New("NOT_FOUND")
		}
		metric.Value = &value.Float64
	case "counter":
		query := "SELECT delta FROM content.counter_metrics WHERE name = $1;"
		row := pg.db.QueryRowContext(ctx, query, metric.ID)
		var value sql.NullInt64
		err := row.Scan(&value)
		if err != nil {
			return err
		}
		if !value.Valid {
			return errors.New("NOT_FOUND")
		}
		metric.Delta = &value.Int64
	default:
		return errors.New("INVALID_METRIC_TYPE")
	}
	return nil
}
