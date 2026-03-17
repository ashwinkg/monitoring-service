# Monitoring Demo — Go + Kafka + Postgres + Redis

A full monitoring and alerting stack using **Prometheus**, **Grafana**, and **Alertmanager**, built around a Go application that produces and consumes events through Kafka, persists them to Postgres, and caches in Redis.

---

## Prerequisites

Make sure you have the following installed before starting:

| Tool | Version | Check |
|------|---------|-------|
| Go | 1.22+ | `go version` |
| Docker | 24+ | `docker --version` |
| Docker Compose | 2.0+ | `docker compose version` |
| Git | any | `git --version` |

---

## Project Structure

```
monitoring-service/
├── app/                              # Go application
│   ├── cmd/
│   │   └── main.go                  # Entry point
│   ├── internal/
│   │   ├── config/config.go         # Env-based config loader
│   │   ├── handler/event.go         # HTTP handlers (Gin)
│   │   ├── metrics/metrics.go       # Prometheus custom metrics
│   │   ├── repository/
│   │   │   ├── postgres.go          # Postgres connection + Event repo
│   │   │   └── redis.go             # Redis client
│   │   └── service/
│   │       ├── event.go             # Core business logic
│   │       ├── producer.go          # Kafka producer
│   │       └── consumer.go          # Kafka consumer
│   ├── Dockerfile                   # Multi-stage build
│   ├── go.mod
│   └── go.sum
├── config/
│   ├── prometheus/
│   │   ├── prometheus.yml           # Scrape targets
│   │   └── alert-rules.yml          # Alert rule definitions
│   ├── alertmanager/
│   │   └── alertmanager.yml         # Slack routing config
│   └── grafana/
│       └── provisioning/
│           ├── datasources/         # Auto-configures Prometheus
│           └── dashboards/          # Dashboard provider config
├── docker-compose.yml               # Full stack definition
└── README.md
```

---

## Option A — Run with Docker Compose (recommended)

This is the easiest way. Docker builds the Go app and starts all services automatically — no local Go setup needed.

### Step 1 — Clone the repository

```bash
git clone https://github.com/ashwinkg/monitoring-service.git
cd monitoring-service
```

### Step 2 — Configure Slack alerts

Open `config/alertmanager/alertmanager.yml` and replace `SLACK_WEBHOOK_URL` with your actual Slack webhook URL:

```yaml
global:
  slack_api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
```

To get a Slack webhook URL:
1. Go to https://api.slack.com/apps and create a new app
2. Enable **Incoming Webhooks**
3. Click **Add New Webhook to Workspace** and select your channel
4. Copy the generated URL

Also create two channels in your Slack workspace:
- `#alerts-critical` — for critical alerts
- `#alerts-warning` — for warning alerts

> If you want to skip Slack for now, leave `SLACK_WEBHOOK_URL` as-is and monitor alerts directly at http://localhost:9093 instead.

### Step 3 — Start the stack

```bash
docker compose up -d
```

This starts 10 containers:

| Container | Purpose | Port |
|-----------|---------|------|
| `monitoring-demo-app` | Go application | 8080 |
| `postgres` | Database | 5432 |
| `postgres-exporter` | Postgres metrics | 9187 |
| `zookeeper` | Kafka dependency | 2181 |
| `kafka` | Message broker | 9092 |
| `kafka-exporter` | Kafka metrics | 9308 |
| `redis` | Cache | 6379 |
| `redis-exporter` | Redis metrics | 9121 |
| `prometheus` | Metrics store | 9090 |
| `alertmanager` | Alert routing | 9093 |
| `grafana` | Dashboards | 3000 |

### Step 4 — Verify all containers are healthy

```bash
docker compose ps
```

All containers should show `healthy` or `running`. Kafka takes the longest (~30 seconds).

### Step 5 — Check the app is running

```bash
curl http://localhost:8080/health
# Expected: {"status":"ok"}
```

---

## Option B — Run the Go app locally

Use this if you want to run and debug the Go app directly on your machine while keeping the infrastructure in Docker.

### Step 1 — Start only the infrastructure

```bash
docker compose up -d postgres kafka redis zookeeper \
  postgres-exporter kafka-exporter redis-exporter \
  prometheus alertmanager grafana
```

### Step 2 — Install Go dependencies

```bash
cd app
go mod tidy
```

### Step 3 — Set environment variables

```bash
export APP_PORT=8080
export POSTGRES_DSN="host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
export KAFKA_BROKER="localhost:9092"
export KAFKA_TOPIC="demo-events"
export KAFKA_GROUP="monitoring-demo-group"
export REDIS_ADDR="localhost:6379"
```

### Step 4 — Run the app

```bash
# Must run from inside app/ where go.mod lives
cd app
go run cmd/main.go
```

> **Important:** always `cd app` first. Running `go run app/cmd/main.go` from the
> project root will fail with a *"no required module provides package"* error
> because Go cannot find `go.mod` from the parent directory.

---

## Grafana Dashboard Setup

Once everything is running, open Grafana and import the community dashboards.

### Step 1 — Open Grafana

Go to http://localhost:3000 and log in:
- Username: `admin`
- Password: `admin`

Prometheus is already configured as a data source automatically via provisioning.

### Step 2 — Import dashboards

Go to **Dashboards → New → Import**, enter the dashboard ID, select **Prometheus** as the data source, and click **Import**.

| Service | Dashboard ID | URL |
|---------|-------------|-----|
| Go app (processes) | 6671 | https://grafana.com/grafana/dashboards/6671 |
| Kafka | 7589 | https://grafana.com/grafana/dashboards/7589 |
| Postgres | 9628 | https://grafana.com/grafana/dashboards/9628 |
| Redis | 763 | https://grafana.com/grafana/dashboards/763 |
| Alertmanager | 9578 | https://grafana.com/grafana/dashboards/9578 |

---

## Testing Alerts

### Send normal traffic

```bash
curl -X POST http://localhost:8080/api/trigger \
  -H "Content-Type: application/json" \
  -d '{"count": 50, "delayMs": 100}'
```

### Send a single event

```bash
curl -X POST http://localhost:8080/api/event
```

### Check event stats

```bash
curl http://localhost:8080/api/stats
```

### Trigger a Kafka consumer lag alert

Flood Kafka faster than the consumer can process:

```bash
curl -X POST http://localhost:8080/api/trigger \
  -H "Content-Type: application/json" \
  -d '{"count": 5000, "delayMs": 0}'
```

Watch the lag build in Prometheus at http://localhost:9090 by querying:
```
kafka_consumergroup_lag
```

### Trigger an InstanceDown alert

Stop an exporter to simulate a service going down:

```bash
# Stop the exporter
docker stop postgres-exporter

# Wait ~15 seconds — alert fires in Alertmanager and Slack
# Check alerts at: http://localhost:9093

# Bring it back
docker start postgres-exporter

# Wait ~15 seconds — resolved notification sent to Slack
```

Repeat with `redis-exporter` or `kafka-exporter` to test other services.

### View raw app metrics

```bash
# All metrics
curl http://localhost:8080/metrics

# Just the custom demo metrics
curl http://localhost:8080/metrics | grep demo_
```

---

## Useful URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| Go App | http://localhost:8080 | — |
| App Metrics | http://localhost:8080/metrics | — |
| App Health | http://localhost:8080/health | — |
| Prometheus | http://localhost:9090 | — |
| Alertmanager | http://localhost:9093 | — |
| Grafana | http://localhost:3000 | admin / admin |

---

## Stopping the Stack

```bash
# Stop all containers (keeps data volumes)
docker compose down

# Stop and delete all data volumes (fresh start)
docker compose down -v
```

---

## Troubleshooting

**`go run` fails with "no required module provides package"**

You are running from the wrong directory. Always run from inside `app/`:
```bash
cd app
go run cmd/main.go
```

**App container keeps restarting**

Kafka or Postgres may not be fully ready. Check logs:
```bash
docker compose logs app --follow
```
The app retries Postgres up to 10 times. If Kafka isn't ready, restart the app:
```bash
docker compose restart app
```

**No metrics appearing in Prometheus**

Check all scrape targets at http://localhost:9090/targets. Every target should show `State: UP`. If any are `DOWN`, check that the corresponding container is running:
```bash
docker compose ps
```

**Slack alerts not arriving**

Verify your webhook URL is correctly set, then restart Alertmanager:
```bash
docker compose restart alertmanager
```

Test the webhook directly:
```bash
curl -X POST -H 'Content-type: application/json' \
  --data '{"text":"Test from Alertmanager"}' \
  https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

**Grafana shows "No data"**

Make sure you selected **Prometheus** as the data source when importing each dashboard, not the default one.



docker compose down -v
docker compose build --no-cache app
docker compose up -d
docker logs monitoring-service