// Borrower Copilot: a stateless self-assessment tool for Indian borrowers.
// State lives only in the form payload the browser round-trips -- there is
// no database.
package main

import (
	"log"
	"net/http"

	"github.com/kaizakin/borrower-copilot/web"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", web.Index)
	mux.HandleFunc("POST /additional", web.Additional)
	mux.HandleFunc("POST /assess", web.AssessHandler)

	const addr = ":8080"
	log.Printf("Borrower Copilot listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
