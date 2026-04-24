-- Seed para la base de datos de superadministrador (Postgres)
-- Crea tablas mínimas y registra datos demo para pruebas de integración.

BEGIN;

CREATE TABLE IF NOT EXISTS administradores (
  id BIGSERIAL PRIMARY KEY,
  email TEXT UNIQUE,
  name TEXT,
  role TEXT DEFAULT 'administrador',
  photo TEXT,
  fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  usuario_creador TEXT,
  estado TEXT DEFAULT 'activo',
  observaciones TEXT
);

CREATE TABLE IF NOT EXISTS tipos_de_licencia (
  id BIGSERIAL PRIMARY KEY,
  nombre TEXT NOT NULL UNIQUE,
  fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  usuario_creador TEXT,
  estado TEXT DEFAULT 'activo',
  observaciones TEXT
);

CREATE TABLE IF NOT EXISTS licencias (
  id BIGSERIAL PRIMARY KEY,
  empresa_id INTEGER,
  tipo_id INTEGER,
  nombre TEXT,
  descripcion TEXT,
  valor NUMERIC DEFAULT 0,
  duracion_dias INTEGER DEFAULT 0,
  modulos_habilitados TEXT,
  super_rol_habilitado INTEGER DEFAULT 0,
  fecha_inicio TIMESTAMP,
  fecha_fin TIMESTAMP,
  activo INTEGER DEFAULT 1,
  fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  usuario_creador TEXT,
  estado TEXT DEFAULT 'activo',
  observaciones TEXT
);

-- Datos demo
INSERT INTO administradores (email, name, role, usuario_creador, estado)
VALUES ('admin@demo.local', 'Admin Demo', 'super_administrador', 'system', 'activo')
ON CONFLICT (email) DO NOTHING;

INSERT INTO tipos_de_licencia (nombre)
VALUES ('Gratis')
ON CONFLICT (nombre) DO NOTHING;

INSERT INTO licencias (empresa_id, tipo_id, pais_codigo, nombre, descripcion, valor, duracion_dias, activo, usuario_creador)
VALUES (1, (SELECT id FROM tipos_de_licencia WHERE nombre = 'Gratis' LIMIT 1), 'CO', 'Licencia Demo', 'Licencia demo para integración', 0, 365, 1, 'system')
ON CONFLICT DO NOTHING;

-- Ajustar secuencias al max(id) existente
SELECT setval(pg_get_serial_sequence('administradores','id'), COALESCE((SELECT MAX(id) FROM administradores), 1), true);
SELECT setval(pg_get_serial_sequence('tipos_de_licencia','id'), COALESCE((SELECT MAX(id) FROM tipos_de_licencia), 1), true);
SELECT setval(pg_get_serial_sequence('licencias','id'), COALESCE((SELECT MAX(id) FROM licencias), 1), true);

COMMIT;
