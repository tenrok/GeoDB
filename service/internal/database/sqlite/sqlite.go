package sqlite

import (
	"database/sql"
	"errors"
	"net/url"

	"geodbsvc/internal/utils"
)

type SqliteDB struct {
	*sql.DB
}

// Open подключается к существующей БД
func Open(dsn string) (*SqliteDB, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	fpath := u.Host + u.Path
	if !utils.IsFileExists(fpath) {
		return nil, errors.New("database doesn't exist")
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
