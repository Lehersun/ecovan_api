-- Sample data fixtures for Eco Van API
-- This file contains realistic test data for development and testing

-- =========================================
-- Users (Authentication & Authorization)
-- =========================================
INSERT INTO users (id, email, password_hash, role, created_at, updated_at) VALUES
-- Admin user (password: admin123456)
('550e8400-e29b-41d4-a716-446655440001', 'admin@example.com', '$argon2id$v=19$m=65536,t=1,p=4$90e3854ee3be36f38422b08f29d88a67$8b33dea951971be17d8f8efdb1084b6b3a997d038e5b6c59c5789be3bfb30822', 'ADMIN', now(), now()),
-- Dispatcher user (password: admin123456)
('550e8400-e29b-41d4-a716-446655440002', 'dispatcher@example.com', '$argon2id$v=19$m=65536,t=1,p=4$90e3854ee3be36f38422b08f29d88a67$8b33dea951971be17d8f8efdb1084b6b3a997d038e5b6c59c5789be3bfb30822', 'DISPATCHER', now(), now()),
-- Driver user (password: admin123456)
('550e8400-e29b-41d4-a716-446655440003', 'driver@example.com', '$argon2id$v=19$m=65536,t=1,p=4$90e3854ee3be36f38422b08f29d88a67$8b33dea951971be17d8f8efdb1084b6b3a997d038e5b6c59c5789be3bfb30822', 'DRIVER', now(), now()),
-- Viewer user (password: admin123456)
('550e8400-e29b-41d4-a716-446655440004', 'viewer@example.com', '$argon2id$v=19$m=65536,t=1,p=4$90e3854ee3be36f38422b08f29d88a67$8b33dea951971be17d8f8efdb1084b6b3a997d038e5b6c59c5789be3bfb30822', 'VIEWER', now(), now())
ON CONFLICT (id) DO NOTHING;

-- =========================================
-- Clients
-- =========================================
INSERT INTO clients (id, name, tax_id, email, phone, notes, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440101', 'ООО "ЭкоСервис"', '7701234567', 'info@ecoservice.ru', '+7 (495) 123-45-67', 'Крупный клиент, регулярные заказы', now(), now()),
('550e8400-e29b-41d4-a716-446655440102', 'ИП Иванов А.С.', '7708765432', 'ivanov@mail.ru', '+7 (495) 987-65-43', 'Частный предприниматель', now(), now()),
('550e8400-e29b-41d4-a716-446655440103', 'ООО "Зеленый Мир"', '7701111111', 'green@world.ru', '+7 (495) 111-11-11', 'Экологическая компания', now(), now()),
('550e8400-e29b-41d4-a716-446655440104', 'АО "Чистый Город"', '7702222222', 'clean@city.ru', '+7 (495) 222-22-22', 'Муниципальное предприятие', now(), now())
ON CONFLICT (id) DO NOTHING;

-- =========================================
-- Warehouses
-- =========================================
INSERT INTO warehouses (id, name, address, notes, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440201', 'Центральный склад', 'г. Москва, ул. Складская, д. 1', 'Основной склад компании', now(), now()),
('550e8400-e29b-41d4-a716-446655440202', 'Северный склад', 'г. Москва, ул. Северная, д. 15', 'Склад для северных районов', now(), now()),
('550e8400-e29b-41d4-a716-446655440203', 'Южный склад', 'г. Москва, ул. Южная, д. 25', 'Склад для южных районов', now(), now())
ON CONFLICT (id) DO NOTHING;

-- =========================================
-- Client Objects
-- =========================================
INSERT INTO client_objects (id, client_id, name, address, geo_lat, geo_lng, notes, created_at, updated_at) VALUES
-- ООО "ЭкоСервис" objects
('550e8400-e29b-41d4-a716-446655440301', '550e8400-e29b-41d4-a716-446655440101', 'Офис на Тверской', 'г. Москва, ул. Тверская, д. 10', 55.7558, 37.6176, 'Главный офис', now(), now()),
('550e8400-e29b-41d4-a716-446655440302', '550e8400-e29b-41d4-a716-446655440101', 'Склад на МКАД', 'г. Москва, МКАД, 45 км', 55.7500, 37.6200, 'Складская зона', now(), now()),
-- ИП Иванов А.С. objects
('550e8400-e29b-41d4-a716-446655440303', '550e8400-e29b-41d4-a716-446655440102', 'Магазин на Арбате', 'г. Москва, ул. Арбат, д. 20', 55.7494, 37.5931, 'Торговая точка', now(), now()),
-- ООО "Зеленый Мир" objects
('550e8400-e29b-41d4-a716-446655440304', '550e8400-e29b-41d4-a716-446655440103', 'Экоцентр', 'г. Москва, ул. Экологическая, д. 5', 55.7600, 37.6000, 'Экологический центр', now(), now()),
-- АО "Чистый Город" objects
('550e8400-e29b-41d4-a716-446655440305', '550e8400-e29b-41d4-a716-446655440104', 'Администрация', 'г. Москва, ул. Административная, д. 1', 55.7700, 37.6100, 'Административное здание', now(), now())
ON CONFLICT (id) DO NOTHING;

-- =========================================
-- Drivers
-- =========================================
INSERT INTO drivers (id, full_name, phone, license_no, license_classes, photo, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440401', 'Петров Иван Сергеевич', '+7 (495) 111-22-33', '77АА123456', '["B", "C", "CE"]', NULL, now(), now()),
('550e8400-e29b-41d4-a716-446655440402', 'Сидоров Алексей Петрович', '+7 (495) 222-33-44', '77ББ654321', '["B", "C"]', NULL, now(), now()),
('550e8400-e29b-41d4-a716-446655440403', 'Козлов Дмитрий Иванович', '+7 (495) 333-44-55', '77ВВ789012', '["B", "C", "CE", "D"]', NULL, now(), now()),
('550e8400-e29b-41d4-a716-446655440404', 'Морозов Сергей Александрович', '+7 (495) 444-55-66', '77ГГ345678', '["B", "C"]', NULL, now(), now())
ON CONFLICT (id) DO NOTHING;

-- =========================================
-- Equipment
-- =========================================
INSERT INTO equipment (id, number, type, volume_l, condition, photo, client_object_id, warehouse_id, created_at, updated_at, deleted_at, transport_id) VALUES
-- Equipment at client objects
('550e8400-e29b-41d4-a716-446655440501', 'EQ-001', 'CONTAINER', 1000, 'GOOD', NULL, '550e8400-e29b-41d4-a716-446655440301', NULL, now(), now(), NULL, NULL),
('550e8400-e29b-41d4-a716-446655440502', 'EQ-002', 'BIN', 200, 'GOOD', NULL, '550e8400-e29b-41d4-a716-446655440302', NULL, now(), now(), NULL, NULL),
('550e8400-e29b-41d4-a716-446655440503', 'EQ-003', 'CONTAINER', 800, 'GOOD', NULL, '550e8400-e29b-41d4-a716-446655440303', NULL, now(), now(), NULL, NULL),
('550e8400-e29b-41d4-a716-446655440504', 'EQ-004', 'BIN', 150, 'DAMAGED', NULL, '550e8400-e29b-41d4-a716-446655440304', NULL, now(), now(), NULL, NULL),
('550e8400-e29b-41d4-a716-446655440505', 'EQ-005', 'CONTAINER', 1200, 'GOOD', NULL, '550e8400-e29b-41d4-a716-446655440305', NULL, now(), now(), NULL, NULL),

-- Equipment at warehouses
('550e8400-e29b-41d4-a716-446655440506', 'EQ-006', 'BIN', 100, 'GOOD', NULL, NULL, '550e8400-e29b-41d4-a716-446655440201', now(), now(), NULL, NULL),
('550e8400-e29b-41d4-a716-446655440507', 'EQ-007', 'CONTAINER', 600, 'GOOD', NULL, NULL, '550e8400-e29b-41d4-a716-446655440201', now(), now(), NULL, NULL),
('550e8400-e29b-41d4-a716-446655440508', 'EQ-008', 'BIN', 300, 'GOOD', NULL, NULL, '550e8400-e29b-41d4-a716-446655440202', now(), now(), NULL, NULL),
('550e8400-e29b-41d4-a716-446655440509', 'EQ-009', 'CONTAINER', 900, 'OUT_OF_SERVICE', NULL, NULL, '550e8400-e29b-41d4-a716-446655440203', now(), now(), NULL, NULL),

-- Equipment that will be assigned to transport (initially at warehouse)
('550e8400-e29b-41d4-a716-446655440510', 'EQ-010', 'CONTAINER', 1500, 'GOOD', NULL, NULL, '550e8400-e29b-41d4-a716-446655440201', now(), now(), NULL, NULL),
('550e8400-e29b-41d4-a716-446655440511', 'EQ-011', 'BIN', 250, 'GOOD', NULL, NULL, '550e8400-e29b-41d4-a716-446655440201', now(), now(), NULL, NULL)
ON CONFLICT (id) DO NOTHING;

-- =========================================
-- Transport
-- =========================================
INSERT INTO transport (id, plate_no, brand, model, capacity_l, current_driver_id, current_equipment_id, status, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440601', 'А123БВ77', 'ГАЗ', 'ГАЗель Next', 1500, '550e8400-e29b-41d4-a716-446655440401', '550e8400-e29b-41d4-a716-446655440510', 'IN_WORK', now(), now()),
('550e8400-e29b-41d4-a716-446655440602', 'В456ГД77', 'Ford', 'Transit', 1200, '550e8400-e29b-41d4-a716-446655440402', '550e8400-e29b-41d4-a716-446655440511', 'IN_WORK', now(), now()),
('550e8400-e29b-41d4-a716-446655440603', 'Е789ЖЗ77', 'Mercedes-Benz', 'Sprinter', 2000, '550e8400-e29b-41d4-a716-446655440403', NULL, 'IN_WORK', now(), now()),
('550e8400-e29b-41d4-a716-446655440604', 'И012КЛ77', 'ГАЗ', 'Соболь', 800, '550e8400-e29b-41d4-a716-446655440404', NULL, 'REPAIR', now(), now())
ON CONFLICT (id) DO NOTHING;

-- =========================================
-- Orders
-- =========================================
INSERT INTO orders (id, client_id, object_id, scheduled_date, scheduled_window_from, scheduled_window_to, status, transport_id, notes, created_by, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440701', '550e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440301', CURRENT_DATE + INTERVAL '1 day', '09:00:00', '12:00:00', 'SCHEDULED', '550e8400-e29b-41d4-a716-446655440601', 'Регулярный вывоз', '550e8400-e29b-41d4-a716-446655440001', now(), now()),
('550e8400-e29b-41d4-a716-446655440702', '550e8400-e29b-41d4-a716-446655440102', '550e8400-e29b-41d4-a716-446655440303', CURRENT_DATE + INTERVAL '2 days', '14:00:00', '17:00:00', 'DRAFT', NULL, 'Разовый заказ', '550e8400-e29b-41d4-a716-446655440002', now(), now()),
('550e8400-e29b-41d4-a716-446655440703', '550e8400-e29b-41d4-a716-446655440103', '550e8400-e29b-41d4-a716-446655440304', CURRENT_DATE, '10:00:00', '13:00:00', 'IN_PROGRESS', '550e8400-e29b-41d4-a716-446655440602', 'Экологический вывоз', '550e8400-e29b-41d4-a716-446655440001', now(), now()),
('550e8400-e29b-41d4-a716-446655440704', '550e8400-e29b-41d4-a716-446655440104', '550e8400-e29b-41d4-a716-446655440305', CURRENT_DATE - INTERVAL '1 day', '08:00:00', '11:00:00', 'COMPLETED', '550e8400-e29b-41d4-a716-446655440603', 'Муниципальный заказ', '550e8400-e29b-41d4-a716-446655440001', now(), now())
ON CONFLICT (id) DO NOTHING;

-- =========================================
-- Update equipment transport_id for assigned equipment
-- =========================================
UPDATE equipment SET transport_id = '550e8400-e29b-41d4-a716-446655440601', warehouse_id = NULL WHERE id = '550e8400-e29b-41d4-a716-446655440510';
UPDATE equipment SET transport_id = '550e8400-e29b-41d4-a716-446655440602', warehouse_id = NULL WHERE id = '550e8400-e29b-41d4-a716-446655440511';
