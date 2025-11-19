package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// User represents a user resource with both noisy and stable fields.
type User struct {
	ID        string    `json:"id"`        // randomId (noise)
	Name      string    `json:"name"`      // stable (must assert)
	Email     string    `json:"email"`     // stable (must assert)
	CreatedAt time.Time `json:"createdAt"` // timestamp (noise)
}

// Order represents an order resource with both noisy and stable fields.
type Order struct {
	OrderID   string    `json:"orderId"`   // random-ish ID (noise)
	UserID    string    `json:"userId"`    // stable-ish
	Amount    float64   `json:"amount"`    // stable (must assert)
	Status    string    `json:"status"`    // stable (must assert)
	Timestamp time.Time `json:"timestamp"` // timestamp (noise)
}

// Health is a simple stable response with no noise.
type Health struct {
	Status string `json:"status"`
}

func main() {
	// Seed rand for randomOrderID suffix
	rand.Seed(time.Now().UnixNano())

	mux := http.NewServeMux()

	// POST /users
	// Demonstrates:
	// - randomId in body:        "id"
	// - timestamp in body:       "createdAt"
	// - stable fields to assert: "name", "email"
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// For simplicity, ignore request body and always return same logical user.
		user := User{
			ID:        uuid.NewString(),    // random each call (noise)
			Name:      "Alice",             // stable
			Email:     "alice@example.com", // stable
			CreatedAt: time.Now().UTC(),    // time-based (noise)
		}

		// Example of noisy headers
		w.Header().Set("X-Request-Id", uuid.NewString()) // header noise
		w.Header().Set("Content-Type", "application/json")

		writeJSON(w, http.StatusCreated, user)
	})

	// GET /orders
	// Demonstrates:
	// - random-like ID in body:  "orderId"
	// - timestamp in body:       "timestamp"
	// - stable fields:           "userId", "amount", "status"
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		order := Order{
			OrderID:   randomOrderID(),  // random-like ID (noise)
			UserID:    "user-123",       // stable-ish
			Amount:    99.99,            // stable
			Status:    "PAID",           // stable
			Timestamp: time.Now().UTC(), // time-based (noise)
		}

		w.Header().Set("X-Request-Id", uuid.NewString()) // header noise
		w.Header().Set("Content-Type", "application/json")

		writeJSON(w, http.StatusOK, order)
	})

	// GET /health
	// Control endpoint:
	// - No noisy fields, always stable.
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		h := Health{Status: "ok"}

		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, http.StatusOK, h)
	})

	addr := ":8080"
	log.Printf("starting demo app on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// writeJSON is a small helper to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// randomOrderID generates a random order ID like "ORD-AB12CD34".
func randomOrderID() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return "ORD-" + string(b)
}