#!/bin/bash
set -euo pipefail

BASE="${1:-http://127.0.0.1:8090}"
FAILS=0

check() {
    local desc="$1" url="$2" expected="$3"
    local status
    status="$(curl -s -o /dev/null -w "%{http_code}" "$url")"
    if echo "$status" | grep -q "$expected"; then
        echo "  PASS: $desc"
    else
        echo "  FAIL: $desc (expected $expected, got $status)"
        FAILS=$((FAILS + 1))
    fi
}

echo "=== simple-daily-termux smoke test ==="
echo "Target: $BASE"

check "health"              "$BASE/api/health"                   "200"
check "list todos"          "$BASE/api/todos"                     "200"
check "list countdown"      "$BASE/api/countdown"                 "200"
check "get summary"         "$BASE/api/summary"                   "200"
check "get pomodoro today"  "$BASE/api/pomodoro/today"            "200"
check "get calendar"        "$BASE/api/calendar?month=2026-06"    "200"
check "get ledger"          "$BASE/api/ledger?month=2026-06"      "200"
check "get diary month"     "$BASE/api/diary?month=2026-06"       "200"
check "serve html"          "$BASE/"                              "200"
check "serve css"           "$BASE/css/style.css"                 "200"
check "serve js"            "$BASE/js/main.js"                    "200"

# Functional test: todo-countdown linkage
echo "---"
echo "Functional: Todo → Countdown linkage"
TODO_ID=$(curl -sf -X POST "$BASE/api/todos" \
    -H 'Content-Type: application/json' \
    -d '{"title":"smoke test","deadline_at":"2026-12-31T23:59:59Z"}' | \
    python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")

if [ -n "$TODO_ID" ]; then
    COUNT=$(curl -sf "$BASE/api/countdown" | python3 -c "
import sys,json
d=json.load(sys.stdin)['data']
print(len([x for x in d if x.get('ref_id')=='$TODO_ID']))" 2>/dev/null || echo "0")
    if [ "$COUNT" = "1" ]; then
        echo "  PASS: countdown auto-created for todo"
        curl -sf -X DELETE "$BASE/api/todos/$TODO_ID" > /dev/null
    else
        echo "  FAIL: countdown not created (count=$COUNT)"
        FAILS=$((FAILS + 1))
    fi
else
    echo "  FAIL: could not create todo"
    FAILS=$((FAILS + 1))
fi

echo "---"
echo "Result: $FAILS failure(s)"
[ "$FAILS" -eq 0 ] && echo "ALL PASSED" || echo "SOME FAILED"
exit $FAILS
