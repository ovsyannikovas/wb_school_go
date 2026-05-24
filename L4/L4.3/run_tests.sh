#!/bin/bash

BASE="http://localhost:8080"
PASS=0
FAIL=0

pass() { PASS=$((PASS + 1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  FAIL: $1"; }

# --- 1. Create Event ---
echo "=== Test 1: Create Event ==="
RESP=$(curl -s -X POST "$BASE/create_event" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user1","date":"2026-06-15","event":"Meeting with team","has_reminder":true,"reminder_at":"2026-06-15 10:50:00"}')
echo "$RESP"
EVENT_ID=$(echo "$RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "$EVENT_ID" ]; then
  pass "Event created: $EVENT_ID"
else
  fail "Event ID not found in response"
  exit 1
fi

# --- 2. Get Events For Day ---
echo -e "\n=== Test 2: Get Events For Day ==="
RESP=$(curl -s "$BASE/events_for_day?user_id=user1&date=2026-06-15")
echo "$RESP"
if echo "$RESP" | grep -q '"count"'; then
  pass "Events for day returned"
else
  fail "Events for day not returned"
fi

# --- 3. Get Events For Week ---
echo -e "\n=== Test 3: Get Events For Week ==="
RESP=$(curl -s "$BASE/events_for_week?user_id=user1&date=2026-06-15")
echo "$RESP"
if echo "$RESP" | grep -q '"count"'; then
  pass "Events for week returned"
else
  fail "Events for week not returned"
fi

# --- 4. Get Events For Month ---
echo -e "\n=== Test 4: Get Events For Month ==="
RESP=$(curl -s "$BASE/events_for_month?user_id=user1&date=2026-06-15")
echo "$RESP"
if echo "$RESP" | grep -q '"count"'; then
  pass "Events for month returned"
else
  fail "Events for month not returned"
fi

# --- 5. Update Event ---
echo -e "\n=== Test 5: Update Event ==="
RESP=$(curl -s -X PUT "$BASE/update_event" \
  -H "Content-Type: application/json" \
  -d '{"id":"'$EVENT_ID'","user_id":"user1","event":"Meeting rescheduled to 15:00"}')
echo "$RESP"
if echo "$RESP" | grep -q '"Event updated"; then
  pass "Event updated"
else
  fail "Event update failed"
fi

# --- 6. Create Without Reminder ---
echo -e "\n=== Test 6: Create Without Reminder ==="
RESP=$(curl -s -X POST "$BASE/create_event" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user1","date":"2026-07-01","event":"Code Review","has_reminder":false}')
echo "$RESP"
if echo "$RESP" | grep -q '"Event created"; then
  pass "Event without reminder created"
else
  fail "Event without reminder failed"
fi

# --- 7. Missing Required Fields ---
echo -e "\n=== Test 7: Missing Required Fields ==="
RESP=$(curl -s -X POST "$BASE/create_event" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user1"}')
echo "$RESP"
if echo "$RESP" | grep -q '"error"'; then
  pass "Returned error for missing fields"
else
  fail "No error for missing fields"
fi

# --- 8. Missing Params for Query ---
echo -e "\n=== Test 8: Missing Query Params ==="
RESP=$(curl -s "$BASE/events_for_day")
echo "$RESP"
if echo "$RESP" | grep -q '"error"'; then
  pass "Missing params returns error"
else
  fail "Missing params did not return error"
fi

# --- Summary ---
echo -e "\n=============================="
echo "PASS: $PASS | FAIL: $FAIL"
echo "=============================="
