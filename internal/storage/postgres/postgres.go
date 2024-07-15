package pgstorage

import (
	"context"
	"database/sql"
	"strings"
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

func (pg *PGStorage) UpdateGaugeValue(ctx context.Context, key string, value float64) error {
	query := `INSERT INTO content.gauge_metrics (value, name) 
			  VALUES ($2, $1)
			  ON CONFLICT (name) DO UPDATE
			  SET value = $2, updated_at = CURRENT_TIMESTAMP;`
	_, err := pg.db.ExecContext(ctx, query, key, value)
	return err
}

func (pg *PGStorage) GetGaugeValue(ctx context.Context, key string) (float64, bool, error) {
	query := "SELECT value FROM content.gauge_metrics WHERE name = $1"
	row := pg.db.QueryRowContext(ctx, query, key)
	var value sql.NullFloat64
	err := row.Scan(&value)
    if err != nil {
        return 0, false, err
    }
	if !value.Valid {
		return 0, false, nil
	}
	return value.Float64, true, nil
}

func (pg *PGStorage) UpdateCounterValue(ctx context.Context, key string, value int64) error {
	query := `INSERT INTO content.counter_metrics (delta, name) 
			  VALUES ($2, $1)
			  ON CONFLICT (name) DO UPDATE
			  SET delta = counter_metrics.delta + $2, updated_at = CURRENT_TIMESTAMP;`
	_, err := pg.db.ExecContext(ctx, query, key, value)
	return err
}

func (pg *PGStorage) GetCounterValue(ctx context.Context, key string) (int64, bool, error) {
	query := "SELECT delta FROM content.counter_metrics WHERE name = $1"
	row := pg.db.QueryRowContext(ctx, query, key)
	var value sql.NullInt64
	err := row.Scan(&value)  // разбираем результат
    if err != nil {
        return 0, false, err
    }
	if !value.Valid {
		return 0, false, nil
	}
	return value.Int64, true, nil
}

func (pg *PGStorage) GetGaugeValues(ctx context.Context) (map[string]float64, error) {
	query := "SELECT value FROM content.gauge_metrics"
	rows, err := pg.db.QueryContext(ctx, query)
	if err != nil {
        return nil, err
    }
	defer rows.Close()

	data := map[string]float64{}
	for rows.Next() {
        var key string
		var val float64
        err = rows.Scan(key, val)
        if err != nil {
            return nil, err
        }

        data[key] = val
    }

    err = rows.Err()
    if err != nil {
        return nil, err
    }
    return data, nil
}

func (pg *PGStorage) GetCounterValues(ctx context.Context) (map[string]int64, error) {
	query := "SELECT value FROM content.counter_metrics"
	rows, err := pg.db.QueryContext(ctx, query)
	if err != nil {
        return nil, err
    }
	defer rows.Close()

	data := map[string]int64{}
	for rows.Next() {
        var key string
		var val int64
        err = rows.Scan(key, val)
        if err != nil {
            return nil, err
        }

        data[key] = val
    }

    err = rows.Err()
    if err != nil {
        return nil, err
    }
    return data, nil
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
