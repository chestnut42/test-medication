#!/usr/bin/env bash

set -e

base_url="http://localhost:8080"

for i in $(seq 1 10); do
  curl "$base_url/health" && break || sleep 1
done

echo "health was waited for"


function check {
  if [[ "$2" != "$3" ]]; then
    echo "❌ Test failed"
    echo "Expected $1: $2, got: $3"
    exit 1
  fi
}


set +e


# Create med for the first owner
response=$(curl -s -w "\n%{http_code}" -X PUT "$base_url/v1/medication/myid1" \
  -H "X-Med-Owner: owner1" \
  -d '{"name":"Paracetamol", "dosage":"500mg", "form":"tablEt"}')
body=$(echo "$response" | head -n1)
status=$(echo "$response" | tail -n1)

check "status" "200" "$status"
check "created name" "Paracetamol" $(echo "$body" | jq -r .name)
check "created dosage" "500mg" $(echo "$body" | jq -r .dosage)
check "created form" "tablet" $(echo "$body" | jq -r .form)


# Try to create the same med for the same owner
response=$(curl -s -w "\n%{http_code}" -X PUT "$base_url/v1/medication/myid1" \
  -H "X-Med-Owner: owner1" \
  -d '{"name":"Paracetamol", "dosage":"100mg", "form":"capsule"}')
body=$(echo "$response" | head -n1)
status=$(echo "$response" | tail -n1)

check "status" "409" "$status"


# Allows to create the same med for different owner
response=$(curl -s -w "\n%{http_code}" -X PUT "$base_url/v1/medication/myid1" \
  -H "X-Med-Owner: owner2" \
  -d '{"name":"Paracetamol", "dosage":"100mg", "form":"capsule"}')
body=$(echo "$response" | head -n1)
status=$(echo "$response" | tail -n1)

check "status" "200" "$status"
check "created name" "Paracetamol" $(echo "$body" | jq -r .name)
check "created dosage" "100mg" $(echo "$body" | jq -r .dosage)
check "created form" "capsule" $(echo "$body" | jq -r .form)


# Bad name
response=$(curl -s -w "\n%{http_code}" -X PUT "$base_url/v1/medication/myid3" \
  -H "X-Med-Owner: owner3" \
  -d '{"dosage":"100mg", "form":"capsule"}')
body=$(echo "$response" | head -n1)
status=$(echo "$response" | tail -n1)

check "status" "400" "$status"


# Bad dosage
response=$(curl -s -w "\n%{http_code}" -X PUT "$base_url/v1/medication/myid3" \
  -H "X-Med-Owner: owner3" \
  -d '{"name":"Paracetamol", "form":"capsule"}')
body=$(echo "$response" | head -n1)
status=$(echo "$response" | tail -n1)

check "status" "400" "$status"


# Bad form
response=$(curl -s -w "\n%{http_code}" -X PUT "$base_url/v1/medication/myid3" \
  -H "X-Med-Owner: owner3" \
  -d '{"name":"Paracetamol", "dosage":"100mg", "form":"supersphere"}')
body=$(echo "$response" | head -n1)
status=$(echo "$response" | tail -n1)

check "status" "400" "$status"


# Metrics
response=$(curl -s -X GET "$base_url/metrics")

if ! echo "$response" | grep "^go_memstats_" >/dev/null; then
  echo "❌ No Go metrics found"
fi
if ! echo "$response" | grep "^http_server_duration_milliseconds_bucket" >/dev/null; then
  echo "❌ No HTTP metrics found"
fi

echo "✅ All Good"
