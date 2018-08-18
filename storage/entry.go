package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Entry struct {
	ProjectId int32      `json:"project_id"`
	Seq       int64      `json:"seq"`
	Published time.Time  `json:"published"`
	Source    string     `json:"source"`
	Type      string     `json:"type"`
	Actor     string     `json:"actor"`
	Object    string     `json:"object"`
	Target    string     `json:"target"`
	Context   ContextMap `json:"context"`
	TraceId   string     `json:"trace_id"`
	SpanId    string     `json:"span_id"`
}

type ContextMap map[string]interface{}

func (c ContextMap) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	j, err := json.Marshal(c)
	return j, err
}

func (c *ContextMap) Scan(src interface{}) error {
	if src == nil {
		*c = nil
		return nil
	}
	s, ok := src.([]byte)
	if !ok {
		return errors.New("ContextMap.Scan: type assertion .([]byte) failed.")
	}

	var o interface{}
	err := json.Unmarshal(s, &o)
	if err != nil {
		return err
	}

	*c, ok = o.(map[string]interface{})
	if !ok {
		return errors.New("ContextMap.Scan: type assertion .(map[string]interface{}) failed.")
	}

	return nil
}

func CreateEntry(e Entry, tx *sql.Tx, ctx context.Context) error {
	query := `INSERT INTO entry` +
		` (project_id, published, source, type, actor, object, target, context, trace_id, span_id)` +
		` VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`
	if _, err := tx.ExecContext(ctx, query, e.ProjectId, e.Published, e.Source,
		e.Type, e.Actor, e.Object, e.Target, e.Context, e.TraceId, e.SpanId); err != nil {
		return fmt.Errorf("failed to insert %v: %v", e, err)
	}
	return nil
}

func ListEntries(projectId int, publishedMin, publishedMax time.Time, limit int, tx *sql.Tx, ctx context.Context) ([]Entry, error) {
	var rows *sql.Rows
	reverse := false

	var err error
	fields := `project_id, seq, published, source, type, actor, object, target, context, trace_id, span_id`

	if !publishedMin.IsZero() && !publishedMax.IsZero() {
		query := `SELECT ` + fields + ` FROM entry WHERE project_id = $1` +
			` AND published BETWEEN $2 AND $3 ORDER BY seq LIMIT $4`
		rows, err = tx.QueryContext(ctx, query, projectId, publishedMin, publishedMax, limit)
	} else if !publishedMin.IsZero() {
		query := `SELECT ` + fields + ` FROM entry WHERE project_id = $1` +
			` AND published >= $2 ORDER BY seq LIMIT $3`
		rows, err = tx.QueryContext(ctx, query, projectId, publishedMin, limit)
	} else if !publishedMax.IsZero() {
		query := `SELECT ` + fields + ` FROM entry WHERE project_id = $1` +
			` AND published <= $2 ORDER BY seq DESC LIMIT $3`
		reverse = true
		rows, err = tx.QueryContext(ctx, query, projectId, publishedMax, limit)
	} else {
		query := `SELECT ` + fields + ` FROM entry WHERE project_id = $1 ORDER BY seq LIMIT $2`
		rows, err = tx.QueryContext(ctx, query, projectId, limit)
	}

	if rows != nil {
		defer rows.Close()
	}

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	}

	entries := make([]Entry, 0)

	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, err
		}
		var e Entry
		if err = rows.Scan(&e.ProjectId, &e.Seq, &e.Published, &e.Source, &e.Type, &e.Actor, &e.Object, &e.Target,
			&e.Context, &e.TraceId, &e.SpanId); err != nil {
			return nil, fmt.Errorf("failed to scan result set: %s", err)
		}
		e.Published = e.Published.UTC()

		entries = append(entries, e)
	}

	if reverse {
		reverseEntries(entries)
	}

	return entries, nil
}

func reverseEntries(entries []Entry) {
	lst := len(entries) - 1
	mid := len(entries) / 2
	for i := 0; i < mid; i++ {
		entries[i], entries[lst-i] = entries[lst-i], entries[i]
	}
}

func DeleteEntries(projectId int, publishedMin, publishedMax time.Time, tx *sql.Tx, ctx context.Context) error {
	var err error
	if !publishedMin.IsZero() && !publishedMax.IsZero() {
		query := `DELETE FROM entry WHERE project_id = $1 AND published BETWEEN $2 AND $3;`
		_, err = tx.ExecContext(ctx, query, projectId, publishedMin, publishedMax)
	} else if !publishedMin.IsZero() {
		query := `DELETE FROM entry WHERE project_id = $1 AND $2 <= published;`
		_, err = tx.ExecContext(ctx, query, projectId, publishedMin)
	} else if !publishedMax.IsZero() {
		query := `DELETE FROM entry WHERE project_id = $1 AND published <= $2;`
		_, err = tx.ExecContext(ctx, query, projectId, publishedMax)
	}
	return err
}
