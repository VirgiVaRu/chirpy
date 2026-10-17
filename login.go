package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/virgivaru/chirpy/internal/auth"
	"github.com/virgivaru/chirpy/internal/database"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, req *http.Request) {
	type Payload struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	decoder := json.NewDecoder(req.Body)
	payload := Payload{}
	err := decoder.Decode(&payload)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	user, err := cfg.db.RetrieveUser(req.Context(), payload.Email)
	if err != nil {
		log.Printf("Error retrieving user: %s", err)
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	authenticated, err := auth.CheckPasswordHash(payload.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Error checking the password: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if authenticated {
		expiresIn := time.Duration(3600 * time.Second)
		token, err := auth.MakeJWT(user.ID, cfg.secret, expiresIn)
		if err != nil {
			log.Printf("Error creating the JWT: %s", err)
			respondWithError(w, 500, "Something went wrong with authorization")
			return
		}

		expiresIn = time.Duration(24 * 60 * time.Hour)
		params := database.CreateRefreshTokenParams{
			Token:     auth.MakeRefreshToken(),
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(expiresIn),
		}
		refreshToken, err := cfg.db.CreateRefreshToken(req.Context(), params)
		if err != nil {
			log.Printf("Error creating the refresh token: %s", err)
			respondWithError(w, 500, "Something went wrong with authorization")
			return
		}

		verifiedUser := User{
			Id:           user.ID,
			CreateAt:     user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			Token:        token,
			RefreshToken: refreshToken,
		}
		respondWithJSON(w, 200, verifiedUser)
	} else {
		respondWithError(w, 401, "Incorrect email or password")
	}

}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Failed to find a refresh token in the header: %s", err)
		respondWithError(w, 500, "Authentication failed")
		return
	}

	refreshToken, err := cfg.db.RetrieveRefreshToken(req.Context(), token)
	if err != nil {
		log.Printf("Failed to find the refresh token in the database: %s", err)
		respondWithError(w, 401, "Authentication failed")
		return
	}
	if time.Now().After(refreshToken.ExpiresAt) {
		respondWithError(w, 401, "Login expired")
		return
	}

	if refreshToken.RevokedAt.Valid {
		respondWithError(w, 401, "Access revoked")
		return
	}

	type Payload struct {
		Token string `json:"token"`
	}

	expiresIn := time.Duration(24 * 60 * time.Hour)
	token, err = auth.MakeJWT(refreshToken.UserID, cfg.secret, expiresIn)
	if err != nil {
		log.Printf("Error creating the JWT: %s", err)
		respondWithError(w, 500, "Something went wrong with authorization")
		return
	}
	payload := Payload{
		Token: token,
	}

	respondWithJSON(w, 200, payload)
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Failed to find a refresh token in the header: %s", err)
		respondWithError(w, 500, "Authentication failed")
		return
	}

	err = cfg.db.RevokeRefreshToken(req.Context(), token)
	if err != nil {
		log.Printf("Failed to revoke the refresh token: %s", err)
		respondWithError(w, 500, "Revoke failed")
		return
	}

	w.WriteHeader(204)
}
