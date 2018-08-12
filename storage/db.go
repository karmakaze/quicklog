package storage

import (
	"database/sql"

	"github.com/karmakaze/quicklog/config"
	_ "github.com/lib/pq"
)

// OpenDB opens the database connection according to the context.
func OpenDB(cfg config.Config) (*sql.DB, error) {
	return sql.Open("postgres", cfg.DbUrl)
}
