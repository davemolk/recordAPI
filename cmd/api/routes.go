package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)


func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/albums", app.listAlbumsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/albums", app.createAlbumHandler)
	router.HandlerFunc(http.MethodGet, "/v1/albums/:id", app.showAlbumHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/albums/:id", app.updateAlbumHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/albums/:id", app.deleteAlbumHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	return app.recoverPanic(app.rateLimit(router))
}