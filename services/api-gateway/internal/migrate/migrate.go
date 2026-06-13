package migrate

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

// Run executes the embedded schema.sql against the database.
func Run(ctx context.Context, pool *pgxpool.Pool) error {
	statements := strings.Split(schemaSQL, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("execute: %w\nSQL: %s", err, stmt[:min(len(stmt), 100)])
		}
	}

	log.Println("schema applied successfully")
	return nil
}
