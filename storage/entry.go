package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Entry struct {
	Id        int64                  `json:"id"`
	Published time.Time              `json:"published"`
	Source    string                 `json:"source"`
	Type      string                 `json:"type"`
	Actor     string                 `json:"actor"`
	Object    string                 `json:"object"`
	Target    string                 `json:"target"`
	Context   map[string]interface{} `json:"context"`
	TraceId   string                 `json:"trace_id"`
	SpanId    string                 `json:"span_id"`
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
	fields := `id, published, source, type, actor, object, target, trace_id, span_id`
	if filterName != "" {
		query := `SELECT ` + fields + ` FROM entry WHERE ` + filterName + ` = $1 ORDER BY id`
		rows, err = tx.QueryContext(ctx, query, filterValue)
	} else {
		query := `SELECT ` + fields + ` FROM entry ORDER BY id`
		rows, err = tx.QueryContext(ctx, query)
	}
	if rows != nil {
		defer rows.Close()
	}

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		return err
	}

	names, _ := rows.Columns()

	for rows.Next() {
		if err = rows.Err(); err != nil {
			return err
		}
		var e Entry
		if err = rows.Scan(&e.Id, &e.Published, &e.Source, &e.Type, &e.Actor, &e.Object, &e.Target, &e.TraceId, &e.SpanId); err != nil {
			return fmt.Errorf("failed to scan result set: %s", err)
		}

		(*entries) = append(*entries, e)
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
