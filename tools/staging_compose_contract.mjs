#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const root = process.cwd();
const file = path.join(root, "deploy", "docker-compose.staging.yml");
const source = fs.readFileSync(file, "utf8");
const failures = [];

function requirePattern(pattern, label) {
  if (!pattern.test(source)) failures.push(label);
}

requirePattern(/^name:\s*powerful-control-system-staging\s*$/m, "El proyecto Compose de staging debe tener nombre propio.");
requirePattern(/^  migrate:\s*[\s\S]*?^    container_name:\s*pcs-staging-migrate\s*$/m, "El migrador de staging debe tener contenedor propio.");
requirePattern(/^  worker:\s*[\s\S]*?^    container_name:\s*pcs-staging-worker\s*$/m, "El worker de staging debe tener contenedor propio.");
requirePattern(/^  backend:\s*[\s\S]*?^    container_name:\s*pcs-staging-backend\s*$/m, "La API de staging debe tener contenedor propio.");
requirePattern(/^  migrate:[\s\S]*?^      PCS_ENV:\s*staging\s*$/m, "El migrador debe declararse como staging.");
requirePattern(/^  worker:[\s\S]*?^      PCS_ENV:\s*staging\s*$/m, "El worker debe declararse como staging.");
requirePattern(/^  worker:[\s\S]*?^      PCS_WORKER_ID:\s*pcs-worker-staging\s*$/m, "El worker de staging necesita identidad separada.");
requirePattern(/^  backend:[\s\S]*?^      PCS_ENV:\s*staging\s*$/m, "La API debe declararse como staging.");
requirePattern(/^      - pcs_staging_private_storage:\/app\/private_storage\s*$/m, "Staging debe usar almacenamiento privado aislado.");
requirePattern(/^  pcs_staging_private_storage:\s*$/m, "Falta el volumen privado de staging.");

if (/^\s*-\s+pcs_private_storage:\/app\/private_storage\s*$/m.test(source)) {
  failures.push("Staging no puede montar el almacenamiento privado de produccion.");
}

const report = {
  status: failures.length === 0 ? "ok" : "error",
  file: path.relative(root, file).replace(/\\/g, "/"),
  failures,
};
console.log(JSON.stringify(report, null, 2));
if (failures.length > 0) process.exitCode = 1;
