BEGIN;
SELECT pg_advisory_xact_lock(750204);
INSERT INTO users (login, password_hash, created_at)
VALUES ('indices_demo_seed', 'disabled-demo-account', CURRENT_TIMESTAMP)
ON CONFLICT (login) DO NOTHING;

INSERT INTO heating_equipment (
    equipment_name, short_description, power_kw, cost_rub, image_name,
    video_name, status, creator_id, created_at, formed_at
)
SELECT
    '[Индексы] Прибор ' || g.n,
    'Демонстрационная карточка отопительного оборудования.',
    CASE WHEN g.n <= 240 THEN 1000 + g.n % 50 ELSE 1 + g.n % 250 END,
    5000 + g.n % 95000, '', '', 'опубликован', u.id,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM generate_series(1, 120000) AS g(n)
CROSS JOIN users AS u
WHERE u.login = 'indices_demo_seed'
  AND NOT EXISTS (
    SELECT 1 FROM heating_equipment AS e
    WHERE e.creator_id = u.id AND e.equipment_name = '[Индексы] Прибор ' || g.n
  );
COMMIT;
ANALYZE heating_equipment;
SELECT count(*) AS published_count FROM heating_equipment WHERE status = 'опубликован';
