#!/bin/bash
BASE_URL="http://localhost:8080/api"

echo "=========================================="
echo "  Critical Edge Case Testing"
echo "=========================================="

LOGIN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"e2e@test.com","password":"Test123!@#"}')
TOKEN=$(echo $LOGIN | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
AUTH="Authorization: Bearer $TOKEN"

# 1. Delete final stage → what happens to stats?
echo -e "\n[1] Delete final stage with tasks..."
PROJ=$(curl -s -X POST "$BASE_URL/projects" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"name":"Delete Test","description":"Test"}')
PROJ_ID=$(echo $PROJ | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

FINAL=$(curl -s -X POST "$BASE_URL/projects/$PROJ_ID/stages" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"name":"Final","position":1,"is_final":true}')
FINAL_ID=$(echo $FINAL | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

# Create task in final stage
TASK=$(curl -s -X POST "$BASE_URL/projects/$PROJ_ID/stages/$FINAL_ID/tasks" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"title":"Task in Final","description":"Test"}')
TASK_ID=$(echo $TASK | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "Created task $TASK_ID in final stage $FINAL_ID"

STATS_BEFORE=$(curl -s "$BASE_URL/projects/$PROJ_ID/stats" -H "$AUTH")
echo "Stats before delete: $STATS_BEFORE"

# Delete the final stage (FK CASCADE should delete task)
DELETE=$(curl -s -X DELETE "$BASE_URL/projects/$PROJ_ID/stages/$FINAL_ID" -H "$AUTH")
echo "Delete stage response: $DELETE"

STATS_AFTER=$(curl -s "$BASE_URL/projects/$PROJ_ID/stats" -H "$AUTH")
echo "Stats after delete: $STATS_AFTER"

# 2. Create task without stage (if possible)
echo -e "\n[2] Try creating task without stage..."
BAD_TASK=$(curl -s -X POST "$BASE_URL/tasks" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"title":"No Stage","description":"Test"}')
echo "Create without stage: $BAD_TASK"

# 3. Check if assigned_to cleanup needed
echo -e "\n[3] Check assigned_to consistency..."
echo "Database tasks with assigned_to:"
sqlite3 taskify.db "SELECT id, assigned_to FROM tasks WHERE assigned_to IS NOT NULL LIMIT 5" 2>/dev/null

echo -e "\n=========================================="
echo "  Critical Test Summary"
echo "=========================================="
