package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func validateChirpHandler(w http.ResponseWriter, req *http.Request) {

	type chirp struct {
		Message string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	c := chirp{}
	err := decoder.Decode(&c)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, 500, "Something went wrong")
	} else if len(c.Message) > 140 {
		respondWithError(w, 400, "Chirp is too long")
	} else {
		type validChirp struct {
			Valid bool `json:"valid"`
		}

		payload := validChirp{
			Valid: true,
		}
		respondWithJSON(w, 200, payload)
	}
}
