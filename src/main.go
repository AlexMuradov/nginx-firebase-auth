package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
)

type AuthServer struct {
	firebaseAuth *auth.Client
}

func NewAuthServer(credentialsFile string) *AuthServer {
	ctx := context.Background()
	opt := option.WithCredentialsFile(credentialsFile)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}

	log.Println("Firebase initialized successfully")
	return &AuthServer{firebaseAuth: client}
}

func (s *AuthServer) VerifySession(r *http.Request) int {
	cookie, err := r.Cookie("session")
	if err != nil {
		log.Printf("session cookie not provided - %v", err)
		return http.StatusUnauthorized
	}

	_, err = s.firebaseAuth.VerifySessionCookieAndCheckRevoked(r.Context(), cookie.Value)
	if err != nil {
		log.Printf("failed to verify session cookie: %v", err)
		return http.StatusUnauthorized
	}

	return http.StatusOK
}

func (s *AuthServer) authHandler(w http.ResponseWriter, r *http.Request) {
	status := s.VerifySession(r)

	res := map[string]int{
		"response": status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *AuthServer) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/auth", s.authHandler).Methods("GET")
	r.HandleFunc("/healthz", healthHandler).Methods("GET")
}

func main() {
	if os.Getenv("FIREBASE_CREDENTIALS") == "" {
		log.Fatal("FIREBASE_CREDENTIALS not set")
	}

	s := NewAuthServer(os.Getenv("FIREBASE_CREDENTIALS"))

	r := mux.NewRouter()
	s.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("Server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("Server stopped")
}
