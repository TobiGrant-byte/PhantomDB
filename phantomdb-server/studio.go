package main

import (
	"embed"
	"net/http"
)

//go:embed studio/*
var studioFiles embed.FS

func registerStudio(mux *http.ServeMux, files embed.FS) {
	mux.Handle("/studio/", http.FileServer(http.FS(files)))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := files.ReadFile("studio/index.html")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(data)
	})
}