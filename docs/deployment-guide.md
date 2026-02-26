# Production Environment Deployment Guide

## Prerequisites

Before deploying to production, ensure you have:

1. **Domain name** configured with DNS pointing to your server
2. **Docker and Docker Compose** installed on the server
3. **Resend API key** for email functionality (get from https://resend.com/api-keys)
4. **Minimum server specs**: 2 CPU cores, 4GB RAM, 20GB disk

## Quick Start

### 1. Setup Secrets

Run the interactive secrets setup script:

```bash
./scripts/setup-prod-secrets.sh
```

This will create 8 secret files in the `secrets/` directory:
- `db_password` - PostgreSQL database password
- `redis_password` - Redis password
- `jwt_secret` - JWT token signing secret
- `totp_key` - TOTP encryption key for 2FA
- `meili_master_key` - Meilisearch master key
- `rustfs_access_key` - S3 access key for RustFS
- `rustfs_secret_key` - S3 secret key for RustFS
- `resend_api_key` - Resend email service API key

### 2. Configure Environment Variables

Create a `.env.prod` file or export environment variables:

```bash
# Domain configuration
DOMAIN=your-domain.com

# Frontend URL (for email links)
FRONTEND_URL=https://your-domain.com

# Database name and user (optional, defaults shown)
DB_NAME=cms
DB_USER=cms_user

# Email configuration (optional)
RESEND_FROM_NAME=Sky Flux CMS
RESEND_FROM_EMAIL=noreply@your-domain.com
```

### 3. Configure Caddy

Edit `Caddyfile.production` and replace `{$DOMAIN}` with your actual domain:

```caddyfile
your-domain.com {
    # ... existing config
}
```

Or use the environment variable approach (current default):

```bash
export DOMAIN=your-domain.com
```

### 4. Start Production Containers

```bash
make docker-prod-up
```

Or manually:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml --env-file .env.prod up -d
```

### 5. Run Database Migrations

```bash
docker exec cms-api-prod ./cms migrate up
```

### 6. Access the Setup Wizard

Open your browser and navigate to:

```
https://your-domain.com
```

You'll be redirected to the setup wizard to create your admin account and first site.

## Architecture

### Container Communication

```
Internet (80/443)
    ↓
Caddy (reverse proxy)
    ↓
├── Web (Astro SSR:3000) → Frontend pages
└── API (Go:8080)        → Backend endpoints
    ↓
├── PostgreSQL (5432)  ← Internal network only
├── Redis (6379)       ← Internal network only
├── Meilisearch (7700) ← Internal network only
└── RustFS (9000)      ← Internal network only
```

### Security Features

1. **Docker Secrets**: Passwords stored as files, not environment variables
2. **HTTPS Only**: Caddy automatically obtains and renews Let's Encrypt certificates
3. **Internal Network**: Infrastructure ports (8080/3000/5432/6379/7700/9000) not exposed
4. **Security Headers**: Caddy adds HSTS, X-Frame-Options, CSP headers
5. **Resource Limits**: Each container has CPU/memory limits

### Volume Persistence

All data persists in Docker volumes:
- `postgres_data` - Database data
- `redis_data` - Redis persistence file
- `meili_data` - Meilisearch indexes
- `rustfs_data` - S3 object storage
- `caddy_data` - SSL certificates
- `caddy_config` - Caddy configuration

## Maintenance

### View Logs

```bash
# All services
docker compose -f docker-compose.yml -f docker-compose.prod.yml logs -f

# Specific service (production containers have -prod suffix)
docker logs -f cms-api-prod
docker logs -f cms-web-prod
docker logs -f cms-caddy-prod
```

### Restart Services

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml restart
```

### Update Deployment

```bash
# Pull latest code
git pull

# Rebuild and restart
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

### Database Backup

```bash
docker exec cms-postgres-prod pg_dump -U cms_user cms > backup_$(date +%Y%m%d).sql
```

### Database Restore

```bash
docker exec -i cms-postgres-prod psql -U cms_user cms < backup_20260226.sql
```

## Troubleshooting

### Setup Wizard Shows 503 Error

Check if all containers are healthy:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml ps
```

Wait for all services to show "healthy" status.

### Caddy SSL Certificate Errors

Check Caddy logs:

```bash
docker logs cms-caddy-prod
```

Ensure your domain's DNS correctly points to the server IP.

### Database Migration Errors

Enter the API container:

```bash
docker exec -it cms-api-prod sh
```

Manually run migrations:

```bash
./cms migrate status
./cms migrate up
```

### Media Upload Fails

Check RustFS container status:

```bash
docker logs cms-rustfs-prod
```

Verify the S3 credentials in `secrets/rustfs_access_key` and `secrets/rustfs_secret_key`.

## Performance Tuning

### PostgreSQL

Current limits: 1.5 CPU, 1.5GB RAM

Edit `docker-compose.prod.yml`:

```yaml
postgres:
  deploy:
    resources:
      limits:
        cpus: "2.0"
        memory: 2048M
```

### Redis

Current limits: 0.5 CPU, 512MB RAM

Max memory: 256MB with LRU eviction

### Meilisearch

Current limits: 0.5 CPU, 512MB RAM

Index size depends on content volume.

### API/Web

Current limits: 1.0 CPU (API), 0.5 CPU (Web)

Scale web containers for high traffic:

```yaml
web:
  deploy:
    replicas: 3
```

## Monitoring

### Health Checks

Each container has built-in health checks:

```bash
docker inspect cms-api-prod | grep -A 5 Health
```

### Metrics (Future)

Consider adding:
- Prometheus + Grafana for metrics
- Sentry for error tracking
- Log aggregation (ELK/Loki)

## Backup Strategy

### Daily Backup Script

Create `scripts/backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR="/backups/$(date +%Y%m%d)"
mkdir -p "$BACKUP_DIR"

# Database
docker exec cms-postgres-prod pg_dump -U cms_user cms > "$BACKUP_DIR/database.sql"

# Redis
docker exec cms-redis-prod redis-cli --rdb /data/dump.rdb
docker cp cms-redis-prod:/data/dump.rdb "$BACKUP_DIR/redis.rdb"

# Meilisearch indexes
docker cp cms-meilisearch-prod:/meili_data "$BACKUP_DIR/meili_data"

# Media files
docker cp cms-rustfs-prod:/data "$BACKUP_DIR/rustfs_data"
```

### Offsite Backup

Consider syncing to object storage:

```bash
rsync -av /backups/ s3://your-backup-bucket/cms/
```

## Security Checklist

- [ ] Secrets directory has 700 permissions
- [ ] All secret files have 600 permissions
- [ ] `.env.prod` not committed to git
- [ ] `secrets/` added to `.gitignore`
- [ ] Firewall allows only ports 80 and 443
- [ ] SSH key-based authentication enabled
- [ ] Regular security updates applied
- [ ] SSL certificates auto-renewing via Caddy
- [ ] Database backups automated
- [ ] Monitoring configured

## Further Reading

- [Caddy Documentation](https://caddyserver.com/docs/)
- [Docker Secrets Best Practices](https://docs.docker.com/engine/swarm/secrets/)
- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Meilisearch Guide](https://docs.meilisearch.com/)
