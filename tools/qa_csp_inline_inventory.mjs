import fs from "node:fs";
import path from "node:path";

const webRoot = path.resolve(process.argv[2] || path.join(process.cwd(), "web"));
const outputDir = path.resolve(process.argv[3] || path.join(process.cwd(), "tmp", "qa_csp_inline"));
const baselinePath = path.resolve(process.argv[4] || path.join(process.cwd(), "documentos", "seguridad", "csp_inline_baseline.json"));

function walk(dir) {
  return fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) return walk(full);
    return entry.isFile() && entry.name.toLowerCase().endsWith(".html") ? [full] : [];
  });
}

function count(pattern, source) {
  return [...source.matchAll(pattern)].length;
}

function area(relative) {
  const first = relative.split("/")[0];
  return relative.includes("/") ? first : "publico";
}

function inspect(file) {
  const source = fs.readFileSync(file, "utf8");
  const relative = path.relative(webRoot, file).replaceAll("\\", "/");
  const scripts = [...source.matchAll(/<script\b([^>]*)>[\s\S]*?<\/script\s*>/gi)];
  const inlineScripts = scripts.filter((match) => !/\bsrc\s*=/i.test(match[1] || ""));
  const protectedScripts = inlineScripts.filter((match) => /\b(?:nonce|integrity)\s*=/i.test(match[1] || ""));
  const inlineStyles = [...source.matchAll(/<style\b([^>]*)>[\s\S]*?<\/style\s*>/gi)];
  const protectedStyles = inlineStyles.filter((match) => /\bnonce\s*=/i.test(match[1] || ""));
  return {
    file: relative,
    area: area(relative),
    inline_scripts: inlineScripts.length,
    unprotected_inline_scripts: inlineScripts.length - protectedScripts.length,
    inline_style_blocks: inlineStyles.length,
    unprotected_inline_style_blocks: inlineStyles.length - protectedStyles.length,
    style_attributes: count(/\sstyle\s*=\s*(?:"[^"]*"|'[^']*')/gi, source),
    event_attributes: count(/\son[a-z]+\s*=\s*(?:"[^"]*"|'[^']*')/gi, source),
    javascript_urls: count(/(?:href|src|action)\s*=\s*(?:"\s*javascript:|'\s*javascript:)/gi, source),
  };
}

function blockerCount(row) {
  return row.unprotected_inline_scripts + row.unprotected_inline_style_blocks +
    row.style_attributes + row.event_attributes + row.javascript_urls;
}

const rows = walk(webRoot).map(inspect).filter((row) => blockerCount(row) > 0);
rows.sort((a, b) => blockerCount(b) - blockerCount(a) || a.file.localeCompare(b.file));
const totals = rows.reduce((acc, row) => {
  for (const key of ["unprotected_inline_scripts", "unprotected_inline_style_blocks", "style_attributes", "event_attributes", "javascript_urls"]) {
    acc[key] += row[key];
  }
  return acc;
}, { unprotected_inline_scripts: 0, unprotected_inline_style_blocks: 0, style_attributes: 0, event_attributes: 0, javascript_urls: 0 });
totals.blockers = Object.values(totals).reduce((sum, value) => sum + value, 0);
totals.files = rows.length;

const areas = {};
for (const row of rows) {
  areas[row.area] ??= { blockers: 0, files: 0 };
  areas[row.area].blockers += blockerCount(row);
  areas[row.area].files += 1;
}

let baseline = null;
let regression = [];
if (fs.existsSync(baselinePath)) {
  baseline = JSON.parse(fs.readFileSync(baselinePath, "utf8"));
  for (const [key, value] of Object.entries(totals)) {
    if (Number(value) > Number(baseline.totals?.[key] ?? 0)) regression.push(`totals.${key}: ${value} > ${baseline.totals?.[key] ?? 0}`);
  }
  for (const [key, value] of Object.entries(areas)) {
    if (value.blockers > Number(baseline.areas?.[key]?.blockers ?? 0)) regression.push(`areas.${key}.blockers: ${value.blockers} > ${baseline.areas?.[key]?.blockers ?? 0}`);
  }
} else {
  regression.push(`baseline ausente: ${baselinePath}`);
}

fs.mkdirSync(outputDir, { recursive: true });
const jsonPath = path.join(outputDir, "inventario_csp_inline.json");
const markdownPath = path.join(outputDir, "inventario_csp_inline.md");
fs.writeFileSync(jsonPath, JSON.stringify({ generated_at: new Date().toISOString(), web_root: webRoot, totals, areas, regression, files: rows }, null, 2) + "\n");
fs.writeFileSync(markdownPath, [
  "# Inventario CSP inline",
  "",
  `- Archivos con bloqueos: **${totals.files}**`,
  `- Bloqueos totales: **${totals.blockers}**`,
  `- Scripts inline sin protección: **${totals.unprotected_inline_scripts}**`,
  `- Bloques style sin protección: **${totals.unprotected_inline_style_blocks}**`,
  `- Atributos style: **${totals.style_attributes}**`,
  `- Eventos inline: **${totals.event_attributes}**`,
  `- URLs javascript: **${totals.javascript_urls}**`,
  `- Regresiones contra baseline: **${regression.length}**`,
  "",
  "| Página | Área | Bloqueos | Script | Style tag | Style attr | Evento | JS URL |",
  "|---|---|---:|---:|---:|---:|---:|---:|",
  ...rows.slice(0, 80).map((row) => `| \`${row.file}\` | ${row.area} | ${blockerCount(row)} | ${row.unprotected_inline_scripts} | ${row.unprotected_inline_style_blocks} | ${row.style_attributes} | ${row.event_attributes} | ${row.javascript_urls} |`),
  "",
].join("\n"));

process.stdout.write(JSON.stringify({ ok: regression.length === 0, totals, areas, regression, json: jsonPath, markdown: markdownPath }) + "\n");
if (regression.length > 0) process.exitCode = 1;
