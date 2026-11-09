CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS schools (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) CHECK (role IN ('driver', 'parent', 'principal')),
    school_id INT REFERENCES schools(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS routes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    school_id INT REFERENCES schools(id) ON DELETE CASCADE,
    start_time TIME,
    end_time TIME
);

CREATE TABLE IF NOT EXISTS buses (
    id SERIAL PRIMARY KEY,
    bus_number VARCHAR(50) UNIQUE NOT NULL,
    route_id INT REFERENCES routes(id) ON DELETE SET NULL,
    driver_id INT REFERENCES users(id) ON DELETE SET NULL,
    current_status VARCHAR(50) DEFAULT 'inactive'
);

CREATE TABLE IF NOT EXISTS bus_stops (
    id SERIAL PRIMARY KEY,
    route_id INT REFERENCES routes(id) ON DELETE CASCADE,
    stop_name VARCHAR(255) NOT NULL,
    location GEOMETRY(Point, 4326),
    sequence_order INT
);

CREATE TABLE IF NOT EXISTS live_telemetry_logs (
    id SERIAL PRIMARY KEY,
    bus_id INT REFERENCES buses(id) ON DELETE CASCADE,
    location GEOMETRY(Point, 4326),
    speed FLOAT,
    heading FLOAT,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bus_stops_location ON bus_stops USING GIST (location);
CREATE INDEX IF NOT EXISTS idx_telemetry_location ON live_telemetry_logs USING GIST (location);
