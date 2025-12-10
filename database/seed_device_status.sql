-- ============================================================
-- SEED DATA: DEVICE STATUS
-- Insert initial device status data for testing
-- ============================================================

USE smart_home_iot;

-- Insert initial lamp status (ON)
INSERT INTO lamp_status (status, mode, timestamp) VALUES 
('on', 'manual', NOW());

-- Insert initial door status (LOCKED)
INSERT INTO door_status (status, method, timestamp) VALUES 
('locked', 'remote', NOW());

-- Insert initial curtain status (CLOSED - position 0)
INSERT INTO curtain_status (position, mode, timestamp) VALUES 
(0, 'manual', NOW());

-- Verify inserts
SELECT 'Lamp Status:' as Info, status, mode, timestamp FROM lamp_status ORDER BY timestamp DESC LIMIT 1;
SELECT 'Door Status:' as Info, status, method, timestamp FROM door_status ORDER BY timestamp DESC LIMIT 1;
SELECT 'Curtain Status:' as Info, position, mode, timestamp FROM curtain_status ORDER BY timestamp DESC LIMIT 1;
