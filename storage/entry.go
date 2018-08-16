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

func ListEntries(filterName, filterValue string, entries *[]Entry, tx *sql.Tx, ctx context.Context) error {
	var rows *sql.Rows
	var err error
	fields := `project_id, seq, published, source, type, actor, object, target, context, trace_id, span_id`
	if filterName != "" {
		query := `SELECT ` + fields + ` FROM entry WHERE ` + filterName + ` = $1 ORDER BY project_id, seq`
		rows, err = tx.QueryContext(ctx, query, filterValue)
	} else {
		query := `SELECT ` + fields + ` FROM entry ORDER BY project_id, seq`
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

	for rows.Next() {
		if err = rows.Err(); err != nil {
			return err
		}
		var e Entry
		if err = rows.Scan(&e.ProjectId, &e.Seq, &e.Published, &e.Source, &e.Type, &e.Actor, &e.Object, &e.Target,
			&e.Context, &e.TraceId, &e.SpanId); err != nil {
			return fmt.Errorf("failed to scan result set: %s", err)
		}
		e.Published = e.Published.UTC()

		(*entries) = append(*entries, e)
	}
	return nil
}

func DeleteEntriesOlderThan(projectId int, publishedBefore time.Time, tx *sql.Tx, ctx context.Context) error {
	query := `DELETE FROM entry WHERE project_id = $1 AND published < $2;`
	if _, err := tx.ExecContext(ctx, query, projectId, publishedBefore); err != nil {
		return fmt.Errorf("failed to delete project_id %d entries older than %v: %s", projectId, publishedBefore, err)
	}
	return nil
}
