package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/virgivaru/chirpy/internal/database"
)

type Chirp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, req *http.Request) {

	type Payload struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(req.Body)
	payload := Payload{}
	err := decoder.Decode(&payload)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, 500, "Something went wrong")
	} else if validateChirp(payload.Body) {
		params := database.CreateChirpParams{
			Body:   payload.Body,
			UserID: payload.UserID,
		}
		c, err := cfg.db.CreateChirp(req.Context(), params)
		chirp := Chirp{
			Id:        uuid.UUID(c.ID),
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		}
		if err != nil {
			log.Printf("Error creating chirp: %s", err)
			respondWithError(w, 500, "Couldn't create the chirp")
		} else {
			respondWithJSON(w, 201, chirp)
		}
	} else {
		respondWithError(w, 400, "Chirp is too long")
	}
}
