import fs from "node:fs";
import path from "node:path";
import vm from "node:vm";

const repoRoot = process.cwd();
const outArgIndex = process.argv.indexOf("--out");
const outDir = outArgIndex >= 0 && process.argv[outArgIndex + 1]
  ? path.resolve(repoRoot, process.argv[outArgIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

const now = new Date();
const stamp = now.toISOString().replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
const report = {
  generated_at: now.toISOString(),
  status: "ok",
  checks: [],
  summary: {},
};

function read(rel) {
  return fs.readFileSync(path.join(repoRoot, rel), "utf8");
}

function exists(rel) {
  return fs.existsSync(path.join(repoRoot, rel));
}

function walk(dirRel, predicate = () => true) {
  const dir = path.join(repoRoot, dirRel);
  const out = [];
  if (!fs.existsSync(dir)) return out;
  const stack = [dir];
  while (stack.length) {
    const current = stack.pop();
    for (const entry of fs.readdirSync(current, { withFileTypes: true })) {
      const full = path.join(current, entry.name);
      if (entry.isDirectory()) {
        if (entry.name === "node_modules" || entry.name === ".git") continue;
        stack.push(full);
      } else if (predicate(full)) {
        out.push(path.relative(repoRoot, full).replace(/\\/g, "/"));
      }
    }
  }
  return out.sort();
}

function addCheck(name, ok, details = {}) {
  report.checks.push({ name, ok, ...details });
  if (!ok) report.status = "warning";
}

function wordCount(value) {
  return String(value || "").trim().split(/\s+/).filter(Boolean).length;
}

function extractRegexAll(text, regex, group = 1) {
  return [...text.matchAll(regex)].map((m) => String(m[group] || "").trim()).filter(Boolean);
}

function loadPlantillasCatalog() {
  const source = read("web/js/plantillas_nuevas_catalogo.js");
  const ctx = { window: {} };
  vm.createContext(ctx);
  vm.runInContext(source, ctx, { filename: "plantillas_nuevas_catalogo.js" });
  return Array.isArray(ctx.window.PCS_NUEVAS_PLANTILLAS) ? ctx.window.PCS_NUEVAS_PLANTILLAS : [];
}

function validateInlineScripts() {
  const htmlFiles = walk("web", (full) => /\.(html|ht)$/i.test(full));
  const failures = [];
  let scriptCount = 0;
  for (const rel of htmlFiles) {
    const html = read(rel);
    let idx = 0;
    for (const match of html.matchAll(/<script(?![^>]*\bsrc=)[^>]*>([\s\S]*?)<\/script>/gi)) {
      idx += 1;
      scriptCount += 1;
      try {
        new Function(match[1]);
      } catch (err) {
        failures.push(`${rel} script ${idx}: ${err.message}`);
      }
    }
  }
  addCheck("frontend_inline_scripts_syntax", failures.length === 0, {
    files: htmlFiles.length,
    scripts: scriptCount,
    failures,
  });
}

function auditPlantillasCatalog() {
  const plantillas = loadPlantillasCatalog();
  // Colegio/academia is retired from PCS and must not return through a
  // frontend fallback catalog.
  const expectedVisibleTemplates = 19;
  const missing = plantillas.filter((item) => {
    return !item.module || !item.title || !item.fullTitle || !item.description || wordCount(item.description) < 25 || !Array.isArray(item.sections) || item.sections.length < 5;
  }).map((item) => item.module || item.id || item.title || "sin_clave");
  const duplicateModules = [...new Set(plantillas.map((x) => x.module).filter((x, i, arr) => x && arr.indexOf(x) !== i))];
  addCheck("plantillas_nuevas_catalogo", plantillas.length === expectedVisibleTemplates && missing.length === 0 && duplicateModules.length === 0, {
    expected_visible_total: expectedVisibleTemplates,
    total: plantillas.length,
    missing_or_incomplete: missing,
    duplicate_modules: duplicateModules,
  });
  report.summary.plantillas_nuevas = plantillas.map((item) => item.module);
}

function auditPermissionsAndMenu() {
  const handlers = read("backend/handlers/empresa_permisos.go");
  const adminJs = read("web/js/administrar_empresa.js");
  const backendModules = extractRegexAll(handlers, /permModule[A-Za-z0-9_]+\s+=\s+"([^"]+)"/g);
  const menuLinkRefs = extractRegexAll(adminJs, /getElementById\("([^"]+)"\)/g).filter((id) => id.startsWith("link"));
  const htmlIds = new Set(extractRegexAll(walk("web", (full) => /\.(html|ht)$/i.test(full)).map(read).join("\n"), /\bid="([^"]+)"/g));
  const missingLinkIds = [...new Set(menuLinkRefs.filter((id) => !htmlIds.has(id)))].sort();
  const wrappers = extractRegexAll(walk("backend", (full) => /\.go$/i.test(full)).map(read).join("\n"), /\bfunc\s+(WithEmpresa[A-Za-z0-9]+Permissions)\b/g);
  addCheck("permisos_roles_menu_integridad", backendModules.length >= 40 && wrappers.length >= 40, {
    backend_modules: backendModules.length,
    permission_wrappers: wrappers.length,
    menu_link_refs: menuLinkRefs.length,
    missing_menu_link_ids_warning: missingLinkIds.slice(0, 40),
    missing_menu_link_ids_total: missingLinkIds.length,
  });
  report.summary.permission_modules = backendModules.sort();
}

function auditPublicCommercialFlow() {
  const indexHtml = read("web/index.html");
  const detailHtml = read("web/descripcion_de_los_sistemas.html");
  const checks = {
    index_uses_vertical_description: /item\.description \|\| item\.descripcion_larga/.test(indexHtml),
    trial_carries_context: /accion', 'probar_gratis'/.test(indexHtml) && /tipo_empresa/.test(indexHtml),
    detail_uses_vertical_description: /item\.description \|\| item\.descripcion_larga/.test(detailHtml),
    contact_page_public: exists("web/Informacion_de_contacto.html"),
  };
  addCheck("portal_publico_y_probar_gratis", Object.values(checks).every(Boolean), checks);
}

function auditOperationsDocs() {
  const requiredDocs = [
    "documentos/manual_de_instalacion.md",
    "documentos/docker_vps_operacion.md",
    "deploy/README-compose-platform.md",
    "documentos/matriz_roles_permisos_pos_multiempresa.md",
    "documentos/descripcion_de_modulos",
    "documentos/descripcion_del_proyecto",
    "documentos/CHANGELOG.md",
    "documentos/plan_profesional_12_puntos.md",
    "documentos/api/openapi.generated.yaml",
    "documentos/gobernanza_tecnica/runbooks/runbook_recuperacion_desastre_docker_vps.md",
  ];
  const missing = requiredDocs.filter((rel) => !exists(rel));
  addCheck("documentacion_operativa_base", missing.length === 0, { required: requiredDocs, missing });
}

function auditCriticalScripts() {
  const required = [
    "scripts/sync_to_vps.ps1",
    "scripts/rs.ps1",
    "deploy/docker-compose.platform.yml",
    "deploy/docker-compose.staging.yml",
    ".github/workflows/professional-ci.yml",
  ];
  const missing = required.filter((rel) => !exists(rel));
  const syncText = exists("scripts/sync_to_vps.ps1") ? read("scripts/sync_to_vps.ps1") : "";
  const hasCleanup = /CleanupRemoteUnusedFiles/.test(syncText) && /docker builder prune/.test(syncText);
  addCheck("scripts_operacion_profesional", missing.length === 0 && hasCleanup, { required, missing, sync_cleanup: hasCleanup });
}

validateInlineScripts();
auditPlantillasCatalog();
auditPermissionsAndMenu();
auditPublicCommercialFlow();
auditOperationsDocs();
auditCriticalScripts();

fs.mkdirSync(outDir, { recursive: true });
const jsonPath = path.join(outDir, `professional_audit_${stamp}.json`);
const mdPath = path.join(outDir, `professional_audit_${stamp}.md`);
fs.writeFileSync(jsonPath, JSON.stringify(report, null, 2), "utf8");

const lines = [];
lines.push(`# Auditoria profesional de plataforma`);
lines.push("");
lines.push(`Fecha: ${report.generated_at}`);
lines.push(`Estado: ${report.status}`);
lines.push("");
for (const check of report.checks) {
  lines.push(`## ${check.ok ? "OK" : "REVISAR"} - ${check.name}`);
  const clone = { ...check };
  delete clone.name;
  delete clone.ok;
  lines.push("```json");
  lines.push(JSON.stringify(clone, null, 2));
  lines.push("```");
  lines.push("");
}
fs.writeFileSync(mdPath, lines.join("\n"), "utf8");

console.log(JSON.stringify({ status: report.status, json: jsonPath, markdown: mdPath }, null, 2));
if (report.status !== "ok") {
  process.exitCode = 2;
}
