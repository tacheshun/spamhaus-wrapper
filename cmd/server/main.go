package main

import (
	"context"
	"errors"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"log"
	"net/http"
	"os"
	"os/signal"
	"spamhaus-wrapper/graph/generated"
	"spamhaus-wrapper/internal/middleware"
	"spamhaus-wrapper/internal/repository"
	"spamhaus-wrapper/internal/resolver"
	"syscall"
	"time"
)

const defaultPort = "3030"
const defaultDBPath = "./database/ip_details.sqlite"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	ipDetailsRepo, err := repository.NewIPDetailsRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to create IP details repository: %v", err)
	}
	defer func() {
		if err := ipDetailsRepo.Close(); err != nil {
			log.Printf("Error closing repository: %v", err)
		}
	}()

	rslv := &resolver.Resolver{
		IPDetailsRepo: ipDetailsRepo,
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: rslv}))

	http.Handle("/query", middleware.BasicAuth(srv.ServeHTTP))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: http.DefaultServeMux,
	}

	go func() {
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
