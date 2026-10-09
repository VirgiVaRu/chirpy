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
	pf             string
	db             *database.Queries
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
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
		pf: platform,
		db: dbQueries,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(handler)))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.chirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getSingleChirpHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.getAllChirpsHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/users", apiCfg.usersHandler)

	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Started server... Serving files from \"%s\". Port: %s\n", rootFilePath, port)
	log.Fatal(server.ListenAndServe())
}
