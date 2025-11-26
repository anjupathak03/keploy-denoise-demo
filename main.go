package main

import (
	"encoding/json"
	"fmt"
	"log"
	math_rand "math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/araddon/dateparse"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"github.com/segmentio/ksuid"
)

// ============================================================================
// DATA MODELS
// ============================================================================

// User represents a user profile with account information
// Field Types:
// - accountNumber: TEMPLATIZED BUT NOT NOISY (stable identifier)
// - userId: TEMPLATIZED AND NOISY (UUID changes each run)
// - sessionId, uuidV4, ulidId, ksuidId: NOISY BUT NOT TEMPLATIZED (library-generated IDs)
// - createdAt, requestId, processingTimeMs, serverTimestamp, parsedTimestamp: NOISY BUT NOT TEMPLATIZED
// - name, email: STABLE FIELDS (exact assertions)
type User struct {
	AccountNumber    string    `json:"accountNumber"`    // TEMPLATIZED BUT NOT NOISY
	UserID           string    `json:"userId"`           // TEMPLATIZED AND NOISY
	Name             string    `json:"name"`             // STABLE FIELD
	Email            string    `json:"email"`            // STABLE FIELD
	SessionID        string    `json:"sessionId"`        // NOISY BUT NOT TEMPLATIZED (raw UUID)
	UUIDV4           string    `json:"uuidV4"`           // NOISY BUT NOT TEMPLATIZED (google/uuid)
	ULIDID           string    `json:"ulidId"`           // NOISY BUT NOT TEMPLATIZED (oklog/ulid)
	KSUIDID          string    `json:"ksuidId"`          // NOISY BUT NOT TEMPLATIZED (segmentio/ksuid)
	CreatedAt        time.Time `json:"createdAt"`        // NOISY BUT NOT TEMPLATIZED
	ParsedTimestamp  string    `json:"parsedTimestamp"`  // NOISY BUT NOT TEMPLATIZED (dateparse)
	RequestID        string    `json:"requestId"`        // NOISY BUT NOT TEMPLATIZED
	ProcessingTimeMs int64     `json:"processingTimeMs"` // NOISY BUT NOT TEMPLATIZED
	ServerTimestamp  string    `json:"serverTimestamp"`  // NOISY BUT NOT TEMPLATIZED
}

// Order represents a customer order
// Field Types:
// - orderId: TEMPLATIZED AND NOISY (changes per order)
// - accountNumber: TEMPLATIZED BUT NOT NOISY (reused from user)
// - sessionId, uuidV4, ulidId, ksuidId: NOISY BUT NOT TEMPLATIZED (library-generated IDs)
// - amount, status, items: STABLE FIELDS
// - createdAt, requestId, processingTimeMs, transactionHash, serverTimestamp, parsedTimestamp: NOISY BUT NOT TEMPLATIZED
type Order struct {
	OrderID          string    `json:"orderId"`          // TEMPLATIZED AND NOISY
	AccountNumber    string    `json:"accountNumber"`    // TEMPLATIZED BUT NOT NOISY
	Amount           float64   `json:"amount"`           // STABLE FIELD
	Status           string    `json:"status"`           // STABLE FIELD
	Items            []string  `json:"items"`            // STABLE FIELD
	SessionID        string    `json:"sessionId"`        // NOISY BUT NOT TEMPLATIZED (raw UUID)
	UUIDV4           string    `json:"uuidV4"`           // NOISY BUT NOT TEMPLATIZED (google/uuid)
	ULIDID           string    `json:"ulidId"`           // NOISY BUT NOT TEMPLATIZED (oklog/ulid)
	KSUIDID          string    `json:"ksuidId"`          // NOISY BUT NOT TEMPLATIZED (segmentio/ksuid)
	CreatedAt        time.Time `json:"createdAt"`        // NOISY BUT NOT TEMPLATIZED
	ParsedTimestamp  string    `json:"parsedTimestamp"`  // NOISY BUT NOT TEMPLATIZED (dateparse)
	RequestID        string    `json:"requestId"`        // NOISY BUT NOT TEMPLATIZED
	ProcessingTimeMs int64     `json:"processingTimeMs"` // NOISY BUT NOT TEMPLATIZED
	TransactionHash  string    `json:"transactionHash"`  // NOISY BUT NOT TEMPLATIZED
	ServerTimestamp  string    `json:"serverTimestamp"`  // NOISY BUT NOT TEMPLATIZED
}

// Health represents health check response (control endpoint)
type Health struct {
	Status string `json:"status"` // STABLE FIELD
}

// ============================================================================
// REQUEST/RESPONSE MODELS
// ============================================================================

// CreateUserRequest for POST /users
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateOrderRequest for POST /orders
// IMPORTANT: accountNumber in request creates producer-consumer chain
type CreateOrderRequest struct {
	AccountNumber string   `json:"accountNumber"` // CONSUMER (from POST /users)
	Amount        float64  `json:"amount"`
	Items         []string `json:"items"`
}

// ErrorResponse for error cases
type ErrorResponse struct {
	Error string `json:"error"`
}

// ============================================================================
// IN-MEMORY STORAGE
// ============================================================================

var (
	// User storage: key = accountNumber
	userStore = make(map[string]User)

	// Order storage: key = orderId
	orderStore = make(map[string]Order)

	// Counters for generating IDs
	accountCounter = 0
	orderCounter   = 0

	// Thread-safe mutex
	storeMutex sync.RWMutex
)

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// generateAccountNumber creates stable account numbers like "ACC-00000001"
// This is TEMPLATIZED BUT NOT NOISY - format is consistent
func generateAccountNumber() string {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	accountCounter++
	return fmt.Sprintf("ACC-%08d", accountCounter)
}

// generateOrderID creates order IDs like "ORD-00000001"
// This is TEMPLATIZED AND NOISY - changes for each order
func generateOrderID() string {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	orderCounter++
	return fmt.Sprintf("ORD-%08d", orderCounter)
}

// generateRequestID creates unique request IDs like "REQ-uuid"
// This is NOISY BUT NOT TEMPLATIZED - changes every request
func generateRequestID() string {
	return fmt.Sprintf("REQ-%s", uuid.NewString())
}

// generateTransactionHash creates a fake hash for orders
// This is NOISY BUT NOT TEMPLATIZED - unique per transaction
func generateTransactionHash() string {
	return fmt.Sprintf("TXN-%s", uuid.NewString()[:16])
}

// getServerTimestamp returns current server time in RFC3339 format
// This is NOISY BUT NOT TEMPLATIZED - changes every call
func getServerTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// generateUUIDV4 generates a UUID v4 using google/uuid library
// This is NOISY BUT NOT TEMPLATIZED - raw UUID from library
func generateUUIDV4() string {
	return uuid.New().String()
}

// generateULID generates a ULID using oklog/ulid library
// This is NOISY BUT NOT TEMPLATIZED - time-sortable unique ID
func generateULID() string {
	entropy := ulid.Monotonic(math_rand.New(math_rand.NewSource(time.Now().UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

// generateKSUID generates a KSUID using segmentio/ksuid library
// This is NOISY BUT NOT TEMPLATIZED - K-Sortable Unique Identifier
func generateKSUID() string {
	return ksuid.New().String()
}

// parseAndFormatTimestamp demonstrates dateparse library usage
// This is NOISY BUT NOT TEMPLATIZED - parses and reformats current time
func parseAndFormatTimestamp() string {
	// Generate a timestamp string in various formats and parse it back
	timeStr := time.Now().UTC().Format("2006-01-02 15:04:05")
	parsed, err := dateparse.ParseAny(timeStr)
	if err != nil {
		return time.Now().UTC().Format(time.RFC3339Nano)
	}
	return parsed.Format(time.RFC3339Nano)
}

// writeJSON is a helper to write JSON responses consistently
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// ============================================================================
// MAIN FUNCTION
// ============================================================================

func main() {
	mux := http.NewServeMux()

	// ========================================================================
	// ENDPOINT 1: POST /users - Create User Profile (PRODUCER)
	// ========================================================================
	// Produces: accountNumber (templatized, stable), userId (templatized, noisy)
	// Also includes: requestId, processingTimeMs, serverTimestamp (noisy, not templatized)
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}

		// Validate required fields
		if req.Name == "" || req.Email == "" {
			w.Header().Set("Content-Type", "application/json")
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "name and email are required"})
			return
		}

		// Create new user profile with noisy fields
		user := User{
			AccountNumber:    generateAccountNumber(),              // TEMPLATIZED BUT NOT NOISY
			UserID:           uuid.NewString(),                     // TEMPLATIZED AND NOISY
			Name:             req.Name,                             // STABLE FIELD
			Email:            req.Email,                            // STABLE FIELD
			SessionID:        uuid.NewString(),                     // NOISY BUT NOT TEMPLATIZED (raw UUID)
			UUIDV4:           generateUUIDV4(),                     // NOISY BUT NOT TEMPLATIZED (google/uuid)
			ULIDID:           generateULID(),                       // NOISY BUT NOT TEMPLATIZED (oklog/ulid)
			KSUIDID:          generateKSUID(),                      // NOISY BUT NOT TEMPLATIZED (segmentio/ksuid)
			CreatedAt:        time.Now().UTC(),                     // NOISY BUT NOT TEMPLATIZED
			ParsedTimestamp:  parseAndFormatTimestamp(),            // NOISY BUT NOT TEMPLATIZED (dateparse)
			RequestID:        generateRequestID(),                  // NOISY BUT NOT TEMPLATIZED
			ProcessingTimeMs: time.Since(startTime).Milliseconds(), // NOISY BUT NOT TEMPLATIZED
			ServerTimestamp:  getServerTimestamp(),                 // NOISY BUT NOT TEMPLATIZED
		}

		// Store user by accountNumber
		storeMutex.Lock()
		userStore[user.AccountNumber] = user
		storeMutex.Unlock()

		log.Printf("Created user: AccountNumber=%s, UserID=%s, SessionID=%s, UUIDV4=%s, ULID=%s, KSUID=%s, RequestID=%s, ProcessingTime=%dms",
			user.AccountNumber, user.UserID, user.SessionID, user.UUIDV4, user.ULIDID, user.KSUIDID, user.RequestID, user.ProcessingTimeMs)

		// NO X-Request-Id header per requirements - only Content-Type
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, http.StatusCreated, user)
	})

	// ========================================================================
	// ENDPOINT 2: POST /orders - Create Order (CONSUMER + PRODUCER)
	// ========================================================================
	// Consumes: accountNumber (from POST /users)
	// Produces: orderId (templatized, noisy for each order)
	// Also includes: requestId, processingTimeMs, transactionHash, serverTimestamp (noisy, not templatized)
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}

		// Validate required fields
		if req.AccountNumber == "" || req.Amount <= 0 || len(req.Items) == 0 {
			w.Header().Set("Content-Type", "application/json")
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "accountNumber, amount, and items are required"})
			return
		}

		// Verify accountNumber exists (validates producer-consumer chain)
		storeMutex.RLock()
		_, exists := userStore[req.AccountNumber]
		storeMutex.RUnlock()

		if !exists {
			w.Header().Set("Content-Type", "application/json")
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "account not found"})
			return
		}

		// Create new order with noisy fields
		order := Order{
			OrderID:          generateOrderID(),                    // TEMPLATIZED AND NOISY
			AccountNumber:    req.AccountNumber,                    // TEMPLATIZED BUT NOT NOISY
			Amount:           req.Amount,                           // STABLE FIELD
			Status:           "PENDING",                            // STABLE FIELD
			Items:            req.Items,                            // STABLE FIELD
			SessionID:        uuid.NewString(),                     // NOISY BUT NOT TEMPLATIZED (raw UUID)
			UUIDV4:           generateUUIDV4(),                     // NOISY BUT NOT TEMPLATIZED (google/uuid)
			ULIDID:           generateULID(),                       // NOISY BUT NOT TEMPLATIZED (oklog/ulid)
			KSUIDID:          generateKSUID(),                      // NOISY BUT NOT TEMPLATIZED (segmentio/ksuid)
			CreatedAt:        time.Now().UTC(),                     // NOISY BUT NOT TEMPLATIZED
			ParsedTimestamp:  parseAndFormatTimestamp(),            // NOISY BUT NOT TEMPLATIZED (dateparse)
			RequestID:        generateRequestID(),                  // NOISY BUT NOT TEMPLATIZED
			ProcessingTimeMs: time.Since(startTime).Milliseconds(), // NOISY BUT NOT TEMPLATIZED
			TransactionHash:  generateTransactionHash(),            // NOISY BUT NOT TEMPLATIZED
			ServerTimestamp:  getServerTimestamp(),                 // NOISY BUT NOT TEMPLATIZED
		}

		// Store order by orderId
		storeMutex.Lock()
		orderStore[order.OrderID] = order
		storeMutex.Unlock()

		log.Printf("Created order: OrderID=%s, AccountNumber=%s, SessionID=%s, UUIDV4=%s, ULID=%s, KSUID=%s, RequestID=%s, TxHash=%s, ProcessingTime=%dms",
			order.OrderID, order.AccountNumber, order.SessionID, order.UUIDV4, order.ULIDID, order.KSUIDID, order.RequestID, order.TransactionHash, order.ProcessingTimeMs)

		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, http.StatusCreated, order)
	})

	// ========================================================================
	// ENDPOINT 3: GET /orders/{orderId} - Get Order Details (CONSUMER)
	// ========================================================================
	// Consumes: orderId (from POST /orders)
	// Shows: accountNumber linked to this order
	// Fresh noisy fields generated for each GET request
	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract orderId from path
		orderId := strings.TrimPrefix(r.URL.Path, "/orders/")
		if orderId == "" {
			w.Header().Set("Content-Type", "application/json")
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "order id required"})
			return
		}

		// Retrieve order
		storeMutex.RLock()
		order, exists := orderStore[orderId]
		storeMutex.RUnlock()

		if !exists {
			w.Header().Set("Content-Type", "application/json")
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "order not found"})
			return
		}

		// Add fresh noisy fields for this GET request
		order.SessionID = uuid.NewString()                // NEW session ID for this GET
		order.UUIDV4 = generateUUIDV4()                   // NEW UUID for this GET
		order.ULIDID = generateULID()                     // NEW ULID for this GET
		order.KSUIDID = generateKSUID()                   // NEW KSUID for this GET
		order.ParsedTimestamp = parseAndFormatTimestamp() // NEW parsed timestamp for this GET
		order.RequestID = generateRequestID()             // NEW request ID for this GET
		order.ProcessingTimeMs = time.Since(startTime).Milliseconds()
		order.ServerTimestamp = getServerTimestamp()

		log.Printf("Retrieved order: OrderID=%s, AccountNumber=%s, SessionID=%s, UUIDV4=%s, ULID=%s, KSUID=%s, RequestID=%s",
			order.OrderID, order.AccountNumber, order.SessionID, order.UUIDV4, order.ULIDID, order.KSUIDID, order.RequestID)

		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, http.StatusOK, order)
	})

	// ========================================================================
	// ENDPOINT 4: GET /users/{accountNumber} - Get User Profile (CONSUMER)
	// ========================================================================
	// Consumes: accountNumber (from POST /users)
	// Fresh noisy fields generated for each GET request
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract accountNumber from path
		accountNumber := strings.TrimPrefix(r.URL.Path, "/users/")
		if accountNumber == "" {
			w.Header().Set("Content-Type", "application/json")
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "account number required"})
			return
		}

		// Retrieve user
		storeMutex.RLock()
		user, exists := userStore[accountNumber]
		storeMutex.RUnlock()

		if !exists {
			w.Header().Set("Content-Type", "application/json")
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}

		// Add fresh noisy fields for this GET request
		user.SessionID = uuid.NewString()                // NEW session ID for this GET
		user.UUIDV4 = generateUUIDV4()                   // NEW UUID for this GET
		user.ULIDID = generateULID()                     // NEW ULID for this GET
		user.KSUIDID = generateKSUID()                   // NEW KSUID for this GET
		user.ParsedTimestamp = parseAndFormatTimestamp() // NEW parsed timestamp for this GET
		user.RequestID = generateRequestID()             // NEW request ID for this GET
		user.ProcessingTimeMs = time.Since(startTime).Milliseconds()
		user.ServerTimestamp = getServerTimestamp()

		log.Printf("Retrieved user: AccountNumber=%s, SessionID=%s, UUIDV4=%s, ULID=%s, KSUID=%s, RequestID=%s",
			user.AccountNumber, user.SessionID, user.UUIDV4, user.ULIDID, user.KSUIDID, user.RequestID)

		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, http.StatusOK, user)
	})

	// ========================================================================
	// ENDPOINT 5: GET /health - Health Check (CONTROL)
	// ========================================================================
	// Control endpoint with no noisy fields - always returns stable response
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		h := Health{Status: "ok"} // STABLE FIELD only

		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, http.StatusOK, h)
	})

	// ========================================================================
	// SERVER CONFIGURATION
	// ========================================================================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	log.Printf("🚀 Starting E-Commerce Order System on %s", addr)
	log.Printf("📋 Available endpoints:")
	log.Printf("   POST   /users                - Create user profile")
	log.Printf("   POST   /orders               - Create order")
	log.Printf("   GET    /orders/{orderId}     - Get order details")
	log.Printf("   GET    /users/{accountNumber} - Get user profile")
	log.Printf("   GET    /health               - Health check")
	log.Printf("")
	log.Printf("🎯 Field Types Demonstrated:")
	log.Printf("   - TEMPLATIZED BUT NOT NOISY: accountNumber")
	log.Printf("   - TEMPLATIZED AND NOISY: userId, orderId")
	log.Printf("   - NOISY BUT NOT TEMPLATIZED (Library-Generated):")
	log.Printf("     * sessionId (uuid.NewString)")
	log.Printf("     * uuidV4 (google/uuid)")
	log.Printf("     * ulidId (oklog/ulid - time-sortable)")
	log.Printf("     * ksuidId (segmentio/ksuid - K-Sortable UID)")
	log.Printf("     * parsedTimestamp (araddon/dateparse)")
	log.Printf("     * createdAt, requestId, processingTimeMs, transactionHash, serverTimestamp")
	log.Printf("   - STABLE FIELDS: name, email, amount, status, items")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
