package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Albums AlbumModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Albums: AlbumModel{DB: db},
	}
}