#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outArgIndex = process.argv.indexOf("--out");
const strict = process.argv.includes("--strict");
const outDir = outArgIndex >= 0 && process.argv[outArgIndex + 1]
  ? path.resolve(repoRoot, process.argv[outArgIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

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
        if (entry.name === ".git" || entry.name === "node_modules") continue;
        stack.push(full);
      } else if (predicate(full)) {
        out.push(path.relative(repoRoot, full).replace(/\\/g, "/"));
      }
    }
  }
  return out.sort();
}

const checks = [];
function add(name, ok, details = {}, severity = "medium") {
  checks.push({ name, ok, severity, ...details });
}

const backendGo = walk("backend", (full) => full.endsWith(".go")).map((rel) => [rel, read(rel)]);
const allGo = backendGo.map(([, text]) => text).join("\n");
const mainGo = exists("backend/main.go") ? read("backend/main.go") : "";
const utilsGo = exists("backend/utils/utils.go") ? read("backend/utils/utils.go") : "";

add("cookies_http_only", /HttpOnly:\s*true/.test(allGo), {
  evidence: "Busca HttpOnly:true en cookies de sesion.",
}, "high");

add("cookies_same_site", /SameSite:\s*http\.SameSite(Lax|Strict)Mode/.test(allGo), {
  evidence: "Busca SameSite Lax o Strict en cookies.",
}, "high");

add("cookies_secure_helper", /SessionCookieSecure/.test(allGo) && /Secure:\s*handlers\.SessionCookieSecure|Secure:\s*SessionCookieSecure/.test(allGo), {
  evidence: "Usa helper de cookie segura segun request/entorno.",
}, "high");

add("session_revocation_logout", /RevokeSession|revoke session|Revocar/i.test(allGo) && /\/logout/.test(mainGo), {
  evidence: "Logout debe revocar token y limpiar cookie.",
}, "medium");

add("public_route_allowlist", /publicExact\s*:=|publicPaths\s*:=|allowedPublic|public route|rutas p.blicas/i.test(utilsGo) && /\/super\/api\/administradores\/login/.test(utilsGo), {
  evidence: "Middleware central con rutas publicas explicitas.",
}, "high");

add("login_rate_limit_contract", /rate.?limit|intentos|bloqueo|failed.?login|login.?attempt/i.test(allGo), {
  recommendation: "Agregar limitador por IP/correo si este chequeo queda en warning.",
}, "high");

add("recaptcha_contract", /RECAPTCHA|recaptcha/i.test(allGo + "\n" + walk("web", (full) => /\.(html|js)$/i.test(full)).map(read).join("\n")), {
  evidence: "Recaptcha o verificacion humana presente en login/flujo publico.",
}, "medium");

add("super_admin_2fa_readiness", /two.?factor|2fa|totp|mfa|otp/i.test(allGo + "\n" + walk("web", (full) => /\.(html|js)$/i.test(full)).map(read).join("\n")), {
  recommendation: "Activar 2FA/TOTP obligatorio para super administrador antes de delegar despliegues automaticos a produccion.",
}, "medium");

add("cors_not_wildcard_credentials", !/Access-Control-Allow-Origin",\s*"\*"/.test(allGo), {
  evidence: "No se detecta CORS wildcard directo en Go.",
}, "high");

add("empresa_scope_middleware", /WithEmpresa|empresa_id|Empresa.*Scope/.test(allGo) && /WithEmpresaPublicScope/.test(mainGo), {
  evidence: "Existe middleware/scope multiempresa.",
}, "high");

add("secret_files_not_committed", !exists("deploy/.env.platform") && !exists("deploy/.env.staging"), {
  evidence: "No versionar env reales con secretos.",
}, "critical");

const report = {
  generated_at: new Date().toISOString(),
  status: checks.every((c) => c.ok || c.severity === "medium") ? "ok" : "warning",
  checks,
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
const jsonPath = path.join(outDir, `security_audit_${stamp}.json`);
const mdPath = path.join(outDir, `security_audit_${stamp}.md`);
fs.writeFileSync(jsonPath, JSON.stringify(report, null, 2), "utf8");
fs.writeFileSync(mdPath, [
  "# Auditoria de seguridad",
  "",
  `Fecha: ${report.generated_at}`,
  `Estado: ${report.status}`,
  "",
  ...checks.flatMap((check) => [
    `## ${check.ok ? "OK" : "REVISAR"} - ${check.name}`,
    "```json",
    JSON.stringify(check, null, 2),
    "```",
    "",
  ]),
].join("\n"), "utf8");

console.log(JSON.stringify({ status: report.status, json: jsonPath, markdown: mdPath }, null, 2));
if (strict && report.status !== "ok") process.exitCode = 2;
