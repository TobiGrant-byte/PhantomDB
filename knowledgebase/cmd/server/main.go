package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"knowledgebase/internal/handlers"
	"knowledgebase/internal/store"
)

func main() {
	s, err := store.Open("knowledgebase")
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	h := handlers.New(s)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", h.Index)
	r.Post("/notes", h.CreateNote)
	r.Get("/notes/{id}", h.ViewNote)
	r.Get("/notes/{id}/edit", h.EditNoteForm)
	r.Post("/notes/{id}", h.UpdateNote)
	r.Delete("/notes/{id}", h.DeleteNote)
	r.Get("/search", h.Search)
	r.Get("/tags/{name}", h.NotesByTag)

	log.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", r)
}