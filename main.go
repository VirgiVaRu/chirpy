package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	const rootFilePath = "."

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(rootFilePath))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
	})

	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Started server... Serving files from \"%s\". Port: %s\n", rootFilePath, port)
	log.Fatal(server.ListenAndServe())
}
