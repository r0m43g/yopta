-- +goose Up
CREATE TABLE IF NOT EXISTS stations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    user_id INT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS tracks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    station_id INT NOT NULL,
    track_number VARCHAR(50) NOT NULL,
    type ENUM('through', 'dead_end') NOT NULL DEFAULT 'through',
    rule ENUM('fifo', 'filo') NOT NULL DEFAULT 'fifo',
    exceptions BOOLEAN NOT NULL DEFAULT FALSE,
    positions INT NOT NULL DEFAULT 1,
    length INT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (station_id) REFERENCES stations(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_station_track (station_id, track_number)
);

-- +goose Down
DROP TABLE IF EXISTS tracks;
DROP TABLE IF EXISTS stations;

