package main

import (
	"encoding/json"
	"net/http"

	"github.com/TobiGrant-byte/phantomdb"
)

func registerAPI(mux *http.ServeMux, db *phantomdb.DB) {
	auth := requireAuth(db, RoleRead)
	authWrite := requireAuth(db, RoleWrite)
	adminOnly := requireAuth(db, RoleAdmin)

	mux.HandleFunc("/api/put", authWrite(func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Key, Value string }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := db.Put([]byte(body.Key), []byte(body.Value)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(200)
	}))

	mux.HandleFunc("/api/get", auth(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val, err := db.Get([]byte(key))
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"key": key, "value": string(val)})
	}))

	mux.HandleFunc("/api/scan", auth(func(w http.ResponseWriter, r *http.Request) {
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")
		if end == "" {
			end = start + "\xff"
		}
		pairs, err := db.Scan([]byte(start), []byte(end))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		type kv struct{ Key, Value string }
		out := make([]kv, len(pairs))
		for i, p := range pairs {
			out[i] = kv{Key: string(p.Key), Value: string(p.Value)}
		}
		json.NewEncoder(w).Encode(out)
	}))

	mux.HandleFunc("/api/delete", authWrite(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if err := db.Delete([]byte(key)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(200)
	}))

	mux.HandleFunc("/api/collections", auth(func(w http.ResponseWriter, r *http.Request) {
		all, err := db.Scan([]byte(""), []byte("\xff"))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		seen := map[string]int{}
		for _, p := range all {
			key := string(p.Key)
			for i := 0; i < len(key); i++ {
				if key[i] == ':' {
					seen[key[:i]]++
					break
				}
			}
		}
		json.NewEncoder(w).Encode(seen)
	}))

	mux.HandleFunc("/api/users", adminOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var body struct {
				Username, Password string
				Role                Role
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if err := createUser(db, body.Username, body.Password, body.Role); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			w.WriteHeader(201)
			return
		}

		users, err := listUsers(db)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		type safeUser struct {
			Username string
			Role     Role
		}
		out := make([]safeUser, len(users))
		for i, u := range users {
			out[i] = safeUser{Username: u.Username, Role: u.Role}
		}
		json.NewEncoder(w).Encode(out)
	}))
}