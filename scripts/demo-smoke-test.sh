#!/bin/bash
set -e

API_URL="${SURGE_API_URL:-https://diffsurge-k5wd.onrender.com}"
API_KEY="${SURGE_API_KEY}"
PROJECT_ID="${SURGE_PROJECT_ID}"

if [ -z "$API_KEY" ]; then
  echo "Error: SURGE_API_KEY not set"
  echo "  export SURGE_API_KEY=diffsurge_live_..."
  exit 1
fi

if [ -z "$PROJECT_ID" ]; then
  echo "Error: SURGE_PROJECT_ID not set"
  echo "  export SURGE_PROJECT_ID=your-project-uuid"
  exit 1
fi

echo "=== DiffSurge Demo Smoke Test ==="
echo "  API URL:    $API_URL"
echo "  Project ID: $PROJECT_ID"
echo ""

PASS=0
FAIL=0

check() {
  local name="$1"
  shift
  printf "%-30s" "$name"
  if "$@" > /dev/null 2>&1; then
    echo "✓"
    PASS=$((PASS + 1))
  else
    echo "✗ FAIL"
    FAIL=$((FAIL + 1))
  fi
}

# 1. Health
check "1. Health check" \
  curl -sf "$API_URL/api/v1/health"

# 2. Ready
check "2. Ready check" \
  curl -sf "$API_URL/api/v1/ready"

# 3. Auth (Organizations)
check "3. API Auth" \
  curl -sf -H "X-API-Key: $API_KEY" "$API_URL/api/v1/organizations"

# 4. Projects list
check "4. Projects list" \
  curl -sf -H "X-API-Key: $API_KEY" "$API_URL/api/v1/projects"

# 5. Project detail
check "5. Project detail" \
  curl -sf -H "X-API-Key: $API_KEY" "$API_URL/api/v1/projects/$PROJECT_ID"

# 6. Traffic list
check "6. Traffic list" \
  curl -sf -H "X-API-Key: $API_KEY" "$API_URL/api/v1/projects/$PROJECT_ID/traffic"

# 7. Traffic POST (new endpoint)
check "7. Traffic POST" \
  curl -sf -X POST -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" \
  "$API_URL/api/v1/projects/$PROJECT_ID/traffic" \
  -d '{"method":"GET","path":"/smoke-test","status_code":200,"latency_ms":1}'

# 8. Traffic stats
check "8. Traffic stats" \
  curl -sf -H "X-API-Key: $API_KEY" "$API_URL/api/v1/projects/$PROJECT_ID/traffic/stats"

# 9. Schemas list
check "9. Schemas list" \
  curl -sf -H "X-API-Key: $API_KEY" "$API_URL/api/v1/projects/$PROJECT_ID/schemas"

# 10. Schema POST
check "10. Schema POST" \
  curl -sf -X POST -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" \
  "$API_URL/api/v1/projects/$PROJECT_ID/schemas" \
  -d "{\"version\":\"smoke-test-$(date +%s)\",\"schema_type\":\"openapi\",\"schema_content\":{\"openapi\":\"3.0.0\",\"info\":{\"title\":\"Smoke Test\",\"version\":\"1.0\"}}}"

# 11. Replays list
check "11. Replays list" \
  curl -sf -H "X-API-Key: $API_KEY" "$API_URL/api/v1/projects/$PROJECT_ID/replays"

# 12. Environments list
check "12. Environments list" \
  curl -sf -H "X-API-Key: $API_KEY" "$API_URL/api/v1/projects/$PROJECT_ID/environments"

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
