package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"uuid"
)

type User struct {
	Id        uuid.UUID `json:"id"`
	CreateAt  time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) usersHandler(w http.ResponseWriter, req *http.Request) {

	type parameters struct {
		Email string
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, 500, "Something went wrong")
	} else {
		user, err := cfg.db.CreateUser(req.Context(), params.Email)
		if err != nil {
			log.Printf("Error creating user: %s", err)
			respondWithError(w, 500, "Something went wrong")
		} else {
			payload := User{
				Id:        uuid.UUID(user.ID),
				CreateAt:  user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				Email:     user.Email,
			}
			respondWithJSON(w, 201, payload)
		}
	}
}
