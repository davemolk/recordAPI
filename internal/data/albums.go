package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/davemolk/recordAPI/internal/validator"
	"github.com/lib/pq"
)

type Album struct {
	ID int64 `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title string `json:"title"`
	Artist string `json:"artist"`
	Genres []string `json:"genres,omitempty"`
	Version int32 `json:"version"`
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

	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
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

func (a AlbumModel) GetAll(title, artist string, genres []string, filters Filters) ([]*Album, error) {
	query := `
		SELECT id, created_at, title, artist, genres, version
		FROM albums
		ORDER BY id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	albums := []*Album{}
	for rows.Next() {
		var album Album
		err := rows.Scan(
			&album.ID,
			&album.CreatedAt,
			&album.Title,
			&album.Artist,
			pq.Array(&album.Genres),
			&album.Version,
		)
		if err != nil {
			return nil, err
		}
		albums = append(albums, &album)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return albums, nil
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


