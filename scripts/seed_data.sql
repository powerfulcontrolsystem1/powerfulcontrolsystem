BEGIN TRANSACTION;

-- Seed mínimo para pruebas: empresa y un código de descuento
CREATE TABLE IF NOT EXISTS empresas (
  id INTEGER PRIMARY KEY,
  nombre TEXT,
  nit TEXT,
  fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
  fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
  usuario_creador TEXT,
  estado TEXT DEFAULT 'activo'
);
INSERT OR IGNORE INTO empresas (id, nombre, nit, usuario_creador) VALUES (1, 'Empresa Test', '900000000-1', 'seed');

CREATE TABLE IF NOT EXISTS codigos_de_descuento (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  empresa_id INTEGER NOT NULL,
  codigo TEXT NOT NULL,
  tipo_descuento TEXT DEFAULT 'valor_fijo',
  valor REAL DEFAULT 0,
  moneda TEXT DEFAULT 'COP',
  monto_minimo_compra REAL DEFAULT 0,
  segmento_cliente TEXT DEFAULT 'todos',
  canal_venta TEXT DEFAULT 'todos',
  fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
  fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
  usuario_creador TEXT,
  estado TEXT DEFAULT 'activo'
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_codigos_descuento_empresa_codigo ON codigos_de_descuento(empresa_id, codigo);

INSERT OR IGNORE INTO codigos_de_descuento (empresa_id, codigo, tipo_descuento, valor, moneda, usuario_creador) VALUES
  (1, 'SEED-TEST-0001', 'valor_fijo', 1000, 'COP', 'seed'),
  (1, 'SEED-TEST-0010', 'porcentaje', 10, 'COP', 'seed');

COMMIT;
