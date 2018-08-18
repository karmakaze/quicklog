package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type Project struct {
	Id   int32  `json:"id"`
	Name string `json:"name"`
}

func CreateProject(p Project, tx *sql.Tx, ctx context.Context) error {
	query := `INSERT INTO project (name) VALUES ($1)`
	if _, err := tx.ExecContext(ctx, query, p.Name); err != nil {
		return fmt.Errorf("failed to insert %v: %v", p, err)
	}
	return nil
}

func ListProjects(filterName, filterValue string, projects *[]Project, tx *sql.Tx, ctx context.Context) error {
	var rows *sql.Rows
	var err error
	fields := `id, name`
	if filterName != "" {
		query := `SELECT ` + fields + ` FROM project WHERE ` + filterName + ` = $1 ORDER BY name, id`
		rows, err = tx.QueryContext(ctx, query, filterValue)
	} else {
		query := `SELECT ` + fields + ` FROM project ORDER BY name, id`
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
		var p Project
		if err = rows.Scan(&p.Id, &p.Name); err != nil {
			return fmt.Errorf("failed to scan result set: %s", err)
		}

		(*projects) = append(*projects, p)
	}
	return nil
}
