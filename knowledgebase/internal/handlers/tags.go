package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) NotesByTag(w http.ResponseWriter, r *http.Request) {
	tagName := chi.URLParam(r, "name")
	noteIDs, err := h.Store.ListNotesByTag(tagName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var notes []interface{}
	for _, id := range noteIDs {
		n, err := h.Store.GetNote(id)
		if err == nil {
			notes = append(notes, n)
		}
	}
	h.Tmpl.ExecuteTemplate(w, "search_results.html", notes)
}