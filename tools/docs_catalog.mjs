// Dependency-free documentation inventory. It never connects to PCS or providers.
import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { fileURLToPath } from 'node:url';

const outputs = ['documentos/catalogo_documental.json', 'documentos/catalogo_documental.md'];
const policyPath = 'documentos/gobernanza_tecnica/politica_catalogo.json';
const normalize = value => value.replace(/^\uFEFF/, '').replace(/\r\n?/g, '\n');
const compare = (a, b) => a < b ? -1 : a > b ? 1 : 0;
const isProse = p => /\.(md|rst|adoc|txt)$/i.test(p) || !path.extname(p);
const isDocument = p => /\.(md|rst|adoc)$/i.test(p)
  || /^documentos\/.*\.(txt|yaml|yml|json|mmd|puml)$/i.test(p)
  || /^documentos\/(descripcion_de_archivos|descripcion_de_modulos|descripcion_del_proyecto|historial_de_cambios)$/.test(p);

function visibleMarkdown(text, keepInlineText = false) {
  // Keep line positions while excluding code examples from link/heading checks.
  let fence = null;
  return text.split('\n').map(line => {
    const marker = line.match(/^\s{0,3}(`{3,}|~{3,})/);
    if (marker) {
      if (!fence) fence = marker[1];
      else if (marker[1][0] === fence[0] && marker[1].length >= fence.length) fence = null;
      return '';
    }
    return fence ? '' : line.replace(/(`+)(.*?)\1/g, keepInlineText ? '$2' : '');
  }).join('\n');
}

function anchors(text) {
  const found = new Set();
  const duplicates = new Map();
  for (const line of visibleMarkdown(text, true).split('\n')) {
    const match = line.match(/^#{1,6}\s+(.+?)\s*#*\s*$/);
    if (match) {
      const slug = match[1].toLowerCase().replace(/<[^>]*>/g, '').replace(/[^\p{L}\p{N}_\-\s]/gu, '').replace(/\s/g, '-');
      const n = duplicates.get(slug) || 0;
      duplicates.set(slug, n + 1);
      found.add(n ? `${slug}-${n}` : slug);
    }
    for (const m of line.matchAll(/\b(?:id|name)=["']([^"']+)["']/g)) found.add(m[1]);
  }
  return found;
}

export function buildCatalog(root, policy) {
  const listed = execFileSync('git', ['ls-files', '--cached', '--others', '--exclude-standard', '-z'], {
    cwd: root, encoding: 'utf8', maxBuffer: 32 * 1024 * 1024,
  }).split('\0');
  const files = [...new Set([...listed.filter(isDocument), ...outputs])]
    .filter(p => p && (outputs.includes(p) || fs.existsSync(path.join(root, p))))
    .sort(compare);
  const maintained = policy.maintained || {};
  const issues = [];
  const contents = new Map();
  const anchorCache = new Map();
  const read = p => {
    if (!contents.has(p)) contents.set(p, normalize(fs.readFileSync(path.join(root, p), 'utf8')));
    return contents.get(p);
  };
  const classify = p => {
    if (outputs.includes(p)) return 'generado';
    if (maintained[p]) return 'vigente';
    if (policy.overrides?.[p]) return policy.overrides[p];
    if (/(^|\/)(historico|evidencia[^/]*|releases)\//.test(p)) return p.includes('/historico/') ? 'historico' : 'evidencia';
    if (p.includes('/referencias/')) return 'referencia_externa';
    if (/(^|\/)(plan_10[1-9][^/]*|plan_110[^/]*|plan_final_para_produccion|CHANGELOG|historial_de_cambios|actualizaciones_del_repositorio)(\.|$)/i.test(p)) return 'historico';
    if (/(^|\/)(auditoria_|informe_|reporte_|evidencia_|estado_documentacion_|.*_report\.|production_readiness|staging_.*checklist)/i.test(p)) return 'evidencia';
    if (/\.(json|yaml|yml|mmd|puml)$/.test(p) || /\/(inventario_|matriz_rutas_multiempresa|documentacion_tecnica_completa)/.test(p)) return 'generado';
    return 'referencia_por_validar';
  };
  const owner = p => p.includes('/seguridad/') || p === 'SECURITY.md' ? 'Seguridad'
    : /\/(api|arquitectura|erp_multiempresa)\//.test(p) ? 'Ingeniería backend y datos'
    : /\/(operacion|evidencia[^/]*|releases)\//.test(p) || /runbook|_report\./.test(p) ? 'QA/operación'
    : 'Coordinación técnica';
  for (const p of Object.keys(maintained)) {
    if (!files.includes(p)) issues.push({ path: p, line: 1, type: 'fuente_ausente', target: p, blocking: true });
  }
  const documents = files.map(p => {
    const self = outputs.includes(p);
    const text = self ? '' : read(p);
    const state = classify(p);
    const meta = maintained[p];
    if (meta && (!/^# .+/m.test(text) || !/Estado:/.test(text) || !/Responsable:/.test(text)
      || !text.includes(`Revisión documental: ${meta.reviewed}`) || !/^\d{4}-\d{2}-\d{2}$/.test(meta.reviewed))) {
      issues.push({ path: p, line: 1, type: 'metadatos_incompletos', blocking: true });
    }
    if (text.includes('\uFFFD')) issues.push({ path: p, line: 1, type: 'utf8_reemplazo', blocking: state === 'vigente' });
    const review = meta?.reviewed || null;
    return {
      path: p,
      title: self ? 'Catálogo documental generado' : (visibleMarkdown(text).match(/^# (.+)$/m)?.[1] || path.basename(p)).replace(/\|/g, ' / '),
      state,
      owner: meta?.owner || owner(p),
      reviewed: review,
      review_due: review ? new Date(Date.parse(`${review}T00:00:00Z`) + policy.review_days * 86400000).toISOString().slice(0, 10) : null,
      // Self outputs have no content hash to avoid recursive catalog churn.
      sha256_lf: self ? null : createHash('sha256').update(text).digest('hex'),
    };
  });
  for (const entry of documents) {
    if (outputs.includes(entry.path) || !isProse(entry.path)) continue;
    const lines = visibleMarkdown(read(entry.path)).split('\n');
    const refs = new Map();
    const refKey = s => s.trim().replace(/\s+/g, ' ').toLowerCase();
    for (const line of lines) {
      const m = line.match(/^\s{0,3}\[([^\]]+)\]:\s*(?:<([^>]+)>|(\S+))/);
      if (m) refs.set(refKey(m[1]), m[2] || m[3]);
    }
    lines.forEach((line, index) => {
      const targets = [...line.matchAll(/!?\[[^\]\n]*\]\(\s*(?:<([^>]+)>|([^\s)]+))(?:\s+["'][^\n]*?["'])?\s*\)/g)].map(m => m[1] || m[2]);
      if (!/^\s*\[[^\]]+\]:/.test(line)) {
        for (const m of line.matchAll(/!?\[([^\]]+)\]\[([^\]]*)\]/g)) {
          const label = refKey(m[2] || m[1]);
          if (refs.has(label)) targets.push(refs.get(label));
          else issues.push({ path: entry.path, line: index + 1, type: 'referencia_no_definida', target: label, blocking: entry.state === 'vigente' });
        }
      }
      for (const raw of targets) {
        if (/^(?:[a-z][a-z0-9+.-]*:|\/\/)/i.test(raw)) continue;
        let decoded;
        try { decoded = decodeURIComponent(raw); } catch { decoded = raw; }
        const [location, fragment] = decoded.split('#');
        const destination = location.split('?')[0];
        const absolute = destination ? path.resolve(destination.startsWith('/') ? root : path.dirname(path.join(root, entry.path)), destination.replace(/^\//, '')) : path.join(root, entry.path);
        const relative = path.relative(root, absolute).replaceAll('\\', '/');
        let type;
        if (relative.startsWith('../') || path.isAbsolute(relative)) type = 'enlace_fuera_repositorio';
        else if (!fs.existsSync(absolute) && !outputs.includes(relative)) type = 'destino_ausente';
        else if (fragment && /\.md$/i.test(relative) && !outputs.includes(relative) && fs.statSync(absolute).isFile()) {
          if (!anchorCache.has(relative)) anchorCache.set(relative, anchors(read(relative)));
          if (!anchorCache.get(relative).has(fragment)) type = 'ancla_ausente';
        }
        if (type) issues.push({ path: entry.path, line: index + 1, type, target: raw, blocking: entry.state === 'vigente' });
      }
    });
  }
  const hashes = new Map();
  for (const entry of documents) if (entry.sha256_lf) {
    const group = hashes.get(entry.sha256_lf) || [];
    group.push(entry.path);
    hashes.set(entry.sha256_lf, group);
  }
  const counts = {};
  for (const doc of documents) counts[doc.state] = (counts[doc.state] || 0) + 1;
  const result = {
    schema_version: 1,
    policy: policyPath,
    scope: 'Documentos de texto versionados o nuevos no ignorados por Git; metadatos técnicos bajo documentos. Excluye datos privados ignorados y binarios. Sin auditoría semántica ni validación de URLs externas.',
    total: documents.length,
    counts,
    issues,
    duplicates: [...hashes.values()].filter(g => g.length > 1),
    documents,
  };
  const markdown = [
    '# Catálogo documental de PCS', '',
    'Estado: Generado. Responsable: Coordinación técnica. Revisión: determinada por cada fuente.', '',
    'No editar manualmente. Generar con `node tools/docs_catalog.mjs --write`; validar con `--check`.', '',
    'Política y significado de estados: [marco documental](gobernanza_tecnica/marco_documental.md).',
    'Inventario y hallazgos detallados: [JSON](catalogo_documental.json). La clasificación no acredita revisión semántica ni producción.', '',
    `Documentos: ${result.total}. Hallazgos locales: ${issues.length}; bloqueantes: ${issues.filter(i => i.blocking).length}.`, '',
    '| Estado | Cantidad |', '| --- | --- |',
    ...Object.keys(counts).sort(compare).map(k => `| ${k} | ${counts[k]} |`), '',
    '## Índice completo', '',
    '| Documento | Estado | Responsable | Revisado |', '| --- | --- | --- | --- |',
    ...documents.map(d => {
      const href = path.posix.relative('documentos', d.path).split('/').map(encodeURIComponent).join('/');
      return `| [${d.path}](<${href}>) | ${d.state} | ${d.owner} | ${d.reviewed || 'Pendiente / no aplica'} |`;
    }), '',
    '## Hallazgos heredados', '',
    'Las fuentes vigentes fallan el control ante enlaces/metadata inválidos. Los hallazgos de historia, evidencia y referencias por validar se informan sin reescribir esas fuentes. El JSON indica archivo, línea y destino; deben resolverse al revisar el módulo. No son excepciones de seguridad ni resultados de pruebas funcionales.', '',
  ].join('\n');
  return { catalog: result, json: JSON.stringify(result, null, 2) + '\n', markdown };
}

function main() {
  const args = process.argv.slice(2);
  if (args.length !== 1 || !['--write', '--check'].includes(args[0])) throw new Error('Uso: node tools/docs_catalog.mjs --write | --check');
  const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
  const policy = JSON.parse(fs.readFileSync(path.join(root, policyPath), 'utf8'));
  const result = buildCatalog(root, policy);
  const rendered = [result.json, result.markdown];
  const stale = outputs.filter((p, i) => !fs.existsSync(path.join(root, p)) || normalize(fs.readFileSync(path.join(root, p), 'utf8')) !== rendered[i]);
  const blocking = result.catalog.issues.filter(i => i.blocking);
  if (args[0] === '--write') for (let i = 0; i < outputs.length; i++) fs.writeFileSync(path.join(root, outputs[i]), rendered[i], 'utf8');
  console.log(JSON.stringify({ documents: result.catalog.total, states: result.catalog.counts, findings: result.catalog.issues.length, blocking, stale: args[0] === '--check' ? stale : [], mode: args[0] }, null, 2));
  if (blocking.length || (args[0] === '--check' && stale.length)) process.exitCode = 1;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try { main(); } catch (error) { console.error(error.message); process.exitCode = 1; }
}
