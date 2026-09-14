-- 1) Сначала выполнить в терминале: go run ./cmd/migrate
--    Это добавит новые поля помещения/тарифа в уже существующую БД.

-- 2) Ты случайно удалила карточку "Бойлер косвенного нагрева" (id = 2).
--    Возвращаем ее в опубликованные:
UPDATE heating_calculations
SET status = 'опубликован'
WHERE id = 2;

-- 3) Для уже существующих карточек зададим нормальные значения новых полей,
--    чтобы они сразу красиво показывались в ленте.
UPDATE heating_calculations SET
    area_m2 = 65,
    ceiling_height_m = 2.7,
    floor_type = '1 этаж',
    wall_material = 'Газобетонный блок',
    desired_temperature_c = 22,
    electricity_tariff_rub_kwh = 5.50
WHERE id = 1;

UPDATE heating_calculations SET
    area_m2 = 110,
    ceiling_height_m = 2.9,
    floor_type = 'Средний',
    wall_material = 'Кирпич',
    desired_temperature_c = 23,
    electricity_tariff_rub_kwh = 6.10
WHERE id = 2;

UPDATE heating_calculations SET
    area_m2 = 28,
    ceiling_height_m = 2.7,
    floor_type = 'Последний',
    wall_material = 'Дерево',
    desired_temperature_c = 22,
    electricity_tariff_rub_kwh = 5.50
WHERE id = 3;
