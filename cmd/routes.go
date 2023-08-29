package main

import (
	"github.com/go-chi/chi/v5"
)

func (app *application) setRouters() {

	app.router.Route("/", func(r chi.Router) {
		app.router.Get("/history/", app.getHistory)

		app.router.Route("/segments", func(router chi.Router) {

			router.Post("/{slug}", app.createSegment)
			router.Delete("/{slug}", app.deleteSegment)
		})

		app.router.Route("/users-segments", func(router chi.Router) {

			router.Get("/{user_id}", app.getSegments)
			router.Put("/{user_id}", app.updateSegments)
		})

	})

	app.router.NotFound(errorNotFound)
}
