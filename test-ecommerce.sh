#!/bin/bash

echo "🎬 Testing E-Commerce Order System with Templatization..."
echo ""

# =============================================================================
# Test 1: Create User Profile (PRODUCER)
# =============================================================================
echo "=== Test 1: Create User (Producer) ==="
USER_RESPONSE=$(curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Johnson","email":"alice@example.com"}')

echo "$USER_RESPONSE"

ACCOUNT_NUMBER=$(echo "$USER_RESPONSE" | grep -o '"accountNumber":"[^"]*"' | cut -d'"' -f4)
USER_ID=$(echo "$USER_RESPONSE" | grep -o '"userId":"[^"]*"' | cut -d'"' -f4)
SESSION_ID_1=$(echo "$USER_RESPONSE" | grep -o '"sessionId":"[^"]*"' | cut -d'"' -f4)
UUID_V4_1=$(echo "$USER_RESPONSE" | grep -o '"uuidV4":"[^"]*"' | cut -d'"' -f4)
ULID_1=$(echo "$USER_RESPONSE" | grep -o '"ulidId":"[^"]*"' | cut -d'"' -f4)
KSUID_1=$(echo "$USER_RESPONSE" | grep -o '"ksuidId":"[^"]*"' | cut -d'"' -f4)
PARSED_TS_1=$(echo "$USER_RESPONSE" | grep -o '"parsedTimestamp":"[^"]*"' | cut -d'"' -f4)
REQUEST_ID_1=$(echo "$USER_RESPONSE" | grep -o '"requestId":"[^"]*"' | cut -d'"' -f4)
PROCESSING_TIME_1=$(echo "$USER_RESPONSE" | grep -o '"processingTimeMs":[0-9]*' | cut -d':' -f2)

echo "📝 Account Number: $ACCOUNT_NUMBER (TEMPLATIZED BUT NOT NOISY)"
echo "📝 User ID: $USER_ID (TEMPLATIZED AND NOISY)"
echo "🔊 Session ID: $SESSION_ID_1 (NOISY - uuid.NewString)"
echo "🔊 UUID v4: $UUID_V4_1 (NOISY - google/uuid)"
echo "🔊 ULID: $ULID_1 (NOISY - oklog/ulid)"
echo "🔊 KSUID: $KSUID_1 (NOISY - segmentio/ksuid)"
echo "🔊 Parsed Timestamp: $PARSED_TS_1 (NOISY - dateparse)"
echo "🔊 Request ID: $REQUEST_ID_1 (NOISY - custom)"
echo "🔊 Processing Time: ${PROCESSING_TIME_1}ms (NOISY)"
echo ""

sleep 2

# =============================================================================
# Test 2: Create Order 1 (CONSUMER + PRODUCER)
# =============================================================================
echo "=== Test 2: Create Order 1 (Consumer of accountNumber) ==="
ORDER1_RESPONSE=$(curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d "{\"accountNumber\":\"$ACCOUNT_NUMBER\",\"amount\":99.99,\"items\":[\"Laptop\",\"Mouse\"]}")

echo "$ORDER1_RESPONSE"

ORDER1_ID=$(echo "$ORDER1_RESPONSE" | grep -o '"orderId":"[^"]*"' | cut -d'"' -f4)
SESSION_ID_2=$(echo "$ORDER1_RESPONSE" | grep -o '"sessionId":"[^"]*"' | cut -d'"' -f4)
UUID_V4_2=$(echo "$ORDER1_RESPONSE" | grep -o '"uuidV4":"[^"]*"' | cut -d'"' -f4)
ULID_2=$(echo "$ORDER1_RESPONSE" | grep -o '"ulidId":"[^"]*"' | cut -d'"' -f4)
KSUID_2=$(echo "$ORDER1_RESPONSE" | grep -o '"ksuidId":"[^"]*"' | cut -d'"' -f4)
PARSED_TS_2=$(echo "$ORDER1_RESPONSE" | grep -o '"parsedTimestamp":"[^"]*"' | cut -d'"' -f4)
REQUEST_ID_2=$(echo "$ORDER1_RESPONSE" | grep -o '"requestId":"[^"]*"' | cut -d'"' -f4)
TX_HASH_1=$(echo "$ORDER1_RESPONSE" | grep -o '"transactionHash":"[^"]*"' | cut -d'"' -f4)
PROCESSING_TIME_2=$(echo "$ORDER1_RESPONSE" | grep -o '"processingTimeMs":[0-9]*' | cut -d':' -f2)

echo "📝 Order 1 ID: $ORDER1_ID (TEMPLATIZED AND NOISY)"
echo "🔊 Session ID: $SESSION_ID_2 (NOISY - uuid.NewString)"
echo "🔊 UUID v4: $UUID_V4_2 (NOISY - google/uuid)"
echo "🔊 ULID: $ULID_2 (NOISY - oklog/ulid)"
echo "🔊 KSUID: $KSUID_2 (NOISY - segmentio/ksuid)"
echo "🔊 Parsed Timestamp: $PARSED_TS_2 (NOISY - dateparse)"
echo "🔊 Request ID: $REQUEST_ID_2 (NOISY - custom)"
echo "🔊 Transaction Hash: $TX_HASH_1 (NOISY)"
echo "🔊 Processing Time: ${PROCESSING_TIME_2}ms (NOISY)"
echo ""

sleep 2

# =============================================================================
# Test 3: Get Order 1 Details (CONSUMER)
# =============================================================================
echo "=== Test 3: Get Order 1 Details (Consumer of orderId) ==="
ORDER1_DETAILS=$(curl -s http://localhost:8080/orders/$ORDER1_ID)
echo "$ORDER1_DETAILS" | jq .

SESSION_ID_3=$(echo "$ORDER1_DETAILS" | grep -o '"sessionId":"[^"]*"' | cut -d'"' -f4)
REQUEST_ID_3=$(echo "$ORDER1_DETAILS" | grep -o '"requestId":"[^"]*"' | cut -d'"' -f4)
echo "🔊 Fresh Session ID: $SESSION_ID_3 (NOISY - changes on each GET)"
echo "🔊 Fresh Request ID: $REQUEST_ID_3 (NOISY - changes on each GET)"
echo ""

sleep 2

# =============================================================================
# Test 4: Create Order 2 (Same accountNumber, different orderId)
# =============================================================================
echo "=== Test 4: Create Order 2 (Same Account, Different Order) ==="
ORDER2_RESPONSE=$(curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d "{\"accountNumber\":\"$ACCOUNT_NUMBER\",\"amount\":149.99,\"items\":[\"Keyboard\",\"Monitor\"]}")

echo "$ORDER2_RESPONSE"

ORDER2_ID=$(echo "$ORDER2_RESPONSE" | grep -o '"orderId":"[^"]*"' | cut -d'"' -f4)
SESSION_ID_4=$(echo "$ORDER2_RESPONSE" | grep -o '"sessionId":"[^"]*"' | cut -d'"' -f4)
REQUEST_ID_4=$(echo "$ORDER2_RESPONSE" | grep -o '"requestId":"[^"]*"' | cut -d'"' -f4)
TX_HASH_2=$(echo "$ORDER2_RESPONSE" | grep -o '"transactionHash":"[^"]*"' | cut -d'"' -f4)

echo "📝 Order 2 ID: $ORDER2_ID (TEMPLATIZED AND NOISY)"
echo "🔊 Session ID: $SESSION_ID_4 (NOISY - different from Test 2)"
echo "🔊 Request ID: $REQUEST_ID_4 (NOISY - different from Test 2)"
echo "🔊 Transaction Hash: $TX_HASH_2 (NOISY - different from Test 2)"
echo ""

sleep 2

# =============================================================================
# Test 5: Get Order 2 Details (CONSUMER)
# =============================================================================
echo "=== Test 5: Get Order 2 Details (Consumer of orderId) ==="
ORDER2_DETAILS=$(curl -s http://localhost:8080/orders/$ORDER2_ID)
echo "$ORDER2_DETAILS" | jq .

SESSION_ID_5=$(echo "$ORDER2_DETAILS" | grep -o '"sessionId":"[^"]*"' | cut -d'"' -f4)
REQUEST_ID_5=$(echo "$ORDER2_DETAILS" | grep -o '"requestId":"[^"]*"' | cut -d'"' -f4)
echo "🔊 Fresh Session ID: $SESSION_ID_5 (NOISY - changes on each GET)"
echo "🔊 Fresh Request ID: $REQUEST_ID_5 (NOISY - changes on each GET)"
echo ""

sleep 2

# =============================================================================
# Test 6: Get User Profile by Account Number (CONSUMER)
# =============================================================================
echo "=== Test 6: Get User Profile (Consumer of accountNumber) ==="
USER_DETAILS=$(curl -s http://localhost:8080/users/$ACCOUNT_NUMBER)
echo "$USER_DETAILS" | jq .

SESSION_ID_6=$(echo "$USER_DETAILS" | grep -o '"sessionId":"[^"]*"' | cut -d'"' -f4)
REQUEST_ID_6=$(echo "$USER_DETAILS" | grep -o '"requestId":"[^"]*"' | cut -d'"' -f4)
echo "🔊 Fresh Session ID: $SESSION_ID_6 (NOISY - different from Test 1)"
echo "🔊 Fresh Request ID: $REQUEST_ID_6 (NOISY - different from Test 1)"
echo ""

sleep 2

# =============================================================================
# Test 7: Health Check (CONTROL)
# =============================================================================
echo "=== Test 7: Health Check (Control - No Noisy Fields) ==="
curl -s http://localhost:8080/health | jq .
echo ""

# =============================================================================
# Summary
# =============================================================================
echo "════════════════════════════════════════════════════════════════════"
echo "🎉 All tests completed!"
echo "════════════════════════════════════════════════════════════════════"
echo ""
echo "📊 TEMPLATIZED VALUES (tracked across tests):"
echo "  ✅ Account Number: $ACCOUNT_NUMBER (TEMPLATIZED BUT NOT NOISY)"
echo "  ✅ User ID: $USER_ID (TEMPLATIZED AND NOISY)"
echo "  ✅ Order 1 ID: $ORDER1_ID (TEMPLATIZED AND NOISY)"
echo "  ✅ Order 2 ID: $ORDER2_ID (TEMPLATIZED AND NOISY)"
echo ""
echo "🔊 NOISY VALUES (ignored during comparison):"
echo "  ❌ Library-Generated IDs:"
echo "     Session IDs (uuid.NewString):"
echo "       - Test 1: $SESSION_ID_1"
echo "       - Test 2: $SESSION_ID_2"
echo "     UUID v4 (google/uuid): $UUID_V4_1, $UUID_V4_2"
echo "     ULID (oklog/ulid): $ULID_1, $ULID_2"
echo "     KSUID (segmentio/ksuid): $KSUID_1, $KSUID_2"
echo "     Parsed Timestamps (dateparse): $PARSED_TS_1, $PARSED_TS_2"
echo ""
echo "  ❌ Request IDs:"
echo "     - Test 1: $REQUEST_ID_1"
echo "     - Test 2: $REQUEST_ID_2"
echo "     - Test 3: $REQUEST_ID_3 (fresh on GET)"
echo "     - Test 4: $REQUEST_ID_4"
echo "     - Test 5: $REQUEST_ID_5 (fresh on GET)"
echo "     - Test 6: $REQUEST_ID_6 (fresh on GET)"
echo ""
echo "  ❌ Transaction Hashes:"
echo "     - Order 1: $TX_HASH_1"
echo "     - Order 2: $TX_HASH_2"
echo ""
echo "  ❌ Processing Times:"
echo "     - Test 1: ${PROCESSING_TIME_1}ms"
echo "     - Test 2: ${PROCESSING_TIME_2}ms"
echo "     - (varies on each request)"
echo ""
echo "  ❌ Timestamps (createdAt, serverTimestamp):"
echo "     - Change on every test execution"
echo ""
echo "════════════════════════════════════════════════════════════════════"
echo "✅ Expected Templatization After 'keploy templatize':"
echo "════════════════════════════════════════════════════════════════════"
echo ""
echo "  1. {{.accountNumber}}    → TEMPLATIZED BUT NOT NOISY"
echo "                              (stable, reused in tests 2, 4, 6)"
echo ""
echo "  2. {{.userId}}            → TEMPLATIZED AND NOISY"
echo "                              (UUID changes each run, tracked)"
echo ""
echo "  3. {{.orderId_1}}         → TEMPLATIZED AND NOISY"
echo "                              (used in test 3)"
echo ""
echo "  4. {{.orderId_2}}         → TEMPLATIZED AND NOISY"
echo "                              (used in test 5)"
echo ""
echo "  5. sessionId              → NOISY BUT NOT TEMPLATIZED"
echo "                              (raw UUID, ignored during comparison)"
echo ""
echo "  6. requestId              → NOISY BUT NOT TEMPLATIZED"
echo "                              (ignored during comparison)"
echo ""
echo "  7. transactionHash        → NOISY BUT NOT TEMPLATIZED"
echo "                              (ignored during comparison)"
echo ""
echo "  8. processingTimeMs       → NOISY BUT NOT TEMPLATIZED"
echo "                              (ignored during comparison)"
echo ""
echo "  9. createdAt              → NOISY BUT NOT TEMPLATIZED"
echo "                              (ignored during comparison)"
echo ""
echo " 10. serverTimestamp        → NOISY BUT NOT TEMPLATIZED"
echo "                              (ignored during comparison)"
echo ""
echo "════════════════════════════════════════════════════════════════════"
echo "🎓 Field Type Summary:"
echo "════════════════════════════════════════════════════════════════════"
echo ""
echo "  📌 TEMPLATIZED BUT NOT NOISY (1 field):"
echo "     - accountNumber (stable identifier, reused)"
echo ""
echo "  📌 TEMPLATIZED AND NOISY (3 fields):"
echo "     - userId (UUID changes each run)"
echo "     - orderId (different per order)"
echo ""
echo "  📌 NOISY BUT NOT TEMPLATIZED (10 fields):"
echo "     Library-Generated IDs:"
echo "       - sessionId (uuid.NewString)"
echo "       - uuidV4 (google/uuid)"
echo "       - ulidId (oklog/ulid - time-sortable)"
echo "       - ksuidId (segmentio/ksuid - K-Sortable UID)"
echo "     Timestamps & Others:"
echo "       - parsedTimestamp (araddon/dateparse)"
echo "       - createdAt (time.Now)"
echo "       - serverTimestamp (time.Now formatted)"
echo "       - requestId (custom UUID-based)"
echo "       - transactionHash (custom UUID-based)"
echo "       - processingTimeMs (varies by execution)"
echo ""
echo "  📌 STABLE FIELDS (5 fields):"
echo "     - name, email (user fields)"
echo "     - amount, status, items (order fields)"
echo ""
echo "════════════════════════════════════════════════════════════════════"
echo "🚀 Next Steps:"
echo "════════════════════════════════════════════════════════════════════"
echo ""
echo "  1. Run: keploy templatize"
echo "  2. Check: cat keploy/test-set-0/test-1.yaml | grep -E '(sessionId|uuidV4|ulidId|ksuidId|parsedTimestamp|requestId|transactionHash|processingTimeMs)'"
echo "  3. Run: keploy test -c 'go run main.go'"
echo "  4. Verify: All 7 tests pass despite different noisy values"
echo ""