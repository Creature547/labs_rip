-- Выполнить через Adminer -> SQL command ПОСЛЕ: go run ./cmd/migrate

DELETE FROM heating_likes;
DELETE FROM heating_calculations;
DELETE FROM users;

INSERT INTO users (id, login, password) VALUES
  (1, 'anya', 'anya'),
  (2, 'moderator', 'moderator'),
  (3, 'user2', 'user2');

INSERT INTO heating_calculations
(id, title, description, status, image_url, video_url, power_kw, cost_rub,
 area_m2, ceiling_height_m, floor_type, wall_material, desired_temperature_c, electricity_tariff_rub_kwh,
 created_at, formed_at, creator_id)
VALUES
(1, 'Цилиндрическая печь',
 'Современная цилиндрическая печь-камин для эффективного обогрева загородного дома.',
 'опубликован',
 'http://127.0.0.1:9000/heating/%D0%A6%D0%B8%D0%BB%D0%B8%D0%BD%D0%B4%D1%80%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B0%D1%8F%20%D0%BF%D0%B5%D1%87%D1%8C.png',
 'http://127.0.0.1:9000/heating/%D0%A6%D0%B8%D0%BB%D0%B8%D0%BD%D0%B4%D1%80%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B0%D1%8F%20%D0%BF%D0%B5%D1%87%D1%8C.mp4',
 7.0, 85000, 65, 2.7, '1 этаж', 'Газобетонный блок', 22, 5.50, NOW() - INTERVAL '8 days', NOW() - INTERVAL '7 days', 2),
(2, 'Бойлер косвенного нагрева',
 'Высокопроизводительный бойлер для стабильной подачи горячей воды и экономичной работы.',
 'опубликован',
 'http://127.0.0.1:9000/heating/%D0%91%D0%BE%D0%B9%D0%BB%D0%B5%D1%80.png',
 'http://127.0.0.1:9000/heating/%D0%B1%D0%BE%D0%B9%D0%BB%D0%B5%D1%80%D1%8B.mp4',
 24.0, 65000, 110, 2.9, 'Средний', 'Кирпич', 23, 6.10, NOW() - INTERVAL '6 days', NOW() - INTERVAL '5 days', 2),
(3, 'Электрическая печь',
 'Электрокамин с эффектом живого пламени и быстрым безопасным обогревом жилой комнаты.',
 'опубликован',
 'http://127.0.0.1:9000/heating/%D0%AD%D0%BB%D0%B5%D0%BA%D1%82%D1%80%D0%BE%D0%BA%D0%B0%D0%BC%D0%B8%D0%BD.png',
 'http://127.0.0.1:9000/heating/%D1%8D%D0%BB%D0%B5%D0%BA%D1%82%D1%80%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B0%D1%8F%20%D0%BF%D0%B5%D1%87%D1%8C%20%D1%81%20%D0%B8%D0%B7%D0%BE%D0%B1%D1%80%D0%B0%D0%B6%D0%B5%D0%BD%D0%B8%D0%B5%D0%BC%20%D0%BA%D0%B0%D0%BC%D0%B8%D0%BD%D0%B0.mp4',
 3.0, 35000, 28, 2.7, 'Последний', 'Дерево', 22, 5.50, NOW() - INTERVAL '4 days', NOW() - INTERVAL '3 days', 2),
(4, 'Теплогенератор',
 'Запись для проверки логического удаления: в интерфейсе отображаться не должна.',
 'удален',
 'http://127.0.0.1:9000/heating/%D0%A2%D0%B5%D0%BF%D0%BB%D0%BE%D0%B3%D0%B5%D0%BD%D0%B5%D1%80%D0%B0%D1%82%D0%BE%D1%80.png',
 'http://127.0.0.1:9000/heating/%D0%9F%D1%80%D0%BE%D0%BC%D1%8B%D1%88%D0%BB%D0%B5%D0%BD%D0%BD%D1%8B%D0%B9%20%D1%82%D0%B5%D0%BF%D0%BB%D0%BE%D0%B3%D0%B5%D0%BD%D0%B5%D1%80%D0%B0%D1%82%D0%BE%D1%80.mp4',
 200.0, 200000, 500, 4.0, '1 этаж', 'Монолитный бетон', 20, 7.20, NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day', 2),
(5, 'Новый отопительный прибор',
 '',
 'черновик',
 'http://127.0.0.1:9000/heating/%D0%AD%D0%BB%D0%B5%D0%BA%D1%82%D1%80%D0%BE%D0%BA%D0%B0%D0%BC%D0%B8%D0%BD.png',
 'http://127.0.0.1:9000/heating/%D1%8D%D0%BB%D0%B5%D0%BA%D1%82%D1%80%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B0%D1%8F%20%D0%BF%D0%B5%D1%87%D1%8C%20%D1%81%20%D0%B8%D0%B7%D0%BE%D0%B1%D1%80%D0%B0%D0%B6%D0%B5%D0%BD%D0%B8%D0%B5%D0%BC%20%D0%BA%D0%B0%D0%BC%D0%B8%D0%BD%D0%B0.mp4',
 0, 0, 0, 2.7, '1 этаж', 'Газобетонный блок', 22, 5.50, NOW(), NULL, 1);

INSERT INTO heating_likes (id, user_id, heating_calculation_id) VALUES
  (1, 1, 1),
  (2, 3, 1),
  (3, 1, 2),
  (4, 2, 2),
  (5, 3, 2),
  (6, 1, 3);

SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT MAX(id) FROM users));
SELECT setval(pg_get_serial_sequence('heating_calculations', 'id'), (SELECT MAX(id) FROM heating_calculations));
SELECT setval(pg_get_serial_sequence('heating_likes', 'id'), (SELECT MAX(id) FROM heating_likes));
