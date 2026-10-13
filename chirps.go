package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/virgivaru/chirpy/internal/auth"
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
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	payload := Payload{}
	err := decoder.Decode(&payload)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if validateChirp(payload.Body) {
		tokenString, err := auth.GetBearerToken(req.Header)
		if err != nil {
			log.Printf("Failed to get the token string: %s", err)
			respondWithError(w, 500, "Something went wrong")
			return
		}
		userID, err := auth.ValidateJWT(tokenString, cfg.secret)
		if err != nil {
			log.Printf("Failed to validate the user: %s", err)
			respondWithError(w, 401, "Unauthorized access")
			return
		}
		params := database.CreateChirpParams{
			Body:   payload.Body,
			UserID: userID,
		}
		c, err := cfg.db.CreateChirp(req.Context(), params)
		chirp := Chirp{
			Id:        uuid.UUID(c.ID),
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      removeProfanity(c.Body),
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

func (cfg *apiConfig) getAllChirpsHandler(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.db.GetAllChirps(req.Context())
	if err != nil {
		log.Printf("Error when retrieving chirps: %s", err)
		respondWithError(w, 500, "Something went wrong")
	} else {
		var jsonChirps []Chirp
		for _, chirp := range chirps {
			jsonChirps = append(jsonChirps, Chirp{
				Id:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				UserID:    chirp.UserID,
			})
		}
		respondWithJSON(w, 200, jsonChirps)
	}
}

func (cfg *apiConfig) getSingleChirpHandler(w http.ResponseWriter, req *http.Request) {
	parameter := req.PathValue("chirpID")
	id, err := uuid.Parse(parameter)
	if err != nil {
		log.Printf("Error when parsing the Chirp ID: %s", err)
		respondWithError(w, 404, "Not a valid ID")
		return
	}
	chirp, err := cfg.db.GetSingleChirp(req.Context(), uuid.UUID(id))
	if err != nil {
		log.Printf("Error when retrieving chirp: %s", err)
		respondWithError(w, 404, "ID not found")
		return
	}
	payload := Chirp{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	respondWithJSON(w, 200, payload)
}
