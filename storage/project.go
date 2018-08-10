package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type Project struct {
	Id     int32  `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

func CreateProject(p Project, tx *sql.Tx, ctx context.Context) error {
	query := `INSERT INTO project (name, domain) VALUES (?, ?)`
	if _, err := tx.ExecContext(ctx, query, p.Name, StringToNullable(p.Domain)); err != nil {
		return err
	}
	return nil
}

func ListProjects(filterName, filterValue string, projects *[]Project, db *sql.DB, ctx context.Context) error {
	var rows *sql.Rows
	var err error
	fields := `id, name, domain`

	if filterName != "" {
		query := `SELECT ` + fields + ` FROM project WHERE ` + filterName + ` = ? ORDER BY name, id`
		rows, err = db.QueryContext(ctx, query, filterValue)
	} else {
		query := `SELECT ` + fields + ` FROM project ORDER BY name, id`
		rows, err = db.QueryContext(ctx, query)
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
		var domain sql.NullString
		if err = rows.Scan(&p.Id, &p.Name, &domain); err != nil {
			return fmt.Errorf("failed to scan result set: %s", err)
		}
		if domain.Valid {
		    p.Domain = domain.String
		}

		(*projects) = append(*projects, p)
	}
	return nil
}
