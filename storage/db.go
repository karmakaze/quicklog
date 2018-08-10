package storage

import (
	"database/sql"
	"errors"

	"github.com/karmakaze/quicklog/config"
	_ "github.com/lib/pq"
)

var errNoUser = errors.New("no user found")
var errNoPhoto = errors.New("no photos found")

// OpenDB opens the database connection according to the context.
func OpenDB(cfg config.Config) (*sql.DB, error) {
	return sql.Open("postgres", cfg.DBUrl)
}
