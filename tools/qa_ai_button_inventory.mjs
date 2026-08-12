import fs from "node:fs";
import path from "node:path";

const root = path.resolve(process.argv[2] || path.join(process.cwd(), "web", "administrar_empresa"));
const outputDir = path.resolve(process.argv[3] || path.join(process.cwd(), "tmp", "qa_ai_buttons"));

function walk(dir) {
  return fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) return walk(full);
    return entry.isFile() && entry.name.toLowerCase().endsWith(".html") ? [full] : [];
  });
}

function attr(raw, name) {
  const match = raw.match(new RegExp(`\\b${name}\\s*=\\s*(["'])(.*?)\\1`, "i"));
  return match ? match[2].trim() : "";
}

function cleanLabel(raw) {
  return raw
    .replace(/<img\b[^>]*>/gi, " ")
    .replace(/<[^>]+>/g, " ")
    .replace(/&nbsp;/gi, " ")
    .replace(/&aacute;/gi, "á").replace(/&eacute;/gi, "é")
    .replace(/&iacute;/gi, "í").replace(/&oacute;/gi, "ó").replace(/&uacute;/gi, "ú")
    .replace(/&ntilde;/gi, "ñ")
    .replace(/\s+/g, " ")
    .trim();
}

function isAIButton(attrs, label) {
  return /\bdata-ai-button\b/i.test(attrs) ||
    /\bdata-run-ai\b/i.test(attrs) ||
	/\bclass\s*=\s*(["'])[^"']*\bai-(?:action|button)/i.test(attrs) ||
	/IA(?:$|[A-Z0-9_-])/.test(attr(attrs, "id")) ||
    /\b(?:IA|GPT-[0-9.]+)\b/i.test(label) ||
    /inteligencia artificial/i.test(label);
}

const rows = [];
for (const file of walk(root)) {
  const source = fs.readFileSync(file, "utf8");
  const buttonPattern = /<button\b([^>]*)>([\s\S]*?)<\/button>/gi;
  for (const match of source.matchAll(buttonPattern)) {
    const attrs = match[1] || "";
    const label = cleanLabel(match[2] || "");
    if (!isAIButton(attrs, label)) continue;
    const before = source.slice(0, match.index);
    rows.push({
      file: path.relative(process.cwd(), file).replaceAll("\\", "/"),
      line: before.split(/\r?\n/).length,
      id: attr(attrs, "id"),
      action: attr(attrs, "data-run-ai") || attr(attrs, "data-action"),
      label,
      disabled: /\bdisabled\b/i.test(attrs),
      hidden: /\bhidden\b/i.test(attrs),
    });
  }
}

rows.sort((a, b) => a.file.localeCompare(b.file) || a.line - b.line || a.label.localeCompare(b.label));
fs.mkdirSync(outputDir, { recursive: true });
const jsonPath = path.join(outputDir, "inventario_botones_ia.json");
const mdPath = path.join(outputDir, "inventario_botones_ia.md");
fs.writeFileSync(jsonPath, JSON.stringify({ generated_at: new Date().toISOString(), root, total: rows.length, buttons: rows }, null, 2) + "\n");
const markdown = [
  "# Inventario reproducible de botones IA empresariales",
  "",
  `- Total: **${rows.length}**`,
  `- Raíz: \`${path.relative(process.cwd(), root).replaceAll("\\", "/")}\``,
  "",
  "| Página | Línea | ID/acción | Etiqueta | Estado inicial |",
  "|---|---:|---|---|---|",
  ...rows.map((row) => `| \`${row.file}\` | ${row.line} | \`${row.id || row.action || "sin-id"}\` | ${row.label.replaceAll("|", "\\|")} | ${row.hidden ? "oculto" : row.disabled ? "deshabilitado" : "visible"} |`),
  "",
].join("\n");
fs.writeFileSync(mdPath, markdown);
process.stdout.write(JSON.stringify({ ok: true, total: rows.length, json: jsonPath, markdown: mdPath }) + "\n");
