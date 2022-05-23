package main

import (
	"fmt"
	"net/http"
	"time"

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
	
	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) showAlbumHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	album := data.Album{
		ID: id,
		CreatedAt: time.Now(),
		Title: "Kind of Blue",
		Artist: "Miles Davis",
		Genres: []string{"jazz", "modal"},
		Version: 1,		
	}
	
	err = app.writeJSON(w, http.StatusOK, envelope{"album": album}, nil)
	if err != nil {
		app.logger.Println(err)
		app.serverErrorResponse(w, r, err)
	}

}
