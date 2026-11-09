-- =========================================
-- 1. USERS
-- =========================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =========================================
-- 2. SERVERS
-- =========================================

CREATE TABLE servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,
    hostname VARCHAR(255),
    os VARCHAR(100),

    status VARCHAR(20) NOT NULL DEFAULT 'unknown'
        CHECK (status IN ('online', 'offline', 'unknown')),

    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_servers_user_id
ON servers(user_id);

-- =========================================
-- 3. METRICS
-- =========================================

CREATE TABLE metrics (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    server_id UUID NOT NULL
        REFERENCES servers(id)
        ON DELETE CASCADE,

    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    cpu_percent NUMERIC(5,2)
        CHECK (cpu_percent BETWEEN 0 AND 100),

    memory_used_bytes BIGINT
        CHECK (memory_used_bytes >= 0),

    memory_total_bytes BIGINT
        CHECK (memory_total_bytes > 0),

    disk_used_bytes BIGINT
        CHECK (disk_used_bytes >= 0),

    disk_total_bytes BIGINT
        CHECK (disk_total_bytes > 0),

    network_rx_bytes_total BIGINT
        CHECK (network_rx_bytes_total >= 0),

    network_tx_bytes_total BIGINT
        CHECK (network_tx_bytes_total >= 0)
);

CREATE INDEX idx_metrics_server_recorded
ON metrics(server_id, recorded_at DESC);

-- =========================================
-- INITIAL USER
-- =========================================

INSERT INTO users (display_name)
VALUES ('Local User');