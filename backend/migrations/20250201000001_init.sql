-- +goose Up

CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role ENUM('admin', 'user', 'viewer') DEFAULT 'user',
    verified BOOLEAN DEFAULT FALSE,
    user_status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS profiles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    theme VARCHAR(255) DEFAULT 'dim',
    avatar VARCHAR(255) DEFAULT 'none',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS avatar_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    old_path VARCHAR(255),
    new_path VARCHAR(255),
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    token VARCHAR(255) NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS email_verifications (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  code VARCHAR(8) NOT NULL,
  expires_at DATETIME NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS system_settings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    setting_key VARCHAR(255) NOT NULL UNIQUE,
    setting_value TEXT,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS blacklisted_ips (
    id INT AUTO_INCREMENT PRIMARY KEY,
    ip_address VARCHAR(45) NOT NULL UNIQUE,
    reason TEXT,
    added_by INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (added_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Добавляем начальные настройки
INSERT INTO system_settings (setting_key, setting_value, description)
VALUES ('registration_enabled', 'true', 'Controls whether new user registration is allowed');

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS audit_avatar_changes
AFTER UPDATE ON profiles
FOR EACH ROW
BEGIN
    IF OLD.avatar != NEW.avatar THEN
        INSERT INTO avatar_history (user_id, old_path, new_path)
        VALUES (NEW.user_id, OLD.avatar, NEW.avatar);
    END IF;
END
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS audit_avatar_changes;
DROP TABLE IF EXISTS blacklisted_ips;
DROP TABLE IF EXISTS system_settings;
DROP TABLE email_verifications;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS avatar_history;
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS users;

