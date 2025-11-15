package repo

import (
	"database/sql"
)

type SQLRepo struct {
	db *sql.DB
}

func New(db *sql.DB) *SQLRepo {
	return &SQLRepo{db: db}
}
