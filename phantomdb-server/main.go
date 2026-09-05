package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/TobiGrant-byte/phantomdb"
)

func main() {
	dbPath := flag.String("db", "phantomdb-data", "path to the database file (without extension)")
	port := flag.String("port", "8080", "port to listen on")
	bootstrapAdmin := flag.String("bootstrap-admin", "", "username:password — creates an initial admin user if it doesn't already exist")
	certFile := flag.String("cert", "cert.pem", "path to TLS certificate")
	keyFile := flag.String("key", "key.pem", "path to TLS private key")
	useTLS := flag.Bool("tls", false, "enable HTTPS (requires --cert and --key)")
	flag.Parse()

	db, err := phantomdb.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if *bootstrapAdmin != "" {
		parts := strings.SplitN(*bootstrapAdmin, ":", 2)
		if len(parts) == 2 {
			err := createUser(db, parts[0], parts[1], RoleAdmin)
			if err != nil && err.Error() != "user already exists" {
				log.Fatal("bootstrap failed:", err)
			}
			if err == nil {
				fmt.Printf("  Created admin user: %s\n", parts[0])
			}
		}
	}

	mux := http.NewServeMux()
	registerAPI(mux, db)
	registerStudio(mux, studioFiles)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("  Database file: %s.pdb\n\n", *dbPath)

	if *useTLS {
		fmt.Printf("  PhantomDB Server running at https://localhost:%s\n\n", *port)
		if err := http.ListenAndServeTLS(addr, *certFile, *keyFile, mux); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("  PhantomDB Server running at http://localhost:%s (NOT encrypted — use --tls for production)\n\n", *port)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal(err)
		}
	}
}