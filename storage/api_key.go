package storage

import (
	"context"
	"database/sql"
	"fmt"
)

func VerifyApiKey(projectId int, apiKey string, db *sql.DB, ctx context.Context) bool {
	query := `SELECT COUNT(*) FROM api_key JOIN project ON project.id = api_key.project_id` +
		` WHERE project.id = $1 AND api_key.id = $2;`
	row := db.QueryRowContext(ctx, query, projectId, apiKey)
	count := 0
	if err := row.Scan(&count); err != nil {
		fmt.Printf("failed to scan row: %s\n", err)
		return false
	}
	fmt.Printf("row count %d\n", count)
	return count == 1
}
