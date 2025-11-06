# Solution Steps

1. 1. Create and optimize a multi-stage Dockerfile: Use a Golang base for compiling, then an alpine minimal image for runtime. Set up non-root user, healthcheck, and best practices for cache and image shrinking. Use .dockerignore for faster builds.

2. 2. Write a docker-compose.yml that sets up 3 API containers (with rolling-update capabilities/healthcheck), an Nginx container configured as a load balancer for the APIs, plus PostgreSQL and Redis containers, on separate frontend/backend networks; set up persistence for DB and cache.

3. 3. Provide a robust nginx.conf in deploy/nginx.conf to route incoming HTTP requests to all API instances using round-robin or least connections.

4. 4. Refactor the Go server main.go to: 1) use environment variables for DB/Redis config; 2) implement /health and /ready endpoints that validate DB and Redis connectivity; 3) add graceful shutdown handler for SIGINT/SIGTERM; 4) ensure the server stops on request and properly closes DB/Redis clients.

5. 5. Make sure the Docker healthcheck in Dockerfile matches the Go /health endpoint logic, and propagate this in Compose health and deployment configuration.

6. 6. Document endpoints and architecture (in README, not included here), explaining how health/readiness checks, rolling updates, and scaling work; highlight minimal image size and network isolation.

7. 7. Test: Build and run with 'docker-compose up --build', confirm Nginx load-balances requests, API healthchecks pass, DB/Redis connections succeed, and graceful shutdown works (docker-compose stop/up/down).

