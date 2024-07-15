package pgstorage

import (
	"context"
	"database/sql"
	"strings"
	"log"
)

type PGStorage struct {
	db	*sql.DB
}

func getDDL() string {
	return `
		CREATE SCHEMA IF NOT EXISTS content;
		CREATE TABLE IF NOT EXISTS content.metrics (
			name VARCHAR(255),
			type VARCHAR(255) NOT NULL,
			value DOUBLE PRECISION NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (name, type)
		);

		CREATE INDEX IF NOT EXISTS metrics_name_idx ON content.metrics(name);
		CREATE INDEX IF NOT EXISTS metrics_type_idx ON content.metrics(type);
		CREATE INDEX IF NOT EXISTS metrics_created_at_idx ON content.metrics(created_at);
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
	query := "UPDATE content.metrics SET value = $2 WHERE name = $1 AND type = 'gauge'"
	log.Println(query, key, value)
	_, err := pg.db.ExecContext(ctx, query, key, value)
	return err
}

func (pg *PGStorage) GetGaugeValue(ctx context.Context, key string) (float64, bool, error) {
	query := "SELECT value FROM content.metrics WHERE name = $1 AND type = 'gauge'"
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
	query := "UPDATE content.metrics SET value = $2 WHERE name = $1 AND type = 'counter'"
	log.Println(query, key, value)
	_, err := pg.db.ExecContext(ctx, query, key, value)
	return err
}

func (pg *PGStorage) GetCounterValue(ctx context.Context, key string) (int64, bool, error) {
	query := "SELECT value FROM content.metrics WHERE name = $1 AND type = 'gauge'"
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
	query := "SELECT value FROM content.metrics WHERE type = 'gauge'"
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
	query := "SELECT value FROM content.metrics WHERE type = 'counter'"
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
