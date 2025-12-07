-- +goose Up
CREATE TABLE activity_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    timestamp VARCHAR(50) NOT NULL,
    method VARCHAR(10) NOT NULL,
    path VARCHAR(255) NOT NULL,
    query TEXT,
    user_id VARCHAR(20),
    user_role TEXT,
    client_ip VARCHAR(50) NOT NULL,
    user_agent TEXT,
    request_id VARCHAR(100),
    request_body MEDIUMTEXT,
    response_body MEDIUMTEXT,
    status_code INT NOT NULL,
    duration BIGINT NOT NULL,
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_timestamp (timestamp),
    INDEX idx_path (path),
    INDEX idx_method (method),
    INDEX idx_status_code (status_code),
    INDEX idx_client_ip (client_ip)
);

-- +goose Down
DROP TABLE IF EXISTS activity_logs;

