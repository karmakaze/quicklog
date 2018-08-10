package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/karmakaze/quicklog/storage/span_tag"
)

type Entry struct {
	ProjectId    int32      `json:"project_id"`
	Seq          int64      `json:"seq"`
	Published    time.Time  `json:"published"`
	Source       string     `json:"source"`
	Type         string     `json:"type"`
	Actor        string     `json:"actor"`
	Object       string     `json:"object"`
	Target       string     `json:"target"`
	Context      ContextMap `json:"context"`
	Repeated     int32      `json:"repeated"`
	TraceId      string     `json:"trace_id"`
	ParentSpanId string     `json:"parent_span_id"`
	SpanId       string     `json:"span_id"`
}

// true if ProjectId, Source, Type, Actor, Object, Target, Context all match
func (e Entry) matches(entry Entry) bool {
	return e.ProjectId == entry.ProjectId && e.Source == entry.Source && e.Type == entry.Type &&
		e.Actor == entry.Actor && e.Object == entry.Object && e.Target == entry.Target &&
		reflect.DeepEqual(e.Context, entry.Context)
}

var (
	entryCols = "project_id, seq, published, source, type, actor, object, target, context, repeated, trace_id, parent_span_id, span_id"
)

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
		return fmt.Errorf("ContextMap.Scan: type assertion .([]byte) failed.")
	}

	var o interface{}
	err := json.Unmarshal(s, &o)
	if err != nil {
		return err
	}

	*c, ok = o.(map[string]interface{})
	if !ok {
		return fmt.Errorf("ContextMap.Scan: type assertion .(map[string]interface{}) failed.")
	}

	return nil
}

func CreateEntry(e Entry, tx *sql.Tx, ctx context.Context) error {
	if lasts, err := selectLastEntries(e.ProjectId, 2, tx, ctx); err == nil && len(lasts) == 2 {
		last1 := lasts[0]
		last2 := lasts[1]
		if e.matches(last2) && e.matches(last1) {
			query := "UPDATE entry SET published = ?, repeated = repeated + 1," +
				" trace_id = ?, parent_span_id = ?, span_id = ?" +
				" WHERE project_id = ? AND seq = ?"
			if _, err := execTxContext(tx, ctx, query, e.Published, StringToNullable(e.TraceId),
				StringToNullable(e.ParentSpanId), StringToNullable(e.SpanId),
				last1.ProjectId, last1.Seq); err != nil {
				return err
			}
			return nil
		}
	}

	query := `INSERT INTO entry` +
		` (project_id, published, source, type, actor, object, target, context, repeated, trace_id, parent_span_id, span_id)` +
		` VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	if _, err := execTxContext(tx, ctx, query, e.ProjectId, e.Published, e.Source,
		e.Type, e.Actor, e.Object, e.Target, e.Context, e.Repeated, StringToNullable(e.TraceId),
		StringToNullable(e.ParentSpanId), StringToNullable(e.SpanId)); err != nil {
		return err
	}
	return nil
}

func ListEntries(projectId int, seqMin, seqMax int, publishedMin, publishedMax time.Time,
	traceId, spanId, search string, limit int, tx *sql.DB, ctx context.Context) ([]Entry, error) {
	var rows *sql.Rows
	desc := true

	var err error
	fields := entryCols

	traceOrSpanId := ""
	traceOrSpanCols := ""
	if traceId != "" {
		traceOrSpanId = traceId
		if spanId == traceId {
			traceOrSpanCols = "(trace_id, parent_span_id, span_id)"
		} else {
			traceOrSpanCols = "(trace_id)"
		}
	} else if spanId != "" {
		traceOrSpanId = spanId
		traceOrSpanCols = "(parent_span_id, span_id)"
	}

	objectOrTarget := ""
	objectOrTargetCol := ""
	tag := ""
	if strings.HasPrefix(search, "object:") {
		objectOrTargetCol = "object"
		objectOrTarget = search[7:]
	} else if strings.HasPrefix(search, "target:") {
		objectOrTargetCol = "target"
		objectOrTarget = search[7:]
	} else if strings.HasPrefix(search, "tag:") {
		tag = search[4:]
	} else if search != "" {
		tag = search
	}

	if objectOrTarget != "" {
		args := make([]interface{}, 0, 6)
		args = append(args, projectId)
		query := "SELECT " + fields + " FROM entry WHERE project_id = ?"
		query += " AND " + objectOrTargetCol + " = ?"
		args = append(args, objectOrTarget)

		if seqMin != MinInt || seqMax != MaxInt {
			query += " AND seq BETWEEN ? AND ?"
			args = append(args, seqMin, seqMax)
		}
		if seqMin != MinInt {
			desc = false
			query += " ORDER BY seq LIMIT ?"
		} else {
			query += " ORDER BY seq DESC LIMIT ?"
		}
		args = append(args, limit)
		rows, err = queryContext(tx, ctx, query, args...)
	} else if tag != "" {
		spanTags, err := span_tag.ListSpanTags(projectId, "", "", tag, tx, ctx)
		if err != nil {
			return nil, err
		}
		if len(spanTags) == 0 {
			return make([]Entry, 0), nil
		}

		traceIds := span_tag.TraceIds(spanTags)
		spanIds := span_tag.SpanIds(spanTags)

		query := "SELECT " + fields + " FROM entry e WHERE e.project_id = ?"
		args := make([]interface{}, 0, 6)
		args = append(args, projectId)

		if seqMin != MinInt || seqMax != MaxInt {
			query += " AND e.seq BETWEEN ? AND ?"
			args = append(args, seqMin, seqMax)
		}

		quotedTraceIds := "'" + strings.Join(traceIds, "','") + "'"
		quotedSpanIds := "'" + strings.Join(spanIds, "','") + "'"
		if len(traceIds) != 0 {
			if len(spanIds) != 0 {
				query += " AND (trace_id IN (" + quotedTraceIds + ")"
				query += "      OR parent_span_id IS NOT NULL AND parent_span_id IN (" + quotedSpanIds + ")"
				query += "      OR span_id IN (" + quotedSpanIds + "))"
			} else {
				query += " AND trace_id IN (" + quotedTraceIds + ")"
			}
		} else if len(spanIds) != 0 {
			query += " AND (parent_span_id IS NOT NULL AND parent_span_id IN (" + quotedSpanIds + ") OR span_id IN (" + quotedSpanIds + "))"
		}

		if seqMin != MinInt {
			desc = false
			query += " ORDER BY e.seq LIMIT ?"
		} else {
			query += " ORDER BY e.seq DESC LIMIT ?"
		}
		args = append(args, limit)
		rows, err = queryContext(tx, ctx, query, args...)
	} else if seqMin != MinInt && seqMax != MaxInt {
		if traceOrSpanId == "" {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND seq BETWEEN ? AND ? ORDER BY seq DESC LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, seqMin, seqMax, limit)
		} else {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND seq BETWEEN ? AND ? AND ? IN ` + traceOrSpanCols + ` ORDER BY seq DESC LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, seqMin, seqMax, traceOrSpanId, limit)
		}
	} else if seqMin != MinInt {
		if traceOrSpanId == "" {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND seq >= ? ORDER BY seq LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, seqMin, limit)
		} else {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND seq >= ? AND ? IN ` + traceOrSpanCols + ` ORDER BY seq LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, seqMin, traceOrSpanId, limit)
		}
		desc = false
	} else if seqMax != MaxInt {
		if traceOrSpanId == "" {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND seq <= ? ORDER BY seq DESC LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, seqMax, limit)
		} else {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND seq <= ? AND ? IN ` + traceOrSpanCols + ` ORDER BY seq DESC LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, seqMax, traceOrSpanId, limit)
		}
	} else if !publishedMin.IsZero() && !publishedMax.IsZero() {
		if traceOrSpanId == "" {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND published BETWEEN ? AND ? ORDER BY seq DESC LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, publishedMin, publishedMax, limit)
		} else {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND published BETWEEN ? AND ? AND ? IN ` + traceOrSpanCols + ` ORDER BY seq DESC LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, publishedMin, publishedMax, traceOrSpanId, limit)
		}
	} else if !publishedMin.IsZero() {
		if traceOrSpanId == "" {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND published >= ? ORDER BY seq LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, publishedMin, limit)
		} else {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND published >= ? AND ? IN ` + traceOrSpanCols + ` ORDER BY seq LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, publishedMin, traceOrSpanId, limit)
		}
		desc = false
	} else if !publishedMax.IsZero() {
		if traceOrSpanId == "" {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND published <= ? ORDER BY seq DESC LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, publishedMax, limit)
		} else {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND published <= ? AND ? IN ` + traceOrSpanCols + ` ORDER BY seq DESC LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, publishedMax, traceOrSpanId, limit)
		}
	} else {
		if traceOrSpanId == "" {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ? ORDER BY seq DESC LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, limit)
		} else {
			query := `SELECT ` + fields + ` FROM entry WHERE project_id = ?` +
				` AND ? IN ` + traceOrSpanCols + ` ORDER BY seq DESC LIMIT ?`
			rows, err = queryContext(tx, ctx, query, projectId, traceOrSpanId, limit)
		}
	}

	if rows != nil {
		defer rows.Close()
	}

	switch {
	case err == sql.ErrNoRows:
		return make([]Entry, 0), nil
	case err != nil:
		return nil, err
	case rows == nil:
		return make([]Entry, 0), nil
	}

	entries, err := resultEntries(rows)
	if err != nil {
		return nil, err
	}

	if desc {
		reverseEntries(entries)
	}

	return entries, nil
}

func DeleteEntries(projectId int, publishedMin, publishedMax time.Time, tx *sql.Tx, ctx context.Context) error {
	var err error
	if !publishedMin.IsZero() && !publishedMax.IsZero() {
		query := `DELETE FROM entry WHERE project_id = ? AND published BETWEEN ? AND ?;`
		_, err = execTxContext(tx, ctx, query, projectId, publishedMin, publishedMax)
	} else if !publishedMin.IsZero() {
		query := `DELETE FROM entry WHERE project_id = ? AND ? <= published;`
		_, err = execTxContext(tx, ctx, query, projectId, publishedMin)
	} else if !publishedMax.IsZero() {
		query := `DELETE FROM entry WHERE project_id = ? AND published <= ?;`
		_, err = execTxContext(tx, ctx, query, projectId, publishedMax)
	}
	return err
}

func selectLastEntries(projectId int32, limit int, tx *sql.Tx, ctx context.Context) ([]Entry, error) {
	query := "SELECT " + entryCols + " FROM entry WHERE project_id = ?" +
		" ORDER BY seq DESC LIMIT ?"
	rows, err := queryTxContext(tx, ctx, query, projectId, limit)
	if rows != nil {
		defer rows.Close()
	}
	switch {
	case err == sql.ErrNoRows:
		return make([]Entry, 0), nil
	case err != nil:
		return nil, err
	case rows == nil:
		return make([]Entry, 0), nil
	}

	return resultEntries(rows)
}

func resultEntries(rows *sql.Rows) ([]Entry, error) {
	entries := make([]Entry, 0)
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		var e Entry
		var traceId, parentSpanId, spanId sql.NullString
		if err := rows.Scan(&e.ProjectId, &e.Seq, &e.Published, &e.Source, &e.Type, &e.Actor, &e.Object, &e.Target,
			&e.Context, &e.Repeated, &traceId, &parentSpanId, &spanId); err != nil {
			return nil, fmt.Errorf("failed to scan result set: %s", err)
		}
		e.Published = e.Published.UTC()
		e.TraceId = traceId.String
		e.ParentSpanId = parentSpanId.String
		e.SpanId = spanId.String

		entries = append(entries, e)
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
