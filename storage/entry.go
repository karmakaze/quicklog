package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func CreateEntry(at time.Time, client_id, service_id, context string,
	trace_id, span_id string, ctx context.Context, tx *sql.Tx) error {
	const query = `INSERT INTO entry (at, client_id, service_id, context) ` +
		`VALUES ($1, $2, $3, $4);`

	if _, err := tx.ExecContext(ctx, query, at, client_id, service_id, context); err != nil {
		fmt.Errorf("failed to insert entry (%v, %s, %s, %s) : %v", at, client_id, service_id, context, err)
		return err
	}
	return nil
}

func ListEntries(filterName, filterValue string, entries *[]entry, ctx context.Context, tx *sql.Tx) error {
	qquery := `SELECT id, at FROM entry WHERE ` + filterName + ` = $1 ORDER BY id`
	rows, err := tx.QueryContext(ctx, query, filterValue)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		return err
	}
	defer func() { _ = rows.Close() }()

	var count int
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		id := ""
		var at time.Time
		if err := rows.Scan(&id, &at); err != nil {
			return fmt.Errorf("failed to scan result set: %s", err)
		}
		if id != "" {
			(*entries) = append(*entries, entry)
		}
	}
	return nil
}

func DeleteEntriesOlderThan(t time.Time, ctx context.Context, tx *sql.Tx) error {
	const query = `DELETE FROM entry WHERE at < $1;`
	if _, err := tx.ExecContext(ctx, query, t); err != nil {
		return fmt.Errorf("failed to delete entries older than %v: %s", t, err)
	}
	return nil
}
