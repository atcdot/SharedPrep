package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atcdot/SharedPrep/api/event/v1/eventv1connect"
	"github.com/atcdot/SharedPrep/api/item/v1/itemv1connect"
	"github.com/atcdot/SharedPrep/api/participant/v1/participantv1connect"
	"github.com/atcdot/SharedPrep/internal/api"
	"github.com/atcdot/SharedPrep/internal/auth"
	"github.com/atcdot/SharedPrep/internal/service"
	"github.com/atcdot/SharedPrep/internal/storage"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	databaseURL := envOr("DATABASE_URL", "postgres://sharedprep:sharedprep@localhost:5432/sharedprep?sslmode=disable")
	port := envOr("PORT", "8080")
	clientID := envOr("TELEGRAM_CLIENT_ID", "")
	jwtSecret := envOr("JWT_SECRET", "")

	if jwtSecret == "" {
		b := make([]byte, 32)
		rand.Read(b)
		jwtSecret = hex.EncodeToString(b)
		logger.Warn("JWT_SECRET not set, using random secret (sessions will not survive restarts)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	db, err := storage.New(ctx, databaseURL, logger)
	if err != nil {
		logger.Error("failed to init storage", "error", err)
		os.Exit(1)
	}
	cancel()
	defer db.Close()

	if err := storage.RunMigrations(context.Background(), databaseURL); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations applied")

	// auth verifier
	verifier, err := auth.NewVerifier(context.Background(), clientID)
	if err != nil {
		logger.Error("failed to init telegram verifier", "error", err)
		os.Exit(1)
	}

	// services
	eventSvc := service.NewEventService(db)
	participantSvc := service.NewParticipantService(db)
	itemSvc := service.NewItemService(db)

	// handlers
	eventHandler := api.NewEventHandler(eventSvc, participantSvc)
	participantHandler := api.NewParticipantHandler(participantSvc)
	itemHandler := api.NewItemHandler(itemSvc, participantSvc)

	// auth
	authHandler := api.NewAuthHandler(db, verifier, []byte(jwtSecret), logger)
	authMW := api.NewAuthMiddleware([]byte(jwtSecret), logger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/telegram", authHandler.TelegramLogin)
	mux.HandleFunc("GET /auth/me", authHandler.Me)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)
	mux.HandleFunc("PATCH /auth/profile", authHandler.UpdateProfile)
	mux.Handle(eventv1connect.NewEventServiceHandler(eventHandler))
	mux.Handle(participantv1connect.NewParticipantServiceHandler(participantHandler))
	mux.Handle(itemv1connect.NewItemServiceHandler(itemHandler))

	// serve SPA static files if web/dist exists
	staticDir := "web/dist"
	if _, err := os.Stat(staticDir); err == nil {
		fs := http.FileServer(http.Dir(staticDir))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == "/" {
				path = "/index.html"
			}
			if _, err := os.Stat(staticDir + path); err != nil {
				r.URL.Path = "/"
			}
			fs.ServeHTTP(w, r)
		})
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      api.CORS(authMW.Wrap(mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
