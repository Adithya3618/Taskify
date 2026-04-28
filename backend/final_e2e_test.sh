#!/bin/bash
BASE_URL="http://localhost:8080/api"

echo "=========================================="
echo "  Final E2E Test: All 3 New Endpoints"
echo "=========================================="

# Login to get token
LOGIN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"e2e@test.com","password":"Test123!@#"}')
TOKEN=$(echo $LOGIN | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
AUTH="Authorization: Bearer $TOKEN"

# Decode JWT payload to get user_id (base64)
PAYLOAD=$(echo "$TOKEN" | cut -d'.' -f2)
# Add padding if needed and decode
PADDED="${PAYLOAD}=="
USER_ID=$(echo "$PADDED" | base64 -d 2>/dev/null | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
echo "User ID: $USER_ID"

echo -e "\n[1] Creating project..."
PROJ=$(curl -s -X POST "$BASE_URL/projects" \
  -H "Content-Type: application/json" \
  -H "$AUTH" \
  -d '{"name":"Final E2E Test","description":"Testing"}')
PROJ_ID=$(echo $PROJ | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "Project ID: $PROJ_ID"

echo -e "\n[2] Creating stages..."
TODO=$(curl -s -X POST "$BASE_URL/projects/$PROJ_ID/stages" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"name":"To Do","position":1,"is_final":false}')
TODO_ID=$(echo $TODO | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

DONE=$(curl -s -X POST "$BASE_URL/projects/$PROJ_ID/stages" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"name":"Done","position":2,"is_final":true}')
DONE_ID=$(echo $DONE | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "Stages: ToDo=$TODO_ID, Done=$DONE_ID"

echo -e "\n[3] Creating tasks..."
T1=$(curl -s -X POST "$BASE_URL/projects/$PROJ_ID/stages/$TODO_ID/tasks" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"title":"Task 1","description":"P"}')
TASK1_ID=$(echo $T1 | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

T2=$(curl -s -X POST "$BASE_URL/projects/$PROJ_ID/stages/$DONE_ID/tasks" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"title":"Task 2","description":"D"}')
TASK2_ID=$(echo $T2 | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

T3=$(curl -s -X POST "$BASE_URL/projects/$PROJ_ID/stages/$TODO_ID/tasks" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"title":"Task 3","description":"O","deadline":"2020-01-01T00:00:00Z"}')
TASK3_ID=$(echo $T3 | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "Tasks created: $TASK1_ID, $TASK2_ID, $TASK3_ID"

echo -e "\n=========================================="
echo "  TESTING ENDPOINTS"
echo "=========================================="

# TEST 1: GET /api/projects/:id/members
echo -e "\n[TEST 1] GET /api/projects/:id/members"
MEMBERS=$(curl -s "$BASE_URL/projects/$PROJ_ID/members" -H "$AUTH")
if echo "$MEMBERS" | grep -q '"user_id"' && echo "$MEMBERS" | grep -q '"name"' && ! echo "$MEMBERS" | grep -q '"id":'; then
  echo "✅ PASS: Correct format"
else
  echo "❌ FAIL"
fi

# TEST 2: PUT /api/tasks/:id/assign
echo -e "\n[TEST 2] PUT /api/tasks/:id/assign"
ASSIGN=$(curl -s -X PUT "$BASE_URL/tasks/$TASK1_ID/assign" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d "{\"user_id\":\"$USER_ID\"}")
echo "Response: $ASSIGN"
if echo "$ASSIGN" | grep -q '"success":true'; then
  echo "✅ PASS: Assigned to owner"
else
  echo "❌ FAIL"
fi

# TEST 3: GET /api/projects/:id/stats
echo -e "\n[TEST 3] GET /api/projects/:id/stats"
STATS=$(curl -s "$BASE_URL/projects/$PROJ_ID/stats" -H "$AUTH")
echo "$STATS"
TOTAL=$(echo "$STATS" | grep -o '"total_tasks":[0-9]*' | cut -d':' -f2)
COMP=$(echo "$STATS" | grep -o '"completed_tasks":[0-9]*' | cut -d':' -f2)
OVER=$(echo "$STATS" | grep -o '"overdue_tasks":[0-9]*' | cut -d':' -f2)
echo "Results: Total=$TOTAL, Completed=$COMP, Overdue=$OVER"

[ "$TOTAL" = "3" ] && echo "✅ Total=3" || echo "❌ Total expected 3"
[ "$COMP" = "1" ] && echo "✅ Completed=1" || echo "❌ Completed expected 1"
[ "$OVER" = "1" ] && echo "✅ Overdue=1" || echo "❌ Overdue expected 1"

echo -e "\n=========================================="
echo "  ALL TESTS COMPLETE!"
echo "=========================================="
