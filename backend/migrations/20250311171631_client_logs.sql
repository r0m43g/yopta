-- +goose Up
CREATE TABLE IF NOT EXISTS client_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NULL,
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    client_timestamp VARCHAR(50) NOT NULL,
    url TEXT NULL,
    path VARCHAR(255) NULL,
    route VARCHAR(50) NULL,
    data MEDIUMTEXT NOT NULL,
    server_timestamp VARCHAR(50) NOT NULL,
    ip_address VARCHAR(50) NOT NULL,
    user_agent TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_client_logs_level (level),
    INDEX idx_client_logs_user_id (user_id),
    INDEX idx_client_logs_client_timestamp (client_timestamp),
    INDEX idx_client_logs_server_timestamp (server_timestamp),
    INDEX idx_client_logs_path (path),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- +goose Down
DROP TABLE IF EXISTS client_logs;

