#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outArgIndex = process.argv.indexOf("--out");
const strict = process.argv.includes("--strict");
const outDir = outArgIndex >= 0 && process.argv[outArgIndex + 1]
  ? path.resolve(repoRoot, process.argv[outArgIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

function walk(dir) {
  const root = path.join(repoRoot, dir);
  const out = [];
  const stack = [root];
  while (stack.length) {
    const current = stack.pop();
    for (const entry of fs.readdirSync(current, { withFileTypes: true })) {
      const full = path.join(current, entry.name);
      if (entry.isDirectory()) {
        if (entry.name === "uploads" || entry.name === "source") continue;
        stack.push(full);
      } else if (entry.isFile() && /\.(html|ht)$/i.test(entry.name)) {
        out.push(full);
      }
    }
  }
  return out;
}

const files = walk("web");
const issues = [];
let buttonCount = 0;
let submenuCount = 0;
for (const file of files) {
  const rel = path.relative(repoRoot, file).replace(/\\/g, "/");
  const html = fs.readFileSync(file, "utf8");
  buttonCount += (html.match(/<button\b/gi) || []).length;
  if (/submenu|sub-menu|tabs|tab-button|section-nav/i.test(html)) submenuCount += 1;
  for (const match of html.matchAll(/<button\b([^>]*)>/gi)) {
    const attrs = match[1] || "";
    if (!/\b(class|aria-label|title)=/i.test(attrs)) {
      issues.push({ file: rel, issue: "button_without_class_or_label", sample: match[0].slice(0, 140) });
      if (issues.length >= 80) break;
    }
  }
}

const report = {
  generated_at: new Date().toISOString(),
  status: issues.length <= 80 && submenuCount >= 20 ? "ok" : "warning",
  summary: { files: files.length, buttons: buttonCount, pages_with_submenu_signal: submenuCount, issue_samples: issues.length },
  issues,
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
const jsonPath = path.join(outDir, `ux_consistency_audit_${stamp}.json`);
const mdPath = path.join(outDir, `ux_consistency_audit_${stamp}.md`);
fs.writeFileSync(jsonPath, JSON.stringify(report, null, 2), "utf8");
fs.writeFileSync(mdPath, `# Auditoria UX global\n\nEstado: ${report.status}\n\n\`\`\`json\n${JSON.stringify(report, null, 2)}\n\`\`\`\n`, "utf8");
console.log(JSON.stringify({ status: report.status, json: jsonPath, markdown: mdPath }, null, 2));
if (strict && report.status !== "ok") process.exitCode = 2;
