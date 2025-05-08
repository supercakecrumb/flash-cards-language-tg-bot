# Deployment

This document describes the deployment setup for the Flash Cards Language Telegram Bot, including Docker configuration, GitHub Actions workflow, and environment variables.

## Dockerfile

The Dockerfile defines how to build the bot's container image.

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bot ./cmd/bot

# Final stage
FROM alpine:3.18

WORKDIR /app

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/bot /app/bot

# Create directory for migrations
RUN mkdir -p /app/internal/infrastructure/database/migrations

# Copy migrations
COPY --from=builder /app/internal/infrastructure/database/migrations /app/internal/infrastructure/database/migrations

# Set the timezone
ENV TZ=UTC

# Run as non-root user
RUN adduser -D appuser
USER appuser

# Command to run the application
CMD ["/app/bot"]
```

## Docker Compose

The docker-compose.yml file defines the services needed to run the bot locally or in a development environment.

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]
      interval: 10s
      timeout: 5s
      retries: 5

  bot:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      - TELEGRAM_TOKEN=${TELEGRAM_TOKEN}
      - ADMIN_IDS=${ADMIN_IDS}
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=${DB_NAME}
      - LOG_LEVEL=${LOG_LEVEL}
      - DICTIONARY_API=${DICTIONARY_API}
      - DICTIONARY_API_KEY=${DICTIONARY_API_KEY}
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped

volumes:
  postgres_data:
```

## Environment Variables

The .env.example file provides a template for the required environment variables.

```
# Telegram Bot Configuration
TELEGRAM_TOKEN=your_telegram_bot_token
ADMIN_IDS=123456789,987654321

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=flashcards
DB_PASSWORD=your_secure_password
DB_NAME=flashcards_db

# Dictionary API Configuration
DICTIONARY_API=freedictionary
DICTIONARY_API_KEY=your_api_key_if_needed

# Logging Configuration
LOG_LEVEL=info  # debug, info, warn, error
```

## GitHub Actions Workflow

The GitHub Actions workflow automates the build and deployment process.

```yaml
name: Build and Deploy

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install dependencies
        run: go mod download

      - name: Run tests
        run: go test -v ./...

  build-and-push:
    name: Build and Push
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Log in to GitHub Container Registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v4
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=ref,event=branch
            type=sha,format=short

      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: .
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy:
    name: Deploy
    needs: build-and-push
    if: startsWith(github.ref, 'refs/tags/v')
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to server
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.DEPLOY_HOST }}
          username: ${{ secrets.DEPLOY_USER }}
          key: ${{ secrets.DEPLOY_KEY }}
          script: |
            cd /opt/flash-cards-bot
            docker-compose pull
            docker-compose up -d
```

## Deployment Instructions

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/username/flash-cards-language-tg-bot.git
   cd flash-cards-language-tg-bot
   ```

2. Create a `.env` file based on `.env.example`:
   ```bash
   cp .env.example .env
   ```

3. Edit the `.env` file with your configuration.

4. Start the services using Docker Compose:
   ```bash
   docker-compose up -d
   ```

5. Check the logs:
   ```bash
   docker-compose logs -f bot
   ```

### Production Deployment

1. Set up a server with Docker and Docker Compose installed.

2. Create a deployment directory:
   ```bash
   mkdir -p /opt/flash-cards-bot
   ```

3. Create a `docker-compose.yml` file in the deployment directory:
   ```bash
   nano /opt/flash-cards-bot/docker-compose.yml
   ```

4. Add the following content to the file (adjust as needed):
   ```yaml
   version: '3.8'

   services:
     postgres:
       image: postgres:14-alpine
       environment:
         POSTGRES_USER: ${DB_USER}
         POSTGRES_PASSWORD: ${DB_PASSWORD}
         POSTGRES_DB: ${DB_NAME}
       volumes:
         - postgres_data:/var/lib/postgresql/data
       restart: unless-stopped
       healthcheck:
         test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]
         interval: 10s
         timeout: 5s
         retries: 5

     bot:
       image: ghcr.io/username/flash-cards-language-tg-bot:latest
       environment:
         - TELEGRAM_TOKEN=${TELEGRAM_TOKEN}
         - ADMIN_IDS=${ADMIN_IDS}
         - DB_HOST=postgres
         - DB_PORT=5432
         - DB_USER=${DB_USER}
         - DB_PASSWORD=${DB_PASSWORD}
         - DB_NAME=${DB_NAME}
         - LOG_LEVEL=${LOG_LEVEL}
         - DICTIONARY_API=${DICTIONARY_API}
         - DICTIONARY_API_KEY=${DICTIONARY_API_KEY}
       depends_on:
         postgres:
           condition: service_healthy
       restart: unless-stopped

   volumes:
     postgres_data:
   ```

5. Create a `.env` file in the deployment directory:
   ```bash
   nano /opt/flash-cards-bot/.env
   ```

6. Add your environment variables to the `.env` file.

7. Pull the images and start the services:
   ```bash
   cd /opt/flash-cards-bot
   docker-compose pull
   docker-compose up -d
   ```

8. Set up GitHub Actions secrets for automated deployment:
   - `DEPLOY_HOST`: The hostname or IP address of your server
   - `DEPLOY_USER`: The SSH username for your server
   - `DEPLOY_KEY`: The SSH private key for authentication

9. Create a new release tag in GitHub to trigger the deployment workflow.

## Backup and Restore

### Database Backup

To backup the PostgreSQL database:

```bash
docker-compose exec postgres pg_dump -U ${DB_USER} ${DB_NAME} > backup_$(date +%Y%m%d_%H%M%S).sql
```

### Database Restore

To restore the PostgreSQL database from a backup:

```bash
cat backup_file.sql | docker-compose exec -T postgres psql -U ${DB_USER} ${DB_NAME}
```

## Monitoring

For basic monitoring, you can use Docker's built-in health checks and logging:

```bash
# Check container status
docker-compose ps

# View logs
docker-compose logs -f

# Check container health
docker inspect --format "{{.State.Health.Status}}" flash-cards-bot_postgres_1
```

For more advanced monitoring, consider setting up Prometheus and Grafana.