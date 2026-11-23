\echo 'Без индекса'
BEGIN;
DROP INDEX IF EXISTS idx_heating_equipment_status_power;
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM heating_equipment
WHERE status = 'опубликован' AND power_kw >= 1000
ORDER BY power_kw ASC NULLS LAST, id ASC LIMIT 12 OFFSET 12;
ROLLBACK;

\echo 'С индексом'
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM heating_equipment
WHERE status = 'опубликован' AND power_kw >= 1000
ORDER BY power_kw ASC NULLS LAST, id ASC LIMIT 12 OFFSET 12;
