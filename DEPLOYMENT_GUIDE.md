# JO4 Dev: Zero-Downtime Red/Blue (Blue/Green) Deployment Guide

This guide covers how the zero-downtime Blue/Green (Red/Blue) deployment system operates, how to check which environment is active, and how to perform automated deployments or instant manual rollbacks.

---

## 1. Architecture Overview

```
                      [ Internet / Cloudflare Tunnel ]
                                     │
                                     ▼
                         [ Nginx Reverse Proxy ]
                          (Ports 80 & 6010)
                                     │
                    ┌────────────────┴────────────────┐
                    ▼                                 ▼
          [ web-blue (Container) ]          [ web-green (Container) ]
             Host Port: 5001                   Host Port: 5002
          ┌────────────────────────────────────────────────────────┐
          │     Shared Volume: SQLite WAL (/app/data/jo4dev.db)    │
          │     Shared Volume: Uploads (/app/static/uploads)       │
          └────────────────────────────────────────────────────────┘
```

- **Nginx Reverse Proxy**: Receives traffic on ports `80` and `6010` (Cloudflare Tunnel) and proxies requests to `jo4_backend` defined in `nginx/upstream.inc`.
- **`web-blue`**: Candidate / live Go Fiber container running on internal port 5000 (mapped to host 5001 for health checks).
- **`web-green`**: Candidate / live Go Fiber container running on internal port 5000 (mapped to host 5002 for health checks).
- **Shared Volumes**: Both containers share `/app/data` (SQLite in WAL mode) and `/app/static/uploads`. Only the active container serves traffic.

---

## 2. How to Check Which Environment is Active (Indicators)

### Indicator 1: Check Nginx Upstream (Instant)
```bash
cat nginx/upstream.inc
```
- Outputs `server web-blue:5000;` $\rightarrow$ **BLUE is LIVE**.
- Outputs `server web-green:5000;` $\rightarrow$ **GREEN is LIVE**.

### Indicator 2: Check Running Containers
```bash
docker ps
```
You will see `jo4-nginx` running alongside only **one** active web container:
- `jo4-web-blue` (port 5001)
- **OR** `jo4-web-green` (port 5002)

*(The idle container is stopped to save server RAM and CPU).*

### Indicator 3: Run the Deploy Script
Running `./deploy.sh` immediately prints the active and target environments:
```text
==========================================================
🚀 JO4 Dev Zero-Downtime Deployment
   Current Live Service       : BLUE (web-blue)
   Deploying New Release To   : GREEN (web-green on port 5002)
==========================================================
```

---

## 3. Automated Zero-Downtime Deployment

Whenever you push new code to GitHub, deploy to your server with two commands:

```bash
cd ~/my_website
git pull
./deploy.sh
```

### What `deploy.sh` does automatically:
1. **Detects** the live environment (`BLUE` or `GREEN`).
2. **Builds** the idle container with the latest code.
3. **Starts** the idle container in the background.
4. **Verifies Health**: Polls `http://127.0.0.1:<port>/health` until HTTP 200 is received.
   - *If health check fails*: Aborts deployment, dumps container logs, stops the failed container, and leaves the live environment untouched.
5. **Switches Upstream**: Atomically updates `nginx/upstream.inc` and reloads Nginx (`nginx -s reload`). **Zero dropped connections.**
6. **Drains & Shuts Down**: Waits 5 seconds for existing requests to complete, then stops the previous container.

---

## 4. Instant Manual Switching & Rollback (1 Second)

If you ever need to manually switch traffic or immediately roll back without rebuilding:

### Switch to BLUE:
```bash
docker compose up -d web-blue
echo "server web-blue:5000;" > nginx/upstream.inc
docker compose exec -T nginx nginx -s reload
docker compose stop web-green
```

### Switch to GREEN:
```bash
docker compose up -d web-green
echo "server web-green:5000;" > nginx/upstream.inc
docker compose exec -T nginx nginx -s reload
docker compose stop web-blue
```

*(Nginx reloads in ~5 milliseconds, so the switch is instantaneous).*

---

## 5. Troubleshooting & Logs

### View Nginx access & error logs:
```bash
docker compose logs -f nginx
```

### View active web container logs:
```bash
# If Blue is live:
docker compose logs -f web-blue

# If Green is live:
docker compose logs -f web-green
```

### Check Nginx configuration syntax:
```bash
docker compose exec -T nginx nginx -t
```
