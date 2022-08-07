package database

import (
	"nats-server/internal/config"
	"net/url"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Register postgres database
)

func Open(cfg config.DbConfig) (*sqlx.DB, error) {
	q := url.Values{}

	q.Set("sslmode", "disable")
	q.Set("timezone", "utc")
	q.Set("port", cfg.Port)

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Path,
		RawQuery: q.Encode(),
	}

	return sqlx.Open("postgres", u.String())
}
