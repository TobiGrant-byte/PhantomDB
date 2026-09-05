package handlers

import (
	"log"
	"net/http"
)

func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		notes, _ := h.Store.ListNotes()
		if err := h.Tmpl.ExecuteTemplate(w, "note_list_items.html", notes); err != nil {
			log.Println("template error (empty search):", err)
		}
		return
	}
	notes, err := h.Store.SearchNotes(q)
	if err != nil {
		log.Println("search error:", err)
		http.Error(w, err.Error(), 500)
		return
	}
	if err := h.Tmpl.ExecuteTemplate(w, "note_list_items.html", notes); err != nil {
		log.Println("template error (search):", err)
	}
}