package sqlstore

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func NewMySQL(dsn string) (*SQLStore, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql: open: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("mysql: ping: %w", err)
	}

	s := &SQLStore{db: db, driver: "mysql"}
	if err := s.migrateMySQL(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

func (s *SQLStore) migrateMySQL() error {
	_, err := s.db.Exec(schemaMySQL)
	if err != nil {
		return fmt.Errorf("mysql: migrate: %w", err)
	}
	return nil
}
