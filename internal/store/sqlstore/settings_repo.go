package sqlstore

import (
	"context"
	"database/sql"
)

type settingsRepo struct {
	db     *sql.DB
	driver string
}

func (r *settingsRepo) Get(ctx context.Context, key string) (string, error) {
	var v string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key=?`, key).Scan(&v)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return v, err
}
func (r *settingsRepo) Set(ctx context.Context, key, value string) error {
	var err error
	if r.driver == "mysql" {
		_, err = r.db.ExecContext(ctx,
			`INSERT INTO settings (`+"`key`"+`,value) VALUES (?,?) ON DUPLICATE KEY UPDATE value=VALUES(value)`, key, value)
	} else {
		_, err = r.db.ExecContext(ctx,
			`INSERT INTO settings (key,value) VALUES (?,?) ON CONFLICT(key) DO UPDATE SET value=?`, key, value, value)
	}
	return err
}
