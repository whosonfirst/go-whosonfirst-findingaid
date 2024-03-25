package sql

import (
	"context"
	"fmt"
	"net/url"

	gosql "database/sql"
)

func CreateDB(ctx context.Context, uri string) (*gosql.DB, string, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, "", fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	dsn := q.Get("dsn")

	// scheme is assumed to be sql://
	engine := u.Host

	db, err := gosql.Open(engine, dsn)

	if err != nil {
		return nil, "", fmt.Errorf("Failed to open database, %v", err)
	}

	err = db.Ping()

	if err != nil {
		return nil, "", fmt.Errorf("Failed to ping database, %v", err)
	}

	if engine == "sqlite3" {

		pragma := []string{
			"PRAGMA JOURNAL_MODE=OFF",
			"PRAGMA SYNCHRONOUS=OFF",
			"PRAGMA LOCKING_MODE=EXCLUSIVE",
			// https://www.gaia-gis.it/gaia-sins/spatialite-cookbook/html/system.html
			"PRAGMA PAGE_SIZE=4096",
			"PRAGMA CACHE_SIZE=1000000",
		}

		for _, p := range pragma {

			_, err = db.ExecContext(ctx, p)

			if err != nil {
				return nil, "", fmt.Errorf("Failed to set '%s', %w", p, err)
			}
		}
	}

	// END OF put me in a package or something

	// START OF put me in a package or something

	// CHECK IF TABLES EXIST ALREADY...

	tables_sql := []string{
		"CREATE TABLE IF NOT EXISTS sources (id INTEGER, name TEXT PRIMARY KEY)",
		"CREATE TABLE IF NOT EXISTS catalog (id INTEGER PRIMARY KEY, repo_id INTEGER)",
	}

	for _, q := range tables_sql {

		_, err = db.ExecContext(ctx, q)

		if err != nil {
			return nil, "", fmt.Errorf("Failed to execute tables SQL '%s', %v", q, err)
		}
	}

	return db, engine, nil
}

func AddToCatalog(ctx context.Context, db *gosql.DB, id int64, repo_id int64) error {

	q := "INSERT INTO catalog (id, repo_id) VALUES(?, ?) ON CONFLICT(id) DO UPDATE SET repo_id=? WHERE id=?"

	_, err := db.ExecContext(ctx, q, id, repo_id, repo_id, id)

	if err != nil {
		return fmt.Errorf("Failed to store %d, %w", id, err)
	}

	return nil
}

func AddToSources(ctx context.Context, db *gosql.DB, repo_name string, repo_id int64) error {

	q := "INSERT INTO sources (id, name) VALUES(?, ?) ON CONFLICT(name) DO UPDATE SET id=? WHERE name=?"
	_, err := db.ExecContext(ctx, q, repo_id, repo_name, repo_id, repo_name)

	if err != nil {
		return fmt.Errorf("Failed to store %s, %w", repo_name, err)
	}

	return nil
}
