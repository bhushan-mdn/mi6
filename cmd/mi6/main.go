package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"mi6/internal/agent"
	"mi6/internal/api"
	"mi6/internal/db"
)

const dbPath = "agents.db"

func must(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	port := flag.String("port", "6969", "port to run the main MI6 server on")
	flag.Parse()

	// 1. Initialize DB and Repository
	dbConn, err := sql.Open("sqlite3", dbPath)
	must(err)
	defer dbConn.Close()

	// Run migrations (ensure tables exist)
	must(db.RunMigrations(dbConn))

	// Instantiate the concrete repository implementation
	repo := db.NewSQLiteRepository(dbConn)

	// 2. Initialize Agent Registry
	mgr := agent.NewRegistry(repo)

	// 3. Setup Router and Handlers
	r := api.NewRouter(mgr)

	// 4. Start Main Server
	srv := &http.Server{
		Addr:    ":" + *port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("starting MI6 main server on :%s", *port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("MI6 main server failed: %v", err)
		}
	}()

	// Block until OS signal received
	<-quit

	// 5. Graceful Shutdown
	log.Println("shutting down MI6 main server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the main API server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("MI6 main server forced to shutdown: %v", err)
	}
	log.Println("MI6 main server stopped.")

	// Shutdown all dynamic Agents
	mgr.ShutdownAll()
}
