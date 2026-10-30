package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
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
	authorID := req.URL.Query().Get("author_id")
	sortType := req.URL.Query().Get("sort")
	log.Printf("sortType is %v", sortType)
	if authorID == "" {
		chirps, err := cfg.db.GetAllChirps(req.Context())
		if err != nil {
			log.Printf("Error when retrieving chirps: %s", err)
			respondWithError(w, 500, "Something went wrong")
			return
		}
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
		if sortType == "desc" {
			sort.Slice(jsonChirps, func(i, j int) bool {
				return jsonChirps[i].CreatedAt.After(jsonChirps[j].CreatedAt)
			})
		}
		respondWithJSON(w, 200, jsonChirps)
	} else {
		userID, err := uuid.Parse(authorID)
		if err != nil {
			log.Printf("Failed to parse the user ID from the Author ID: %s", err)
			respondWithError(w, 404, "Wrong user ID")
			return
		}
		chirps, err := cfg.db.GetChirpsFromAuthor(req.Context(), userID)
		if err != nil {
			log.Printf("Error when retrieving chirps: %s", err)
			respondWithError(w, 500, "Something went wrong")
			return
		}
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
		if sortType == "desc" {
			sort.Slice(jsonChirps, func(i, j int) bool {
				return jsonChirps[i].CreatedAt.Before(jsonChirps[j].CreatedAt)
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
		respondWithError(w, 404, "Chirp not found")
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

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error retrieving the token from the header: %v", err)
		respondWithError(w, 401, "Authentication failed")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("Error validating the user: %v", err)
		respondWithError(w, 401, "Authentication failed")
		return
	}

	parameter := req.PathValue("chirpID")
	chirpID, err := uuid.Parse(parameter)
	if err != nil {
		log.Printf("Error when parsing the Chirp ID: %s", err)
		respondWithError(w, 404, "Not a valid chirp ID")
		return
	}
	chirp, err := cfg.db.GetSingleChirp(req.Context(), chirpID)
	if err != nil {
		log.Printf("Error when retrieving chirp: %s", err)
		respondWithError(w, 404, "Chirp not found")
		return
	}

	if chirp.UserID != userID {
		w.WriteHeader(403)
		return
	}

	err = cfg.db.DeleteChirp(req.Context(), chirpID)
	if err != nil {
		log.Printf("Error when deleting chirp: %s", err)
		respondWithError(w, 500, "Failed to delete chirp")
		return
	}

	w.WriteHeader(204)
}
