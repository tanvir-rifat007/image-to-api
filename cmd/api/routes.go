package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux:= http.NewServeMux()

	mux.HandleFunc("POST /v1/images", app.uploadImageHandler)

	return mux
}