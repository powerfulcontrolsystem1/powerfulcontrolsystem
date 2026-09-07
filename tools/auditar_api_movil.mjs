#!/usr/bin/env node
// Inventory-only audit for the mobile migration. It never connects to a DB or
// changes runtime code; it maps registered HTTP routes into documentation.
import fs from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const backend = path.join(root, 'backend');
const out = path.join(root, 'documentos', 'api', 'inventario_api_movil.md');

function walk(dir) {
  const entries = fs.readdirSync(dir, {withFileTypes:true});
  return entries.flatMap((entry) => entry.isDirectory() ? walk(path.join(dir, entry.name)) : [path.join(dir, entry.name)]);
}
function family(route) {
  if (route.startsWith('/api/v1/')) return 'movil_v1';
  if (route.startsWith('/api/public/')) return 'publica_webhook';
  if (route.startsWith('/super/api/')) return 'super_administrador';
  if (route.startsWith('/api/empresa/')) return 'empresa_web_legacy';
  if (route.startsWith('/api/')) return 'api_general';
  return 'otro';
}
const routes = [];
for (const file of walk(backend).filter((name) => name.endsWith('.go'))) {
  const relative = path.relative(root, file).replaceAll('\\', '/');
  const text = fs.readFileSync(file, 'utf8');
  for (const match of text.matchAll(/http\.HandleFunc\("([^"]+)"/g)) routes.push({route:match[1], file:relative, family:family(match[1])});
}
routes.sort((a,b) => a.route.localeCompare(b.route) || a.file.localeCompare(b.file));
const grouped = Object.groupBy(routes, (row) => row.family);
let md = '# Inventario de APIs para migracion movil\n\n';
md += 'Generado por `tools/auditar_api_movil.mjs`. Es un inventario de rutas registradas; no sustituye las pruebas de autorizacion, tenant y negocio.\n\n';
md += `Total de rutas detectadas: **${routes.length}**.\n\n`;
for (const name of Object.keys(grouped).sort()) {
  md += `## ${name}\n\n| Ruta | Registro |\n|---|---|\n`;
  for (const row of grouped[name]) md += `| \`${row.route}\` | \`${row.file}\` |\n`;
  md += '\n';
}
fs.writeFileSync(out, `${md.trimEnd()}\n`, 'utf8');
console.log(`Inventario actualizado: ${path.relative(root, out)} (${routes.length} rutas)`);
