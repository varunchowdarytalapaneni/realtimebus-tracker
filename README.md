# Real-Time Bus Tracker (`realtimebus-tracker`)

[![Go Version](https://img.shields.io/badge/go-1.22%2B-blue.svg)](https://go.dev/)
[![Flutter Version](https://img.shields.io/badge/flutter-3.x-02569B.svg)](https://flutter.dev/)
[![React Version](https://img.shields.io/badge/react-18.x-61DAFB.svg)](https://react.dev/)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

A high-throughput, distributed telemetry ingestion and real-time tracking engine built to stream, store, and visualize vehicle positions with low latency and offline resiliency.

---

## 🏗 System Architecture

```
                       +------------------------+
                       |   Flutter Mobile App   |
                       | (SQLite Offline Buffer)|
                       +-----------+------------+
                                   |
                             MQTT / HTTP POST
                                   v
                       +------------------------+
                       |    Go Telemetry API    |
                       +-----------+------------+
                                   |
                  +----------------+----------------+
                  |                                 |
                  v                                 v
        +------------------+              +-------------------+
        |  PostgreSQL /    |              |   Redis Pub/Sub   |
        |     PostGIS      |              +---------+---------+
        +------------------+                        |
                                              WebSockets
                                                    |
                                                    v
                                          +-------------------+
                                          |  React Dashboard  |
                                          |   (Leaflet.js)    |
                                          +-------------------+
```

### Core Architecture Components

* **Ingestion Layer (Go):** Handles concurrent incoming telemetry packets from drivers, validating and parsing spatial data at scale.
* **Storage Layer (PostgreSQL + PostGIS):** Persists geospatial historical routes, stop geometry, and route spatial queries.
* **Real-time Pipeline (Redis):** Handles sub-second location broadcasting via Pub/Sub to WebSockets subscribers.
* **Client Driver Application (Flutter):** Captures high-frequency GPS telemetry with local **SQLite buffering** to retain data during network dead-zones.
* **Admin Dashboard (React + Leaflet.js):** Consumes WebSocket events for real-time map updates, route monitoring, and vehicle status.

---

## 🚀 Quick Start (Docker Compose)

### Prerequisites

Ensure you have the following installed locally:
* [Docker Desktop](https://www.docker.com/products/docker-desktop/) (`>= 24.0`)
* [Docker Compose](https://docs.docker.com/compose/) (`>= v2.20`)
* [Go](https://go.dev/doc/install) (`>= 1.22`) *(for local dev)*
* [Node.js](https://nodejs.org/) (`>= 18.x`) *(for dashboard dev)*

---

## 🛠 Local Setup Instructions

### 1. Clone the Repository

```bash
git clone [https://github.com/varunchowdarytalapaneni/realtimebus-tracker.git](https://github.com/varunchowdarytalapaneni/realtimebus-tracker.git)
cd realtimebus-tracker
```

### 2. Environment Configuration

Copy the example environment file:

```bash
cp .env.example .env
```

Define configuration settings in `.env`:

```env
# Postgres / PostGIS
POSTGRES_DB=tracker_db
POSTGRES_USER=tracker_user
POSTGRES_PASSWORD=tracker_password
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Telemetry Server
SERVER_PORT=8080
```

### 3. Spin Up Infrastructure Services

Launch PostGIS, Redis, the Telemetry Backend, and the React Dashboard via Docker Compose:

```bash
docker-compose up -d --build
```

Verify service status:

```bash
docker-compose ps
```

---

## 📡 Telemetry Ingestion API Specs

### Ingest Location Packet

* **Endpoint:** `POST /api/v1/telemetry`
* **Content-Type:** `application/json`

**Request Payload:**

```json
{
  "bus_id": "BUS-104",
  "route_id": "ROUTE-A1",
  "latitude": 22.5726,
  "longitude": 88.3639,
  "speed": 34.5,
  "heading": 120.0,
  "timestamp": "2026-11-09T00:15:00Z"
}
```

**Response:**

```json
{
  "status": "acknowledged",
  "processed_at": "2026-11-09T00:15:00.012Z"
}
```

---

## 📂 Project Structure

```text
.
├── backend/                # Go telemetry ingestion service
│   ├── cmd/                # Entrypoints
│   ├── pkg/
│   │   ├── spatial/        # PostGIS query builder
│   │   ├── redis/          # Pub/Sub broadcaster
│   │   └── telemetry/      # HTTP/gRPC handlers
│   └── Dockerfile
├── dashboard/              # React + Leaflet Admin Dashboard
│   ├── src/
│   │   ├── components/     # Map components & layer controls
│   │   └── hooks/          # WebSocket hooks
│   └── Dockerfile
├── mobile/                 # Flutter driver telemetry app (SQLite buffer)
├── docker-compose.yml      # Orchestration file
└── README.md
```

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the project repository.
2. Create your feature branch (`git checkout -b feature/AmazingFeature`).
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4. Push to the branch (`git push origin feature/AmazingFeature`).
5. Open a Pull Request.

---

## 📜 License

Distributed under the **Apache License 2.0**. See `LICENSE` for more details.