package handlers

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"knowledgebase/internal/store"
)

type Handlers struct {
	Store *store.Store
	Tmpl  *template.Template
}

func New(s *store.Store) *Handlers {
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	return &Handlers{Store: s, Tmpl: tmpl}
}

func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	notes, err := h.Store.ListNotes()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.Tmpl.ExecuteTemplate(w, "index.html", notes)
}

func (h *Handlers) CreateNote(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	content := r.FormValue("content")

	note, err := h.Store.CreateNote(title, content)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/notes/"+note.ID, http.StatusSeeOther)
}

func (h *Handlers) ViewNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	note, err := h.Store.GetNote(id)
	if err != nil {
		http.Error(w, "note not found", 404)
		return
	}
	h.Tmpl.ExecuteTemplate(w, "note.html", note)
}

func (h *Handlers) DeleteNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Store.DeleteNote(id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}


func (h *Handlers) EditNoteForm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	note, err := h.Store.GetNote(id)
	if err != nil {
		http.Error(w, "note not found", 404)
		return
	}
	h.Tmpl.ExecuteTemplate(w, "edit.html", note)
}

func (h *Handlers) UpdateNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	r.ParseForm()
	title := r.FormValue("title")
	content := r.FormValue("content")

	_, err := h.Store.UpdateNote(id, title, content)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/notes/"+id, http.StatusSeeOther)
}