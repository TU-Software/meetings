package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/TU-Software/meetings/2026-09-23/gen/db"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}

	// Database
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL is not set")
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	queries := db.New(pool)

	// Hub
	hub := NewHub(queries)
	go hub.Run()

	// Routes
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/rooms", handleListRooms(queries))
	mux.HandleFunc("POST /api/rooms", handleCreateRoom(queries, hub))
	mux.HandleFunc("GET /api/rooms/{id}", handleGetRoomMessages(queries))

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "must include username in websocket connection query")
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("error upgrading connection", "error", err)
			return
		}

		client := &Client{
			hub:      hub,
			conn:     conn,
			send:     make(chan *OutboundMessage, 256),
			username: username,
			rooms:    make(map[uuid.UUID]bool),
		}

		client.hub.register <- client

		go client.writePump()
		go client.readPump()
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello from %s\n", hostname)
	})

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server starting", "host", hostname, "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("server shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server failed to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server shutdown gracefully")
}
