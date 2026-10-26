package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/virgivaru/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	secret         string
	pk             string
	pf             string
	db             *database.Queries
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("SECRET")
	polka_key := os.Getenv("POLKA_KEY")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)
	if err != nil {
		log.Printf("Error opening the data base: %s", err)
	}

	const port = "8080"
	const rootFilePath = "."
	handler := http.FileServer(http.Dir(rootFilePath))
	apiCfg := apiConfig{
		pf:     platform,
		db:     dbQueries,
		secret: secret,
		pk:     polka_key,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(handler)))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.chirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getSingleChirpHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirpHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.getAllChirpsHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/users", apiCfg.createUserHandler)
	mux.HandleFunc("PUT /api/users", apiCfg.editUsersHandler)
	mux.HandleFunc("POST /api/login", apiCfg.loginHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.refreshHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeHandler)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.upgradeUsersHandler)

	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Started server... Serving files from \"%s\". Port: %s\n", rootFilePath, port)
	log.Fatal(server.ListenAndServe())
}
