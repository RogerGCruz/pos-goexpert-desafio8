#!/bin/sh
set -e
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

# Bring up docker-compose (uses root docker-compose.yml)

docker-compose up -d --build

# Wait for the app to be healthy
echo "Waiting for app to become available..."
MAX_WAIT=60
for i in $(seq 1 $MAX_WAIT); do
  # try to get 200 OK from the auctions endpoint
  HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/auction?status=0" || true)
  if [ "$HTTP_STATUS" = "200" ]; then
    echo "App is reachable (HTTP 200)"
    break
  fi
  echo "Waiting... ($i/$MAX_WAIT) - status=$HTTP_STATUS"
  sleep 1
done

# final check
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/auction?status=0" || true)
if [ "$HTTP_STATUS" != "200" ]; then
  echo "App did not become healthy in time, final status=$HTTP_STATUS"
  echo "Displaying last 200 lines of app logs for diagnosis:"
  docker-compose logs --no-color --tail=200 app
  docker-compose down
  exit 2
fi

# Export a short auction duration for faster tests if not set
export AUCTION_DURATION=${AUCTION_DURATION:-2}

# Run godog-based tests (go test harness)
cd tests_bdd
# Ensure deps are downloaded
go mod tidy
# Run the tests
if go test -v; then
  echo "BDD tests passed"
  docker-compose down
  exit 0
else
  echo "BDD tests failed"
  echo "--- docker-compose logs (app + mongodb) ---"
  docker-compose logs --no-color --timestamps --tail=200
  docker-compose down
  exit 1
fi
