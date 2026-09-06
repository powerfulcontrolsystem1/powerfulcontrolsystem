import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { execFileSync, spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { buildCatalog } from './docs_catalog.mjs';

const header = '# Guía\n\nEstado: Vigente. Responsable: Ingeniería. Revisión documental: 2026-09-05.\n';
function fixture(t) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'pcs-docs-test-'));
  t.after(() => {
    const resolved = path.resolve(root);
    assert.equal(path.dirname(resolved), path.resolve(os.tmpdir()));
    assert.ok(path.basename(resolved).startsWith('pcs-docs-test-'));
    fs.rmSync(resolved, { recursive: true, force: true });
  });
  execFileSync('git', ['init', '-q'], { cwd: root });
  const write = (p, value) => { fs.mkdirSync(path.dirname(path.join(root, p)), { recursive: true }); fs.writeFileSync(path.join(root, p), value); };
  const policy = { review_days: 90, maintained: { 'README.md': { owner: 'Ingeniería', reviewed: '2026-09-05' } } };
  write('README.md', header);
  return { root, write, policy };
}

test('inventory is deterministic, ignores ignored/private documents, normalizes CRLF', t => {
  const { root, write, policy } = fixture(t);
  write('.gitignore', 'private/\n');
  write('private/secret.md', '# Never catalog this');
  write('documentos/history.md', '# Reference');
  const first = buildCatalog(root, policy);
  assert.equal(first.catalog.documents.some(d => d.path.includes('secret')), false);
  assert.equal(first.catalog.issues.length, 0);
  assert.equal(first.json, buildCatalog(root, policy).json);
  write('README.md', header.replaceAll('\n', '\r\n'));
  assert.equal(first.json, buildCatalog(root, policy).json);
  assert.equal(first.catalog.documents.find(d => d.path === 'README.md').review_due, '2026-12-04');
});

test('maintained links fail for missing files, anchors, references and escaping repository', t => {
  const { root, write, policy } = fixture(t);
  write('README.md', header + '\n[missing](missing.md)\n[anchor](#absent)\n[outside](../outside.md)\n[text][undefined]\n');
  const issues = buildCatalog(root, policy).catalog.issues;
  assert.deepEqual(issues.map(i => i.type).sort(), ['ancla_ausente', 'destino_ausente', 'enlace_fuera_repositorio', 'referencia_no_definida'].sort());
  assert.ok(issues.every(i => i.blocking));
});

test('valid anchors, explicit reference links, directories and code locations pass; examples ignored', t => {
  const { root, write, policy } = fixture(t);
  write('src/main.go', 'package main\n');
  write('other.md', '# Other\n');
  write('README.md', header + '\n## Sección válida\n[ok](#sección-válida)\n## `Código`\n[heading](#código)\n[code](src/main.go#L1)\n[dir](src/)\n[link][ref]\n[ref]: other.md\n\n```markdown\n[example](missing.md)\n```\n`[inline](missing.md)`\n');
  assert.deepEqual(buildCatalog(root, policy).catalog.issues, []);
});

test('legacy findings remain visible and do not become false approvals', t => {
  const { root, write, policy } = fixture(t);
  write('documentos/old.md', '# Old\n[broken](lost.md)\n');
  const result = buildCatalog(root, policy).catalog;
  assert.equal(result.issues.length, 1);
  assert.equal(result.issues[0].blocking, false);
  const entry = result.documents.find(d => d.path === 'documentos/old.md');
  assert.equal(entry.state, 'referencia_por_validar');
  assert.equal(entry.reviewed, null);
});

test('missing governed source and invalid metadata or UTF-8 replacement fail', t => {
  const { root, write, policy } = fixture(t);
  policy.maintained['missing.md'] = { owner: 'Ingeniería', reviewed: '2026-09-05' };
  write('README.md', '# No metadata\n\uFFFD\n');
  const types = buildCatalog(root, policy).catalog.issues.filter(i => i.blocking).map(i => i.type);
  assert.ok(types.includes('fuente_ausente'));
  assert.ok(types.includes('metadatos_incompletos'));
  assert.ok(types.includes('utf8_reemplazo'));
});

test('CLI check is read-only, detects drift and rejects a broken maintained link', t => {
  const { root, write, policy } = fixture(t);
  const source = path.join(path.dirname(fileURLToPath(import.meta.url)), 'docs_catalog.mjs');
  write('tools/docs_catalog.mjs', fs.readFileSync(source));
  write('documentos/gobernanza_tecnica/politica_catalogo.json', JSON.stringify(policy));
  const run = flag => spawnSync(process.execPath, ['tools/docs_catalog.mjs', flag], { cwd: root, encoding: 'utf8' });
  assert.equal(run('--write').status, 0);
  assert.equal(run('--check').status, 0);
  const before = fs.readFileSync(path.join(root, 'documentos/catalogo_documental.json'), 'utf8');
  write('README.md', header + '\nNew content\n');
  assert.equal(run('--check').status, 1);
  assert.equal(fs.readFileSync(path.join(root, 'documentos/catalogo_documental.json'), 'utf8'), before);
  assert.equal(run('--write').status, 0);
  write('README.md', header + '\n[missing](missing.md)\n');
  assert.equal(run('--write').status, 1);
});
