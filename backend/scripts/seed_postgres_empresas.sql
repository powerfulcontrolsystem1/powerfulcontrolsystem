-- Seed para la base de datos de empresas (Postgres)
-- Crea tablas mínimas y registra datos base para integracion local.

BEGIN;

CREATE TABLE IF NOT EXISTS empresas (
  id BIGSERIAL PRIMARY KEY,
  nombre TEXT,
  nit TEXT,
  fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  usuario_creador TEXT,
  estado TEXT DEFAULT 'activo',
  observaciones TEXT
);

CREATE TABLE IF NOT EXISTS clientes (
  id BIGSERIAL PRIMARY KEY,
  empresa_id INTEGER NOT NULL,
  tipo_documento TEXT NOT NULL DEFAULT 'NIT',
  numero_documento TEXT NOT NULL,
  digito_verificacion TEXT,
  tipo_persona TEXT DEFAULT 'juridica',
  nombre_razon_social TEXT NOT NULL,
  nombre_comercial TEXT,
  regimen_fiscal TEXT,
  responsabilidad_tributaria TEXT,
  email TEXT,
  telefono TEXT,
  direccion TEXT,
  pais TEXT DEFAULT 'CO',
  departamento TEXT,
  municipio TEXT,
  codigo_postal TEXT,
  fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  usuario_creador TEXT,
  estado TEXT DEFAULT 'activo',
  observaciones TEXT
);

-- Datos base
INSERT INTO empresas (id, nombre, nit, usuario_creador, estado)
VALUES (1, 'Empresa Base', '900100', 'admin@sistema.local', 'activo')
ON CONFLICT (id) DO NOTHING;

INSERT INTO clientes (empresa_id, tipo_documento, numero_documento, nombre_razon_social, email, estado)
VALUES (1, 'NIT', '900100', 'Cliente Base', 'cliente@sistema.local', 'activo')
ON CONFLICT DO NOTHING;

-- Ajustar secuencias
SELECT setval(pg_get_serial_sequence('empresas','id'), COALESCE((SELECT MAX(id) FROM empresas), 1), true);
SELECT setval(pg_get_serial_sequence('clientes','id'), COALESCE((SELECT MAX(id) FROM clientes), 1), true);

COMMIT;
