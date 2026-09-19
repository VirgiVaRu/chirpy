package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	const rootFilePath = "."

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))

	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Started server... Serving files from \"%s\". Port: %s\n", rootFilePath, port)
	log.Fatal(server.ListenAndServe())
}
