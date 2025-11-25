-- ============================================================
-- DATABASE: smart_home_iot
-- Smart Home IoT System - Complete Database Schema
-- ============================================================

CREATE DATABASE IF NOT EXISTS smart_home_iot;
USE smart_home_iot;

-- ============================================================
-- TABLE: USERS
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    user_id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role ENUM('admin','member','guest') DEFAULT 'member',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: FINGERPRINT_DATA
-- ============================================================
CREATE TABLE IF NOT EXISTS fingerprint_data (
    fingerprint_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    template_index INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_fp_user FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE,
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: PIN_CODES (1 PIN untuk semua user)
-- ============================================================
CREATE TABLE IF NOT EXISTS pin_codes (
    pin_id INT AUTO_INCREMENT PRIMARY KEY,
    pin_code VARCHAR(20) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pin_code (pin_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: ACCESS_LOGS
-- ============================================================
CREATE TABLE IF NOT EXISTS access_logs (
    access_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT,
    method ENUM('fingerprint','pin','remote','unknown') DEFAULT 'unknown',
    status ENUM('success','failed') DEFAULT 'failed',
    image_path TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_access_user FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE SET NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_method (method),
    INDEX idx_status (status),
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: CAMERA_CAPTURES
-- ============================================================
CREATE TABLE IF NOT EXISTS camera_captures (
    capture_id INT AUTO_INCREMENT PRIMARY KEY,
    image_path TEXT NOT NULL,
    detected_face VARCHAR(100),
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: SENSOR_GAS
-- ============================================================
CREATE TABLE IF NOT EXISTS sensor_gas (
    gas_id INT AUTO_INCREMENT PRIMARY KEY,
    ppm_value INT NOT NULL,
    status ENUM('normal','warning','danger') DEFAULT 'normal',
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: BUZZER_LOG
-- ============================================================
CREATE TABLE IF NOT EXISTS buzzer_log (
    buzzer_id INT AUTO_INCREMENT PRIMARY KEY,
    status ENUM('on','off') DEFAULT 'off',
    reason VARCHAR(200),
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: SENSOR_TEMPERATURE
-- ============================================================
CREATE TABLE IF NOT EXISTS sensor_temperature (
    temp_id INT AUTO_INCREMENT PRIMARY KEY,
    temperature FLOAT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: SENSOR_HUMIDITY
-- ============================================================
CREATE TABLE IF NOT EXISTS sensor_humidity (
    humid_id INT AUTO_INCREMENT PRIMARY KEY,
    humidity FLOAT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: SENSOR_LIGHT
-- ============================================================
CREATE TABLE IF NOT EXISTS sensor_light (
    light_id INT AUTO_INCREMENT PRIMARY KEY,
    lux INT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: DOOR_STATUS
-- ============================================================
CREATE TABLE IF NOT EXISTS door_status (
    door_id INT AUTO_INCREMENT PRIMARY KEY,
    status ENUM('locked','unlocked') DEFAULT 'locked',
    method ENUM('fingerprint','pin','remote','auto') DEFAULT 'remote',
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: LAMP_STATUS
-- ============================================================
CREATE TABLE IF NOT EXISTS lamp_status (
    lamp_id INT AUTO_INCREMENT PRIMARY KEY,
    status ENUM('on','off') DEFAULT 'off',
    mode ENUM('auto','manual') DEFAULT 'manual',
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: CURTAIN_STATUS
-- ============================================================
CREATE TABLE IF NOT EXISTS curtain_status (
    curtain_id INT AUTO_INCREMENT PRIMARY KEY,
    position INT NOT NULL DEFAULT 0,
    mode ENUM('auto','manual') DEFAULT 'manual',
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TABLE: NOTIFICATIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS notifications (
    notif_id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    message TEXT,
    type ENUM('gas','door','system','intruder') DEFAULT 'system',
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_type (type),
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- INSERT SAMPLE DATA
-- ============================================================

-- Insert default admin user (password: admin123)
INSERT INTO users (name, email, password, role) VALUES
('Admin', 'admin@smarthome.local', '$2a$10$8K1p/a0dL3.2E7HVy2Z3KeY5I.KH9.nZ8sH1J5xZ6K1xL7Y9Z3KeY', 'admin')
ON DUPLICATE KEY UPDATE name=name;

-- Insert sample member user (password: user123)
INSERT INTO users (name, email, password, role) VALUES
('John Doe', 'john@example.com', '$2a$10$8K1p/a0dL3.2E7HVy2Z3KeY5I.KH9.nZ8sH1J5xZ6K1xL7Y9Z3KeY', 'member')
ON DUPLICATE KEY UPDATE name=name;

-- Insert default PIN code
INSERT INTO pin_codes (pin_code) VALUES
('123456')
ON DUPLICATE KEY UPDATE pin_code=pin_code;

-- Insert initial notifications
INSERT INTO notifications (title, message, type) VALUES
('System Started', 'Smart Home IoT system has been initialized successfully', 'system'),
('Welcome', 'Welcome to your Smart Home IoT Dashboard', 'system')
ON DUPLICATE KEY UPDATE title=title;

-- ============================================================
-- VIEWS FOR EASY QUERYING
-- ============================================================

-- View: Latest sensor readings
CREATE OR REPLACE VIEW latest_sensor_data AS
SELECT 
    (SELECT temperature FROM sensor_temperature ORDER BY timestamp DESC LIMIT 1) as temperature,
    (SELECT humidity FROM sensor_humidity ORDER BY timestamp DESC LIMIT 1) as humidity,
    (SELECT lux FROM sensor_light ORDER BY timestamp DESC LIMIT 1) as light_lux,
    (SELECT ppm_value FROM sensor_gas ORDER BY timestamp DESC LIMIT 1) as gas_ppm,
    (SELECT status FROM sensor_gas ORDER BY timestamp DESC LIMIT 1) as gas_status;

-- View: Latest device status
CREATE OR REPLACE VIEW latest_device_status AS
SELECT 
    (SELECT status FROM door_status ORDER BY timestamp DESC LIMIT 1) as door_status,
    (SELECT status FROM lamp_status ORDER BY timestamp DESC LIMIT 1) as lamp_status,
    (SELECT position FROM curtain_status ORDER BY timestamp DESC LIMIT 1) as curtain_position;

-- View: Recent access logs with user info
CREATE OR REPLACE VIEW recent_access_logs AS
SELECT 
    al.access_id,
    al.user_id,
    u.name as user_name,
    u.email as user_email,
    al.method,
    al.status,
    al.timestamp
FROM access_logs al
LEFT JOIN users u ON al.user_id = u.user_id
ORDER BY al.timestamp DESC
LIMIT 100;

-- ============================================================
-- STORED PROCEDURES
-- ============================================================

-- Procedure: Get dashboard summary
DELIMITER //
CREATE PROCEDURE IF NOT EXISTS GetDashboardSummary()
BEGIN
    SELECT 
        (SELECT COUNT(*) FROM users) as total_users,
        (SELECT COUNT(*) FROM access_logs WHERE status='success' AND DATE(timestamp) = CURDATE()) as today_access_success,
        (SELECT COUNT(*) FROM access_logs WHERE status='failed' AND DATE(timestamp) = CURDATE()) as today_access_failed,
        (SELECT COUNT(*) FROM notifications WHERE DATE(timestamp) = CURDATE()) as today_notifications,
        (SELECT status FROM door_status ORDER BY timestamp DESC LIMIT 1) as current_door_status,
        (SELECT status FROM lamp_status ORDER BY timestamp DESC LIMIT 1) as current_lamp_status;
END //
DELIMITER ;

-- ============================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================

-- Composite indexes for common queries
ALTER TABLE access_logs ADD INDEX idx_user_status_time (user_id, status, timestamp);
ALTER TABLE sensor_gas ADD INDEX idx_status_time (status, timestamp);

-- ============================================================
-- SHOW SUMMARY
-- ============================================================

SELECT 'Database schema created successfully!' as Status;
SELECT TABLE_NAME, TABLE_ROWS 
FROM information_schema.TABLES 
WHERE TABLE_SCHEMA = 'smart_home_iot' 
ORDER BY TABLE_NAME;
