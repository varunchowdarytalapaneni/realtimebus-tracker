# Realtime Bus Tracker Monorepo

[![Build Status](https://img.shields.io/badge/build-passing-lightgrey)](#)
[![Coverage](https://img.shields.io/badge/coverage-00%25-lightgrey)](#)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

Production-grade monorepo scaffold for a real-time transit tracking platform.

## Repository Layout

```text
.
├── apps/
│   └── dashboard/                 # React dashboard frontend
├── services/
│   └── telemetry-ingestion/       # Go telemetry ingestion service
├── infra/
│   └── docker/                    # Docker Compose stack (PostGIS + Redis)
├── docs/
│   └── architecture/              # Architecture module references
└── CONTRIBUTING.md                # Contribution standards
```

## Architecture Breakdown

### 1) Telemetry Ingestion Service (Go)

- Entry point under `cmd/ingester`
- Internal packages organized by responsibility:
  - `internal/config` for runtime config
  - `internal/ingestion` for domain ingestion flows
  - `internal/transport/http` for API transport
- Shared contracts in `pkg/models`

### 2) Dashboard (React)

- App shell in `src/app`
- Feature and reusable UI boundaries under `src/features` and `src/components`
- API integrations in `src/services`
- Global styles in `src/styles`

### 3) Infrastructure (Docker Compose)

- `infra/docker/docker-compose.yml` provisions:
  - PostgreSQL with PostGIS extensions
  - Redis with append-only persistence
- `infra/docker/postgres-init` contains initialization SQL scripts

## Getting Started

### Prerequisites

- Go 1.23+
- Node.js 20+
- Docker + Docker Compose

### Local Infrastructure

```bash
cd infra/docker
docker compose up -d
```

### Next Steps

- Implement ingestion endpoints and validation pipeline in Go service.
- Integrate dashboard map visualizations and live update streams.
- Add CI workflows, linting pipelines, and production deployment manifests.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution and review standards.
