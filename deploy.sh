#!/bin/bash
set -euo pipefail

DOMAIN="${1:-cms.example.com}"
VERSION="${2:-latest}"

echo "==> Pulling images..."
docker pull ghcr.io/sky-flux/cms-backend:${VERSION}
docker pull ghcr.io/sky-flux/cms-frontend:${VERSION}

echo "==> Stopping old services..."
docker compose -f docker-compose.prod.yml down

echo "==> Starting new services..."
DOMAIN=${DOMAIN} docker compose -f docker-compose.prod.yml up -d

echo "==> Waiting for health checks..."
sleep 15
if curl -sf https://${DOMAIN}/health > /dev/null; then
    echo "==> Deployment successful!"
else
    echo "==> Health check failed, viewing logs:"
    docker compose -f docker-compose.prod.yml logs --tail=50
    exit 1
fi
