CREATE INDEX IF NOT EXISTS idx_heating_equipment_status_power
ON heating_equipment (status, power_kw, id);
ANALYZE heating_equipment;
