package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"content-public-api/handler"
	"content-public-api/store"
)

func main() {
	cfg := loadConfig()

	db, err := newDB(cfg.dsn())
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}
	defer db.Close()

	contentStore := store.NewContentStore(db)
	contentHandler := handler.NewContentHandler(contentStore)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/content", contentHandler.GetSections)
	mux.HandleFunc("GET /api/v1/content/search", contentHandler.SearchContent)
	mux.HandleFunc("GET /api/v1/content/slug/{slug}", contentHandler.GetContentBySlug)

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}

type config struct {
	DBUser    string
	DBPass    string
	DBHost    string
	DBPort    string
	DBName    string
	JWTSecret string
	Port      string
}

func (c config) dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName)
}

func loadConfig() config {
	return config{
		DBUser:    getEnv("DB_USER", "nuser"),
		DBPass:    getEnv("DB_PASS", "npass"),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "3306"),
		DBName:    getEnv("DB_NAME", "ndb"),
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Port:      getEnv("PORT", "8081"),
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func newDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
