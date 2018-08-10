package span_tag

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type SpanTag struct {
	ProjectId int32  `json:"project_id"`
	TraceId   string `json:"trace_id"`
	SpanId    string `json:"span_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

func CreateSpanTag(t SpanTag, tx *sql.Tx, ctx context.Context) error {
	query := "INSERT INTO span_tag" +
		" (project_id, trace_id, span_id, `key`, value)" +
		" VALUES (?, ?, ?, ?, ?);"
	if _, err := tx.ExecContext(ctx, query, t.ProjectId, t.TraceId, t.SpanId, t.Key, t.Value); err != nil {
		return err
	}
	return nil
}

func ListSpanTags(projectId int, traceId, spanId, tag string, tx *sql.DB, ctx context.Context) ([]SpanTag, error) {
	var rows *sql.Rows

	var err error
	fields := "project_id, trace_id, span_id, `key`, value"

	traceOrSpanId := ""
	traceOrSpanCols := ""
	if traceId != "" {
		traceOrSpanId = traceId
		if spanId == traceId {
			traceOrSpanCols = "(trace_id, span_id)"
		} else {
			traceOrSpanCols = "(trace_id)"
		}
	} else if spanId != "" {
		traceOrSpanId = spanId
		traceOrSpanCols = "(span_id)"
	}

	if tag != "" {
		key, value := ParseTag(tag)
		if key != "" {
			query := "SELECT " + fields + " FROM span_tag WHERE project_id = ? AND `key` = ? AND value = ?"
			rows, err = tx.QueryContext(ctx, query, projectId, key, value)
		} else {
			query := "SELECT " + fields + " FROM span_tag WHERE project_id = ? AND value = ?"
			rows, err = tx.QueryContext(ctx, query, projectId, value)
		}
	} else {
		query := "SELECT " + fields + " FROM span_tag WHERE project_id = ?" +
			" AND ? IN " + traceOrSpanCols
		rows, err = tx.QueryContext(ctx, query, projectId, traceOrSpanId)
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

	spanTags := make([]SpanTag, 0)

	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, err
		}
		var t SpanTag
		if err = rows.Scan(&t.ProjectId, &t.TraceId, &t.SpanId, &t.Key, &t.Value); err != nil {
			return nil, fmt.Errorf("failed to scan result set: %s", err)
		}
		spanTags = append(spanTags, t)
	}

	return spanTags, nil
}

func ParseTag(tag string) (key, value string) {
	i := strings.Index(tag, ":")
	if i == -1 {
		return "", tag
	}
	key = tag[0:i]
	value = tag[i+1:]
	return
}

func TraceIds(spanTags []SpanTag) []string {
	traceIdSet := make(map[string]struct{}, len(spanTags))
	for _, spanTag := range spanTags {
		if spanTag.TraceId != "" {
			traceIdSet[spanTag.TraceId] = struct{}{}
		}
	}
	traceIds := make([]string, 0, len(traceIdSet))
	for traceId, _ := range traceIdSet {
		traceIds = append(traceIds, traceId)
	}
	return traceIds
}

func SpanIds(spanTags []SpanTag) []string {
	spanIdSet := make(map[string]struct{}, len(spanTags))
	for _, spanTag := range spanTags {
		if spanTag.SpanId != "" {
			spanIdSet[spanTag.SpanId] = struct{}{}
		}
	}

	spanIds := make([]string, 0, len(spanIdSet))
	for spanId, _ := range spanIdSet {
		spanIds = append(spanIds, spanId)
	}
	return spanIds
}
