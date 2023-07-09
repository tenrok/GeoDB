package sqlite

import (
	"database/sql"
	"net/url"
)

type SqliteDB struct {
	*sql.DB
}

// New подключается к существующей БД или создаёт новую
func New(dsn string) (*SqliteDB, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	queries := u.Query()
	queries.Set("_fk", "1")
	queries.Set("mode", "ro")
	u.RawQuery = queries.Encode()

	db, err := sql.Open("sqlite3", u.String())
	if err != nil {
		return nil, err
	}
	d := &SqliteDB{db}
	return d, nil
}
