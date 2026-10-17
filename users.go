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
}

func (cfg *apiConfig) usersHandler(w http.ResponseWriter, req *http.Request) {

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
	} else {
		hash, err := auth.HashPassword(payload.Password)
		if err != nil {
			log.Printf("Error hashing password: %s", err)
			respondWithError(w, 500, "Something went wrong")
		}

		params := database.CreateUserParams{
			Email:          payload.Email,
			HashedPassword: hash,
		}
		user, err := cfg.db.CreateUser(req.Context(), params)
		if err != nil {
			log.Printf("Error creating user: %s", err)
			respondWithError(w, 500, "Something went wrong")
		} else {
			payload := User{
				Id:        user.ID,
				CreateAt:  user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				Email:     user.Email,
			}
			respondWithJSON(w, 201, payload)
		}
	}
}
