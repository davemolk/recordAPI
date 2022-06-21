package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/davemolk/recordAPI/internal/validator"
	"github.com/lib/pq"
)

type Album struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Artist    string    `json:"artist"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

type AlbumModel struct {
	DB *sql.DB
}

func ValidateAlbum(v *validator.Validator, album *Album) {
	v.Check(album.Title != "", "title", "title required")
	v.Check(album.Artist != "", "artist", "artist required")
	v.Check(album.Genres != nil, "genre", "genre required")
	v.Check(validator.Unique(album.Genres), "genres", "no duplicate values")

}

func (a AlbumModel) Insert(album *Album) error {
	query := `
		INSERT INTO albums (title, artist, genres)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, version`

	args := []interface{}{album.Title, album.Artist, pq.Array(album.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return a.DB.QueryRowContext(ctx, query, args...).Scan(&album.ID, &album.CreatedAt, &album.Version)
}

func (a AlbumModel) Get(id int64) (*Album, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, artist, genres, version
		FROM albums
		WHERE id = $1`

	var album Album

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := a.DB.QueryRowContext(ctx, query, id).Scan(
		&album.ID,
		&album.CreatedAt,
		&album.Title,
		&album.Artist,
		pq.Array(&album.Genres),
		&album.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &album, nil

}

func (a AlbumModel) GetAll(title, artist string, genres []string, filters Filters) ([]*Album, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, title, artist, genres, version
		FROM albums
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '') 
		AND (to_tsvector('simple', artist) @@ plainto_tsquery('simple', $2) OR $2 = '') 
		AND (genres @> $3 OR $3 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, artist, pq.Array(genres), filters.limit(), filters.offset()}

	rows, err := a.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	albums := []*Album{}

	for rows.Next() {
		var album Album
		err := rows.Scan(
			&totalRecords,
			&album.ID,
			&album.CreatedAt,
			&album.Title,
			&album.Artist,
			pq.Array(&album.Genres),
			&album.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		albums = append(albums, &album)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return albums, metadata, nil
}

func (a AlbumModel) Update(album *Album) error {
	query := `
		UPDATE albums
		SET title = $1, artist = $2, genres = $3, version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING version`

	args := []interface{}{
		album.Title,
		album.Artist,
		pq.Array(album.Genres),
		album.ID,
		album.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := a.DB.QueryRowContext(ctx, query, args...).Scan(&album.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (a AlbumModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM albums
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := a.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
