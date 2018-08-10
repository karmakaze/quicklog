package storage

import (
    "context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	_ "github.com/lib/pq"
)

const (
	BitsPerWord = 32 << (^uint(0) >> 63)
	MaxInt      = 1<<(BitsPerWord-1) - 1
	MinInt      = -MaxInt - 1
)

// OpenDB opens the database connection according to the context.
func OpenDB(dbUrl string) (*sql.DB, error) {
	return sql.Open("postgres", dbUrl)
}

func IsUniqueViolation(err error) bool {
	return false
	// pqErr, ok := err.(*pq.Error)
	// return ok && pqErr.Code == "23505" // unique_violation
}

func Ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func FirstNonEmpty(values ...string) string {
	for _, s := range values {
		if s != "" {
			return s
		}
	}
	return ""
}

func StringToNullable(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func execContext(tx *sql.DB, ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    return tx.ExecContext(ctx, numberArgs(query), args...)
}

func execTxContext(tx *sql.Tx, ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    return tx.ExecContext(ctx, numberArgs(query), args...)
}

func queryContext(tx *sql.DB, ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    fmt.Printf("SQL: %s\n", query)
    return tx.QueryContext(ctx, numberArgs(query), args...)
}

func queryTxContext(tx *sql.Tx, ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    fmt.Printf("SQL: %s\n", query)
    return tx.QueryContext(ctx, numberArgs(query), args...)
}

func numberArgs(query string) string {
    num := 1
    for {
        if i := strings.Index(query, "?"); i == -1 {
            return query
        } else {
            query = query[0:i] + "$" + strconv.Itoa(num) + query[i+1:]
            num += 1
        }
    }
}
