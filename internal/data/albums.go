package data

import (
	"time"

	"github.com/davemolk/recordAPI/internal/validator"
)

type Album struct {
	ID int64 `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title string `json:"title"`
	Artist string `json:"artist"`
	Genres []string `json:"genres,omitempty"`
	Version int32 `json:"version"`
}

func ValidateAlbum(v *validator.Validator, album *Album) {
	v.Check(album.Title != "", "title", "title required")
	v.Check(album.Artist != "", "artist", "artist required")
	v.Check(album.Genres != nil, "genre", "genre required")
	v.Check(validator.Unique(album.Genres), "genres", "no duplicate values")
	
} 