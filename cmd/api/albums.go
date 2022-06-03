package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/davemolk/recordAPI/internal/data"
	"github.com/davemolk/recordAPI/internal/validator"
)

func (app *application) createAlbumHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		Artist string `json:"artist"`
		Genres []string `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	fmt.Println(input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	album := &data.Album{
		Title: input.Title,
		Artist: input.Artist,
		Genres: input.Genres,
	}

	v := validator.New()

	if data.ValidateAlbum(v, album); !v.Valid() {
		app.failedValidationsResponse(w, r, v.Errors)
		return
	}

	err = app.models.Albums.Insert(album)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/albums/%d", album.ID))
	
	err = app.writeJSON(w, http.StatusCreated, envelope{"album": album}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showAlbumHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	album, err := app.models.Albums.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	
	err = app.writeJSON(w, http.StatusOK, envelope{"album": album}, nil)
	if err != nil {
		app.logger.Println(err)
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateAlbumHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	album, err := app.models.Albums.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	
	var input struct {
		Title *string `json:"title"`
		Artist *string `json:"artist"`
		Genres []string `json:"genres"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		album.Title = *input.Title
	}

	if input.Artist != nil {
		album.Artist = *input.Artist
	}

	if input.Genres != nil {
		album.Genres = input.Genres
	}

	v := validator.New()

	if data.ValidateAlbum(v, album); !v.Valid() {
		app.failedValidationsResponse(w, r, v.Errors)
		return
	}

	err = app.models.Albums.Update(album)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r,)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"album": album}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteAlbumHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Albums.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "album deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}