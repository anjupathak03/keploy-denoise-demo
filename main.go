// Package main provides a test application to verify the auto-detect-noisy-field functionality in Keploy.
// This application exposes various API endpoints that return different types of noisy fields
// including timestamps, UUIDs, KSUIDs, ULIDs, MongoDB ObjectIDs, Snowflake IDs, and more.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"github.com/segmentio/ksuid"
)

// Response structures for various test cases

// TimestampResponse contains various timestamp formats for testing
type TimestampResponse struct {
	RFC3339    string `json:"rfc3339"`
	RFC1123    string `json:"rfc1123"`
	UnixDate   string `json:"unix_date"`
	ISO8601    string `json:"iso8601"`
	CustomDate string `json:"custom_date"`
	UnixNano   int64  `json:"unix_nano"`
	StaticName string `json:"static_name"` // Should NOT be detected as noisy
}

// UUIDResponse contains UUID fields for testing
type UUIDResponse struct {
	ID          string `json:"id"`
	SessionID   string `json:"session_id"`
	RequestID   string `json:"request_id"`
	StaticField string `json:"static_field"` // Should NOT be detected as noisy
}

// RandomIDResponse contains various random ID formats for testing
type RandomIDResponse struct {
	KSUID      string `json:"ksuid"`
	ULID       string `json:"ulid"`
	ObjectID   string `json:"object_id"`   // MongoDB ObjectID format (24 hex chars)
	SnowflakeID string `json:"snowflake_id"` // 18-19 digit number
	NanoID     string `json:"nano_id"`     // 21-22 char alphanumeric with _-
	StaticVal  string `json:"static_val"`  // Should NOT be detected as noisy
}

// PrefixedIDResponse contains prefixed hex strings for testing
type PrefixedIDResponse struct {
	EncryptionKey string `json:"encryption_key"` // enc_<hex>
	TokenID       string `json:"token_id"`       // token_<hex>
	UserID        string `json:"user_id"`        // id_<hex>
	RegularText   string `json:"regular_text"`   // Should NOT be detected as noisy
}

// HashResponse contains hash/token fields for testing
type HashResponse struct {
	SHA256Hash    string `json:"sha256_hash"`    // 64 hex chars
	APIKey        string `json:"api_key"`        // Base64-like token
	SessionToken  string `json:"session_token"`  // 32 hex chars
	RegularString string `json:"regular_string"` // Should NOT be detected as noisy
}

// NestedResponse contains nested objects with noisy fields
type NestedResponse struct {
	Data struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		ID        string `json:"id"`
		Title     string `json:"title"` // Should NOT be detected as noisy
	} `json:"data"`
	Meta struct {
		RequestID   string `json:"request_id"`
		TraceID     string `json:"trace_id"`
		ServiceName string `json:"service_name"` // Should NOT be detected as noisy
	} `json:"meta"`
}

// ArrayResponse contains arrays with noisy fields
type ArrayResponse struct {
	IDs        []string `json:"ids"`
	Timestamps []string `json:"timestamps"`
	Names      []string `json:"names"` // Should NOT be detected as noisy
}

// MixedResponse contains a mix of noisy and non-noisy fields
type MixedResponse struct {
	ID            string  `json:"id"`              // UUID - noisy
	Name          string  `json:"name"`            // static - NOT noisy
	Email         string  `json:"email"`           // static - NOT noisy
	Age           int     `json:"age"`             // static - NOT noisy
	CreatedAt     string  `json:"created_at"`      // timestamp - noisy
	Balance       float64 `json:"balance"`         // static - NOT noisy
	RequestID     string  `json:"request_id"`      // KSUID - noisy
	PhoneNumber   string  `json:"phone_number"`    // NOT noisy
	IPAddress     string  `json:"ip_address"`      // NOT noisy
}

// EdgeCaseResponse tests edge cases and boundary conditions
type EdgeCaseResponse struct {
	EmptyString      string `json:"empty_string"`
	ShortString      string `json:"short_string"`
	LongNormalString string `json:"long_normal_string"`
	PartialUUID      string `json:"partial_uuid"`       // Should be detected as hex-like
	AlmostTimestamp  string `json:"almost_timestamp"`   // Should NOT be detected
	NumberString     string `json:"number_string"`      // Should NOT be detected
	URLString        string `json:"url_string"`         // Should NOT be detected
	EmailString      string `json:"email_string"`       // Should NOT be detected
}

// generateObjectID generates a MongoDB-like ObjectID (24 hex chars)
func generateObjectID() string {
	timestamp := uint32(time.Now().Unix())
	counter := rand.Uint32()
	return fmt.Sprintf("%08x%04x%04x%08x", timestamp, rand.Intn(0xFFFF), rand.Intn(0xFFFF), counter)
}

// generateSnowflakeID generates a Twitter Snowflake-like ID (18-19 digits)
func generateSnowflakeID() string {
	epoch := int64(1288834974657) // Twitter Snowflake epoch
	timestamp := (time.Now().UnixMilli() - epoch) << 22
	sequence := rand.Int63n(4096)
	workerId := int64(rand.Intn(32)) << 12
	datacenterId := int64(rand.Intn(32)) << 17
	snowflake := timestamp | datacenterId | workerId | sequence
	return strconv.FormatInt(snowflake, 10)
}

// generateNanoID generates a NanoID-like string (21 chars)
func generateNanoID() string {
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"
	b := make([]byte, 21)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(b)
}

// generateSHA256 generates a SHA256-like hash (64 hex chars)
func generateSHA256() string {
	const hexChars = "0123456789abcdef"
	b := make([]byte, 64)
	for i := range b {
		b[i] = hexChars[rand.Intn(16)]
	}
	return string(b)
}

// generateAPIKey generates a base64-like API key
func generateAPIKey() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	b := make([]byte, 40)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// generateSessionToken generates a 32-char hex token
func generateSessionToken() string {
	const hexChars = "0123456789abcdef"
	b := make([]byte, 32)
	for i := range b {
		b[i] = hexChars[rand.Intn(16)]
	}
	return string(b)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Test Case 1: Timestamps in various formats
	http.HandleFunc("/timestamps", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Date", time.Now().Format(time.RFC1123)) // Header timestamp - should be noisy
		w.Header().Set("X-Request-Time", time.Now().Format(time.RFC3339))
		
		resp := TimestampResponse{
			RFC3339:    time.Now().Format(time.RFC3339),
			RFC1123:    time.Now().Format(time.RFC1123),
			UnixDate:   time.Now().Format(time.UnixDate),
			ISO8601:    time.Now().Format("2006-01-02"),
			CustomDate: time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"),
			UnixNano:   time.Now().UnixNano(),
			StaticName: "This is a static value",
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Test Case 2: UUIDs
	http.HandleFunc("/uuids", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", uuid.New().String()) // Header UUID - should be noisy
		
		resp := UUIDResponse{
			ID:          uuid.New().String(),
			SessionID:   uuid.New().String(),
			RequestID:   uuid.New().String(),
			StaticField: "static-value-not-uuid",
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Test Case 3: Various Random ID formats (KSUID, ULID, ObjectID, Snowflake, NanoID)
	http.HandleFunc("/random-ids", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Trace-ID", ksuid.New().String()) // Header KSUID - should be noisy
		
		resp := RandomIDResponse{
			KSUID:       ksuid.New().String(),
			ULID:        ulid.Make().String(),
			ObjectID:    generateObjectID(),
			SnowflakeID: generateSnowflakeID(),
			NanoID:      generateNanoID(),
			StaticVal:   "normal-static-value",
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Test Case 4: Prefixed hex strings
	http.HandleFunc("/prefixed-ids", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		resp := PrefixedIDResponse{
			EncryptionKey: fmt.Sprintf("enc_%s", generateSessionToken()),
			TokenID:       fmt.Sprintf("token_%s", generateSessionToken()),
			UserID:        fmt.Sprintf("id_%s", generateSessionToken()),
			RegularText:   "This is regular text content",
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Test Case 5: Hash/Token fields
	http.HandleFunc("/hashes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Auth-Token", generateAPIKey()) // Header token - should be noisy
		
		resp := HashResponse{
			SHA256Hash:    generateSHA256(),
			APIKey:        generateAPIKey(),
			SessionToken:  generateSessionToken(),
			RegularString: "This is a regular string value",
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Test Case 6: Nested objects with noisy fields
	http.HandleFunc("/nested", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		resp := NestedResponse{}
		resp.Data.CreatedAt = time.Now().Format(time.RFC3339)
		resp.Data.UpdatedAt = time.Now().Add(time.Hour).Format(time.RFC3339)
		resp.Data.ID = uuid.New().String()
		resp.Data.Title = "Sample Title"
		resp.Meta.RequestID = ksuid.New().String()
		resp.Meta.TraceID = ulid.Make().String()
		resp.Meta.ServiceName = "noisy-field-detector"
		
		json.NewEncoder(w).Encode(resp)
	})

	// Test Case 7: Arrays containing noisy fields
	http.HandleFunc("/arrays", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		resp := ArrayResponse{
			IDs: []string{
				uuid.New().String(),
				uuid.New().String(),
				uuid.New().String(),
			},
			Timestamps: []string{
				time.Now().Format(time.RFC3339),
				time.Now().Add(time.Hour).Format(time.RFC3339),
			},
			Names: []string{
				"John",
				"Jane",
				"Bob",
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Test Case 8: Mixed noisy and non-noisy fields
	http.HandleFunc("/mixed", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Correlation-ID", uuid.New().String())
		w.Header().Set("Content-Language", "en-US") // Should NOT be noisy
		
		resp := MixedResponse{
			ID:          uuid.New().String(),
			Name:        "John Doe",
			Email:       "john.doe@example.com",
			Age:         30,
			CreatedAt:   time.Now().Format(time.RFC3339),
			Balance:     1234.56,
			RequestID:   ksuid.New().String(),
			PhoneNumber: "+1-555-123-4567",
			IPAddress:   "192.168.1.100",
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Test Case 9: Edge cases and boundary conditions
	http.HandleFunc("/edge-cases", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		resp := EdgeCaseResponse{
			EmptyString:      "",
			ShortString:      "abc",
			LongNormalString: "The quick brown fox jumps over the lazy dog",
			PartialUUID:      "550e8400-e29b-41d4-a716", // Partial UUID - might be detected as hex
			AlmostTimestamp:  "not-a-date-2023",
			NumberString:     "12345",
			URLString:        "https://example.com/path",
			EmailString:      "test@example.com",
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Test Case 10: Non-JSON body (plain text) with noisy content
	http.HandleFunc("/plain-uuid", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%s", uuid.New().String())
	})

	// Test Case 11: Non-JSON body (plain text) with timestamp
	http.HandleFunc("/plain-timestamp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%s", time.Now().Format(time.RFC3339))
	})

	// Test Case 12: Non-JSON body (plain text) with non-noisy content
	http.HandleFunc("/plain-text", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "This is a regular plain text response")
	})

	// Test Case 13: Empty response
	http.HandleFunc("/empty", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Test Case 14: Headers only (no body)
	http.HandleFunc("/headers-only", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", uuid.New().String())
		w.Header().Set("X-Timestamp", time.Now().Format(time.RFC3339))
		w.Header().Set("X-Static-Header", "static-value")
		w.WriteHeader(http.StatusOK)
	})

	// Test Case 15: All noisy field types combined
	http.HandleFunc("/all-noisy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Date", time.Now().Format(time.RFC1123))
		w.Header().Set("X-Request-ID", uuid.New().String())
		w.Header().Set("X-Trace-ID", ksuid.New().String())
		
		resp := map[string]interface{}{
			"uuid":        uuid.New().String(),
			"ksuid":       ksuid.New().String(),
			"ulid":        ulid.Make().String(),
			"object_id":   generateObjectID(),
			"snowflake":   generateSnowflakeID(),
			"nano_id":     generateNanoID(),
			"timestamp":   time.Now().Format(time.RFC3339),
			"sha256":      generateSHA256(),
			"api_key":     generateAPIKey(),
			"prefixed_id": fmt.Sprintf("enc_%s", generateSessionToken()),
			"nested": map[string]interface{}{
				"created_at": time.Now().Format(time.RFC3339),
				"id":         uuid.New().String(),
				"static":     "not-noisy-value",
			},
			"id_array": []string{
				uuid.New().String(),
				uuid.New().String(),
			},
			"static_field": "This is a static non-noisy field",
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	port := ":8080"
	log.Printf("🚀 Noisy Field Detector Test Server starting on port %s", port)
	log.Printf("📝 Available endpoints for testing noisy field detection:")
	log.Printf("   GET /timestamps       - Various timestamp formats")
	log.Printf("   GET /uuids            - UUID fields")
	log.Printf("   GET /random-ids       - KSUID, ULID, ObjectID, Snowflake, NanoID")
	log.Printf("   GET /prefixed-ids     - Prefixed hex strings (enc_, token_, id_)")
	log.Printf("   GET /hashes           - SHA256, API keys, session tokens")
	log.Printf("   GET /nested           - Nested objects with noisy fields")
	log.Printf("   GET /arrays           - Arrays containing noisy fields")
	log.Printf("   GET /mixed            - Mixed noisy and non-noisy fields")
	log.Printf("   GET /edge-cases       - Edge cases and boundary conditions")
	log.Printf("   GET /plain-uuid       - Plain text UUID body")
	log.Printf("   GET /plain-timestamp  - Plain text timestamp body")
	log.Printf("   GET /plain-text       - Plain text non-noisy body")
	log.Printf("   GET /empty            - Empty response (204)")
	log.Printf("   GET /headers-only     - Noisy headers, no body")
	log.Printf("   GET /all-noisy        - All noisy field types combined")
	log.Printf("   GET /health           - Health check")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
