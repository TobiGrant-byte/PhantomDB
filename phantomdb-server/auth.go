package main

import (
	"net/http"

	"github.com/TobiGrant-byte/phantomdb"
)

func requireAuth(db *phantomdb.DB, minRole Role) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="phantomdb"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := getUser(db, username)
			if err != nil || !verifyPassword(user, password) {
				w.Header().Set("WWW-Authenticate", `Basic realm="phantomdb"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if !roleSatisfies(user.Role, minRole) {
				http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}

func roleSatisfies(userRole, required Role) bool {
	rank := map[Role]int{RoleRead: 1, RoleWrite: 2, RoleAdmin: 3}
	return rank[userRole] >= rank[required]
}