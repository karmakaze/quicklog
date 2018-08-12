package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Entry struct {
	Id        int64
	Published time.Time
	Source    string
	Type      string
	Actor     string
	Object    string
	Target    string
	Context   map[string]interface{}
	TraceId   string
	SpanId    string
}

func CreateEntry(e Entry, tx *sql.Tx, ctx context.Context) error {
	query := `INSERT INTO entry (published, source, type, actor, object, target, trace_id, span_id)` +
		` VALUES (NOW(), $1, $2, $3, $4, $5, $6, $7);`
	if _, err := tx.ExecContext(ctx, query, e.Source, e.Target, e.Actor, e.Object, e.Target,
		e.TraceId, e.SpanId); err != nil {
		return fmt.Errorf("failed to insert %v : %v", e, err)
	}
	return nil
}

func ListEntries(filterName, filterValue string, entries *[]Entry, tx *sql.Tx, ctx context.Context) error {
	var rows *sql.Rows
	var err error
	if filterName != "" {
		query := `SELECT id, published FROM entry WHERE ` + filterName + ` = $1 ORDER BY id`
		fmt.Printf("executing query %s\n", query)
		rows, err = tx.QueryContext(ctx, query, filterValue)
	} else {
		query := `SELECT id, published FROM entry ORDER BY id`
		fmt.Printf("executing query %s\n", query)
		rows, err = tx.QueryContext(ctx, query)
	}
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		fmt.Printf("query error: %v\n", err)
	} else {
		names, _ := rows.Columns()
		fmt.Printf("columns: %v\n", names)
	}

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		return err
	}

	for rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		var entry Entry
		if err := rows.Scan(&entry.Id, &entry.Published); err != nil {
			return fmt.Errorf("failed to scan result set: %s", err)
		}

		(*entries) = append(*entries, entry)
	}
	return nil
}

func DeleteEntriesOlderThan(published time.Time, tx *sql.Tx, ctx context.Context) error {
	query := `DELETE FROM entry WHERE published < $1;`
	if _, err := tx.ExecContext(ctx, query, published); err != nil {
		return fmt.Errorf("failed to delete entries older than %v: %s", published, err)
	}
	return nil
}
