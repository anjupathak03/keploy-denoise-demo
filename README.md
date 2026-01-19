# Noisy Field Detector Test Application

This Go application is designed to test and verify Keploy's **auto-detect-noisy-field** functionality. It exposes various API endpoints that return different types of noisy fields (fields that change between test runs and should be ignored during test comparisons).

## What is Auto-Detect Noisy Field?

Keploy's auto-detect-noisy-field feature automatically identifies fields in API responses that are likely to change between requests (such as timestamps, UUIDs, session tokens, etc.) and marks them as "noisy". This helps prevent false test failures when these dynamic values differ between recorded and replayed responses.

## Detected Noisy Field Types

The auto-detect feature identifies the following types of noisy fields:

### 1. Timestamps
- RFC3339: `2023-12-18T10:30:00Z`
- RFC1123: `Mon, 18 Dec 2023 10:30:00 GMT`
- ISO8601: `2023-12-18`
- Unix timestamps
- Various date/time formats (50+ formats supported via dateparse)

### 2. Random IDs
| Type | Format | Example |
|------|--------|---------|
| UUID | 36 chars with hyphens | `550e8400-e29b-41d4-a716-446655440000` |
| KSUID | 27 alphanumeric chars | `0ujsswThIGTUYm2K8FjOOfXtY1K` |
| ULID | 26 uppercase alphanumeric | `01ARZ3NDEKTSV4RRFFQ69G5FAV` |
| MongoDB ObjectID | 24 hex chars | `507f1f77bcf86cd799439011` |
| Snowflake ID | 18-19 digits | `175928847299117063` |
| NanoID | 21-22 chars with `_` `-` | `V1StGXR8_Z5jdHi6B-myT` |

### 3. Prefixed Hex Strings
- `enc_<hex>` - Encryption keys
- `token_<hex>` - Token IDs  
- `id_<hex>` - User/resource IDs

### 4. Hashes and Tokens
- SHA256 hashes (64 hex chars)
- Base64-encoded API keys
- Session tokens (32+ hex chars)

## Installation

```bash
cd examples/noisy-field-detector
go mod tidy
```

## Running the Application

```bash
go run main.go
```

The server will start on port 8080.

## Test Cases

### Test Case 1: Timestamps (`GET /timestamps`)
**Purpose:** Verify detection of various timestamp formats in both headers and body.

**Expected Noisy Fields:**
- `header.date` - RFC1123 formatted date header
- `header.x-request-time` - RFC3339 timestamp header
- `body.rfc3339` - RFC3339 timestamp
- `body.rfc1123` - RFC1123 timestamp
- `body.unix_date` - Unix date format
- `body.iso8601` - ISO8601 date
- `body.custom_date` - Custom date format

**Non-Noisy Fields:**
- `body.static_name` - Static string value

---

### Test Case 2: UUIDs (`GET /uuids`)
**Purpose:** Verify detection of UUID v4 fields.

**Expected Noisy Fields:**
- `header.x-request-id` - UUID in header
- `body.id` - UUID field
- `body.session_id` - UUID field
- `body.request_id` - UUID field

**Non-Noisy Fields:**
- `body.static_field` - Static string

---

### Test Case 3: Random IDs (`GET /random-ids`)
**Purpose:** Verify detection of KSUID, ULID, ObjectID, Snowflake ID, and NanoID.

**Expected Noisy Fields:**
- `header.x-trace-id` - KSUID in header
- `body.ksuid` - KSUID (27 chars)
- `body.ulid` - ULID (26 chars)
- `body.object_id` - MongoDB ObjectID (24 hex chars)
- `body.snowflake_id` - Snowflake ID (18-19 digits)
- `body.nano_id` - NanoID (21 chars)

**Non-Noisy Fields:**
- `body.static_val` - Static string

---

### Test Case 4: Prefixed IDs (`GET /prefixed-ids`)
**Purpose:** Verify detection of prefixed hex strings.

**Expected Noisy Fields:**
- `body.encryption_key` - `enc_<hex>` pattern
- `body.token_id` - `token_<hex>` pattern
- `body.user_id` - `id_<hex>` pattern

**Non-Noisy Fields:**
- `body.regular_text` - Regular text content

---

### Test Case 5: Hashes (`GET /hashes`)
**Purpose:** Verify detection of cryptographic hashes and tokens.

**Expected Noisy Fields:**
- `header.x-auth-token` - API key in header
- `body.sha256_hash` - 64-char hex hash
- `body.api_key` - Base64-like token
- `body.session_token` - 32-char hex token

**Non-Noisy Fields:**
- `body.regular_string` - Regular string

---

### Test Case 6: Nested Objects (`GET /nested`)
**Purpose:** Verify detection in nested JSON structures.

**Expected Noisy Fields:**
- `body.data.created_at` - Nested timestamp
- `body.data.updated_at` - Nested timestamp
- `body.data.id` - Nested UUID
- `body.meta.request_id` - Nested KSUID
- `body.meta.trace_id` - Nested ULID

**Non-Noisy Fields:**
- `body.data.title` - Static title
- `body.meta.service_name` - Static service name

---

### Test Case 7: Arrays (`GET /arrays`)
**Purpose:** Verify detection of noisy fields within arrays.

**Expected Noisy Fields:**
- `body.ids.0`, `body.ids.1`, `body.ids.2` - Array of UUIDs
- `body.timestamps.0`, `body.timestamps.1` - Array of timestamps

**Non-Noisy Fields:**
- `body.names.0`, `body.names.1`, `body.names.2` - Array of static names

---

### Test Case 8: Mixed Fields (`GET /mixed`)
**Purpose:** Verify correct differentiation between noisy and non-noisy fields.

**Expected Noisy Fields:**
- `header.x-correlation-id` - UUID header
- `body.id` - UUID
- `body.created_at` - Timestamp
- `body.request_id` - KSUID

**Non-Noisy Fields:**
- `header.content-language` - Static header
- `body.name` - Name string
- `body.email` - Email address
- `body.age` - Integer
- `body.balance` - Float
- `body.phone_number` - Phone number
- `body.ip_address` - IP address

---

### Test Case 9: Edge Cases (`GET /edge-cases`)
**Purpose:** Test boundary conditions and potential false positives.

**Expected Non-Noisy Fields:**
- `body.empty_string` - Empty string
- `body.short_string` - Too short (3 chars)
- `body.long_normal_string` - Regular sentence
- `body.almost_timestamp` - Invalid date format
- `body.number_string` - Simple number
- `body.url_string` - URL
- `body.email_string` - Email

**Potentially Noisy:**
- `body.partial_uuid` - Partial hex string (may be detected)

---

### Test Case 10: Plain Text UUID (`GET /plain-uuid`)
**Purpose:** Verify detection when entire body is a UUID.

**Expected Noisy Fields:**
- `body` - Entire body marked as noisy

---

### Test Case 11: Plain Text Timestamp (`GET /plain-timestamp`)
**Purpose:** Verify detection when entire body is a timestamp.

**Expected Noisy Fields:**
- `body` - Entire body marked as noisy

---

### Test Case 12: Plain Text (`GET /plain-text`)
**Purpose:** Verify non-detection of regular text.

**Expected Noisy Fields:**
- None

---

### Test Case 13: Empty Response (`GET /empty`)
**Purpose:** Verify handling of empty responses.

**Expected Noisy Fields:**
- None

---

### Test Case 14: Headers Only (`GET /headers-only`)
**Purpose:** Verify detection in headers when body is empty.

**Expected Noisy Fields:**
- `header.x-request-id` - UUID header
- `header.x-timestamp` - Timestamp header

**Non-Noisy Fields:**
- `header.x-static-header` - Static header

---

### Test Case 15: All Noisy Types (`GET /all-noisy`)
**Purpose:** Comprehensive test with all noisy field types.

**Expected Noisy Fields:**
- All header timestamps and IDs
- `body.uuid`, `body.ksuid`, `body.ulid`
- `body.object_id`, `body.snowflake`, `body.nano_id`
- `body.timestamp`, `body.sha256`, `body.api_key`
- `body.prefixed_id`
- `body.nested.created_at`, `body.nested.id`
- `body.id_array.0`, `body.id_array.1`

**Non-Noisy Fields:**
- `body.nested.static`
- `body.static_field`

---

## Testing with Keploy

### Step 1: Record API Calls

```bash
# Start the application with Keploy recording
keploy record -c "go run main.go"

# In another terminal, make requests to all endpoints
curl http://localhost:8080/timestamps
curl http://localhost:8080/uuids
curl http://localhost:8080/random-ids
curl http://localhost:8080/prefixed-ids
curl http://localhost:8080/hashes
curl http://localhost:8080/nested
curl http://localhost:8080/arrays
curl http://localhost:8080/mixed
curl http://localhost:8080/edge-cases
curl http://localhost:8080/plain-uuid
curl http://localhost:8080/plain-timestamp
curl http://localhost:8080/plain-text
curl http://localhost:8080/empty
curl http://localhost:8080/headers-only
curl http://localhost:8080/all-noisy
```

### Step 2: Verify Recorded Test Cases

Check the generated test cases in `keploy/test-set-0/` directory. Each test file should have a `noise` section listing the automatically detected noisy fields.

Example expected noise configuration:
```yaml
noise:
  header.date: []
  header.x-request-id: []
  body.id: []
  body.created_at: []
  body.request_id: []
```

### Step 3: Run Tests

```bash
keploy test -c "go run main.go"
```

All tests should pass because the noisy fields are automatically excluded from comparison.

## Validation Checklist

Use this checklist to verify the auto-detect functionality:

- [ ] Timestamps in various formats are detected
- [ ] UUIDs (v1, v4) are detected
- [ ] KSUIDs (27 chars) are detected
- [ ] ULIDs (26 chars) are detected
- [ ] MongoDB ObjectIDs (24 hex) are detected
- [ ] Snowflake IDs (18-19 digits) are detected
- [ ] NanoIDs (21-22 chars) are detected
- [ ] Prefixed hex strings (`enc_`, `token_`, `id_`) are detected
- [ ] SHA256 hashes (64 hex chars) are detected
- [ ] Base64-like API keys are detected
- [ ] Session tokens (32+ hex chars) are detected
- [ ] Nested object fields are detected
- [ ] Array elements are individually detected
- [ ] Header fields are detected
- [ ] Plain text noisy bodies are detected
- [ ] Regular static strings are NOT detected as noisy
- [ ] Email addresses are NOT detected as noisy
- [ ] Phone numbers are NOT detected as noisy
- [ ] IP addresses are NOT detected as noisy
- [ ] URLs are NOT detected as noisy
- [ ] Short strings are NOT detected as noisy

## Troubleshooting

### False Positives
If a field is incorrectly marked as noisy, you can manually exclude it in the Keploy config:

```yaml
test:
  noise:
    - body.some_field: ["exclude"]
```

### False Negatives
If a noisy field is not being detected, you can manually add it:

```yaml
test:
  noise:
    - body.custom_id: []
```

## Contributing

If you find additional noisy field patterns that should be auto-detected, please open an issue or PR in the main Keploy repository.

## License

This test application is part of the Keploy project and is licensed under the same terms.
