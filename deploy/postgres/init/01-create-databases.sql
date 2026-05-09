SELECT 'CREATE DATABASE pcs_superadministrador'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'pcs_superadministrador')\gexec

SELECT 'CREATE DATABASE pcs_empresas'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'pcs_empresas')\gexec
