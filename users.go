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

type User struct {
	Id           uuid.UUID `json:"id"`
	CreateAt     time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, req *http.Request) {

	type Payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(req.Body)
	payload := Payload{}
	err := decoder.Decode(&payload)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	hash, err := auth.HashPassword(payload.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	params := database.CreateUserParams{
		Email:          payload.Email,
		HashedPassword: hash,
	}
	user, err := cfg.db.CreateUser(req.Context(), params)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	newUser := User{
		Id:          user.ID,
		CreateAt:    user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
	respondWithJSON(w, 201, newUser)
}

func (cfg *apiConfig) editUsersHandler(w http.ResponseWriter, req *http.Request) {

	type Payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(req.Body)
	payload := Payload{}
	err := decoder.Decode(&payload)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error recovering the access token from the header: %s", err)
		respondWithError(w, 401, "Authentication failed")
		return
	}
	tokenUserID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("Error validating the JWT: %s", err)
		respondWithError(w, 401, "Authentication failed")
		return
	}

	hash, err := auth.HashPassword(payload.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	params := database.UpdateUserParams{
		ID:             tokenUserID,
		Email:          payload.Email,
		HashedPassword: hash,
	}
	err = cfg.db.UpdateUser(req.Context(), params)
	if err != nil {
		log.Printf("Error updating user: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user, err := cfg.db.RetrieveUser(req.Context(), payload.Email)
	if err != nil {
		log.Printf("Error retrieving the user: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	updatedUser := User{
		Id:          user.ID,
		CreateAt:    user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
	respondWithJSON(w, 200, updatedUser)
}

func (cfg *apiConfig) upgradeUsersHandler(w http.ResponseWriter, req *http.Request) {
	type RequestBody struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(req.Body)
	requestBody := RequestBody{}
	err := decoder.Decode(&requestBody)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if requestBody.Event != "user.upgraded" {
		log.Printf("Wrong event, cannot upgrade. Expected: \"user.upgraded\" Got: %s", requestBody.Event)
		w.WriteHeader(204)
		return
	}

	err = cfg.db.UpgradeUser(req.Context(), requestBody.Data.UserID)
	if err != nil {
		log.Printf("Failed to upgrade user to chirpy red: %v", err)
		respondWithError(w, 404, "Could not upgrade user")
		return
	}

	log.Printf("User upgraded to chirpy red successfully")
	w.WriteHeader(204)

}
