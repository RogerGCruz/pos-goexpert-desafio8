#!/bin/sh
set -e
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

# Bring up docker-compose (uses root docker-compose.yml)
docker-compose up -d --build

# Wait for the app to be healthy
echo "Waiting for app to become available..."
for i in 1 2 3 4 5 6 7 8 9 10; do
  if curl -sSf http://localhost:8080/auction?status=0 >/dev/null 2>&1; then
    echo "App is reachable"
    break
  fi
  echo "Waiting... ($i)"
  sleep 1
done

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
  docker-compose logs --no-color --timestamps
  docker-compose down
  exit 1
fi
