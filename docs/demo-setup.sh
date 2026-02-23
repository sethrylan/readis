#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Build and install readis
cd "$REPO_ROOT"
go build -o readis ./cmd/readis
sudo install readis /usr/local/bin/readis

# Start Redis
docker run -d --name redis -p 6379:6379 redis:latest
until docker exec redis redis-cli PING >/dev/null 2>&1; do sleep 1; done

# Populate demo data
docker exec -i redis redis-cli <<'EOF'
SET mood:now "☻ Fabulous" EX 333
SET music:now "♫ Bowie" EX 5555
ZADD music:next 0.1111111 "Radiohead" 2.718281828459045 "Fleetwood Mac" 3.141592653589793238462643383279502884197 "Pink Floyd"
HMSET book:now title "The Hitchhiker's Guide to the Galaxy" author "Douglas Adams" genre "☄ Sci-Fi" quote "\"Don't Panic!\""
HMSET book:next title "If on a winter's night a traveler" author "Italo Calvino" genre "✧ Postmodern Fiction" quote "\"You are about to begin reading Italo Calvino's new novel, If on a winter's night a traveler…\""
EXPIRE music:next 2222222
EXPIRE book:now 11111
EOF
