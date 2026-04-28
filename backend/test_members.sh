#!/bin/bash

# Test script for GET /api/projects/:id/members endpoint

echo "=== Testing GET /api/projects/:id/members ==="
echo ""

# JWT Secret (from .env or default for testing)
JWT_SECRET="your-secure-jwt-secret-here-min-32-chars"

# Generate a JWT token for user-1
# Using base64 encoding for header and payload, then signing
HEADER=$(echo -n '{"alg":"HS256","typ":"JWT"}' | base64 -w0 | tr '+/' '-_' | tr -d '=')
PAYLOAD=$(echo -n "{\"user_id\":\"user-1\",\"email\":\"test@example.com\",\"exp\":$(date -d "+24 hours" +%s),\"iat\":$(date +%s),\"iss\":\"taskify\"}" | base64 -w0 | tr '+/' '-_' | tr -d '=')

# Create signature manually (simplified for testing)
SIGNATURE=$(echo -n "${HEADER}.${PAYLOAD}" | openssl dgst -sha256 -hmac "${JWT_SECRET}" -binary | base64 -w0 | tr '+/' '-_' | tr -d '=')

TOKEN="${HEADER}.${PAYLOAD}.${SIGNATURE}"

echo "Generated Token: ${TOKEN:0:50}..."
echo ""

# Test 1: Request without Authorization header (should return 401)
echo "--- Test 1: No Authorization Header ---"
curl -s -w "\nHTTP Status: %{http_code}\n" http://localhost:8080/api/projects/1/members
echo ""

# Test 2: Request with valid JWT (should return 200 with members)
echo ""
echo "--- Test 2: Valid JWT - Get Project Members ---"
curl -s -w "\nHTTP Status: %{http_code}\n" \
  -H "Authorization: Bearer ${TOKEN}" \
  http://localhost:8080/api/projects/1/members | jq .
echo ""

# Test 3: Request with non-existent project (should return 404)
echo "--- Test 3: Non-existent Project ---"
curl -s -w "\nHTTP Status: %{http_code}\n" \
  -H "Authorization: Bearer ${TOKEN}" \
  http://localhost:8080/api/projects/999/members
echo ""

# Test 4: Request with pagination
echo ""
echo "--- Test 4: With Pagination ---"
curl -s -w "\nHTTP Status: %{http_code}\n" \
  -H "Authorization: Bearer ${TOKEN}" \
  "http://localhost:8080/api/projects/1/members?page=1&limit=10" | jq .
echo ""

echo "=== Tests Complete ==="
