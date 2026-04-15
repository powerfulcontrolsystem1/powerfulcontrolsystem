PGPASSWORD=Ndog45CrAMJOpfa58JaqdFWj psql -U pcs_app -d pcs_superadministrador -h 127.0.0.1 -v ON_ERROR_STOP=1 <<'SQL'
UPDATE configuraciones
SET value = 'https://powerfulcontrolsystem.com/auth/google/callback',
    fecha_actualizacion = NOW()
WHERE config_key = 'google.redirect_url';

INSERT INTO configuraciones (config_key, value, encrypted, fecha_creacion, fecha_actualizacion)
SELECT 'google.redirect_url', 'https://powerfulcontrolsystem.com/auth/google/callback', 0, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM configuraciones WHERE config_key = 'google.redirect_url'
);

SELECT config_key, value FROM configuraciones WHERE config_key='google.redirect_url';
SQL
