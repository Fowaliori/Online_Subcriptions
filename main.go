// @title           Online Subscriptions API
// @version         1.0
// @description     HTTP API for user subscriptions with month-granular billing windows. Subscription period fields use MM-YYYY in JSON; created_at/updated_at use RFC3339 UTC.
// @host            localhost:8080
// @BasePath        /
// @schemes         http
package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	storage "effective-mobile/internal/db"
	"effective-mobile/internal/handlers"
	"effective-mobile/internal/pkg/logger"
	"effective-mobile/internal/service"
)

func main() {
	_ = godotenv.Load()

	logger.InitDefault()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := os.Getenv("POSTGRES_HOST")
		port := os.Getenv("POSTGRES_PORT")
		user := os.Getenv("POSTGRES_USER")
		pass := os.Getenv("POSTGRES_PASSWORD")
		dbname := os.Getenv("POSTGRES_DB")
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbname)
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		slog.Error("не удалось открыть базу данных", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := storage.NewPostgresStore(db)
	svc := service.NewSubscriptionService(repo)
	h := handlers.NewHandler(svc)

	r := mux.NewRouter()
	h.RegisterRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	slog.Info("старт сервера", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("сервер завершился с ошибкой", "err", err)
		os.Exit(1)
	}
}
