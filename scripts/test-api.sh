#!/bin/bash
# =====================================================
# Casbin Demo API Test Script (Bash/Linux)
# Usage: bash ./scripts/test-api.sh
# Requires: curl, jq
# =====================================================

BASE_URL="http://localhost:8080"
PASSED=0
FAILED=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

pass() {
    PASSED=$((PASSED + 1))
    echo -e "${GREEN}[PASS]${NC} $1"
}

fail() {
    FAILED=$((FAILED + 1))
    echo -e "${RED}[FAIL]${NC} $1 - $2"
}

api_call() {
    local method=$1
    local url=$2
    local data=$3
    local token=$4

    local headers="-H Content-Type:application/json"
    if [ -n "$token" ]; then
        headers="$headers -H Authorization:Bearer\ $token"
    fi

    if [ -n "$data" ]; then
        curl -s -X "$method" "$BASE_URL$url" $headers -d "$data"
    else
        curl -s -X "$method" "$BASE_URL$url" $headers
    fi
}

echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  Casbin Demo API Test Suite${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

echo -e "${YELLOW}[1] Health Check${NC}"
resp=$(api_call GET /health)
if echo "$resp" | jq -e '.status == "ok"' > /dev/null; then
    pass "Health Endpoint"
else
    fail "Health Endpoint" "$resp"
fi

echo ""
echo -e "${YELLOW}[2] Authentication${NC}"

login_resp=$(api_call POST /api/login '{"username":"admin","password":"123456"}')
admin_token=$(echo "$login_resp" | jq -r '.data.token // empty')
if [ -n "$admin_token" ] && [ "$admin_token" != "null" ]; then
    pass "Admin Login"
else
    fail "Admin Login" "$login_resp"
fi

user_login=$(api_call POST /api/login '{"username":"user","password":"123456"}')
user_token=$(echo "$user_login" | jq -r '.data.token // empty')
if [ -n "$user_token" ] && [ "$user_token" != "null" ]; then
    pass "User Login"
else
    fail "User Login" "$user_login"
fi

bad_login=$(api_call POST /api/login '{"username":"admin","password":"wrong"}')
if echo "$bad_login" | jq -e '.code != 0' > /dev/null; then
    pass "Login with wrong password (rejected)"
else
    fail "Login with wrong password" "Should be rejected"
fi

echo ""
echo -e "${YELLOW}[3] Authorized Endpoints${NC}"

if [ -n "$admin_token" ]; then
    resp=$(api_call GET /api/userinfo "" "$admin_token")
    if echo "$resp" | jq -e '.data.username == "admin"' > /dev/null; then
        pass "Get User Info"
    else
        fail "Get User Info" "$resp"
    fi

    resp=$(api_call GET /api/dashboard "" "$admin_token")
    if echo "$resp" | jq -e '.code == 0' > /dev/null; then
        pass "Get Dashboard (Admin)"
    else
        fail "Get Dashboard" "$resp"
    fi

    resp=$(api_call GET /api/users "" "$admin_token")
    if echo "$resp" | jq -e '.data.list' > /dev/null; then
        pass "Get User List (Admin)"
    else
        fail "Get User List" "$resp"
    fi

    resp=$(api_call GET /api/roles "" "$admin_token")
    if echo "$resp" | jq -e '.code == 0' > /dev/null; then
        pass "Get Role List (Admin)"
    else
        fail "Get Role List" "$resp"
    fi

    resp=$(api_call GET /api/permissions "" "$admin_token")
    if echo "$resp" | jq -e '.code == 0' > /dev/null; then
        pass "Get Permission List (Admin)"
    else
        fail "Get Permission List" "$resp"
    fi
fi

resp=$(api_call GET /api/userinfo)
code=$(echo "$resp" | jq -r '.code // 0')
if [ "$code" = "401" ] || [ "$(echo "$resp" | grep -c '401')" -gt 0 ]; then
    pass "Access without token (401)"
else
    fail "Access without token" "$resp"
fi

echo ""
echo -e "${YELLOW}[4] Permission Control${NC}"

if [ -n "$user_token" ]; then
    resp=$(api_call GET /api/dashboard "" "$user_token")
    if echo "$resp" | jq -e '.code == 0' > /dev/null; then
        pass "User can access dashboard"
    else
        fail "User dashboard" "$resp"
    fi

    resp=$(api_call GET /api/users "" "$user_token")
    if echo "$resp" | jq -e '.code == 403' > /dev/null; then
        pass "User cannot list users (403)"
    else
        fail "User permission denied" "$resp"
    fi
fi

echo ""
echo -e "${CYAN}========================================${NC}"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}  All $PASSED tests passed!${NC}"
else
    echo -e "${RED}  $PASSED passed, $FAILED failed${NC}"
fi
echo -e "${CYAN}========================================${NC}"
echo ""
