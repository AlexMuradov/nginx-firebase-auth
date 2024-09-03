package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
)

func VerifySession(r *http.Request) int {

	cookie, err := r.Cookie("session")

	if err != nil {
		log.Printf("sesssion cookies are not provided - %v", err)
		return http.StatusUnauthorized
	}

	ctx := context.Background()

	opt := option.WithCredentialsFile(os.Getenv("FIREBASE_CREDENTIALS"))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Printf("error initializing app: %v\n", err)
		return http.StatusUnauthorized
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Printf("error getting Auth client: %v\n", err)
		return http.StatusUnauthorized
	}

	_, err = client.VerifySessionCookieAndCheckRevoked(r.Context(), cookie.Value)
	if err != nil {
		log.Printf("Failed to verify session cookie: %v", err)
		return http.StatusUnauthorized
	}

	return http.StatusOK
}

func authHandler(w http.ResponseWriter, r *http.Request) {

	status := VerifySession(r)

	res := map[string]int{
		"response": status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)

}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		next.ServeHTTP(w, r)
	})
}

func main() {

	if os.Getenv("FIREBASE_CREDENTIALS") == "" {
		log.Fatal("FIREBASE_CREDENTIALS not set")
	}

	r := mux.NewRouter()
	r.Use(corsMiddleware)

	r.Headers("Content-Type", "application/json")
	r.HandleFunc("/auth", authHandler).Methods("GET")

	log.Printf("Main server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
