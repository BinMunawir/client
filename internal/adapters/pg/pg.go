package pg

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/BinMunawir/maal_business/config"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
)

// One shared connection pool per process — never open and close a connection per call
// (service standard §6.3). *sql.DB is itself a pool; DB() lazily opens it once and every
// store reuses it via `pg.DB()` inline in QueryContext.
var (
	once sync.Once
	db   *sql.DB
)

// DB returns the process-wide connection pool, opening and health-checking it on first
// use. A failure to reach the database is fatal — the service cannot function without it.
func DB() *sql.DB {
	once.Do(func() {
		conn, err := open()
		if err != nil {
			log.Fatalf("pg: %v", err)
		}
		db = conn
	})
	return db
}

func open() (*sql.DB, error) {
	dsn := dsn()
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if err := conn.PingContext(context.Background()); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return conn, nil
}

// dsn prefers an explicit DB_DSN (used by the justfile migration/codegen recipes so the
// app and the tooling agree), otherwise builds one from the individual Db.* config fields.
func dsn() string {
	if config.CNF.Db.DSN != "" {
		return config.CNF.Db.DSN
	}
	u := url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", config.CNF.Db.Host, config.CNF.Db.Port),
		Path:   "/" + config.CNF.Db.Name,
	}
	u.User = url.UserPassword(config.CNF.Db.User, config.CNF.Db.Password)
	q := url.Values{}
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return u.String()
}
