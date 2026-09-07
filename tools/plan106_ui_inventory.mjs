#!/usr/bin/env node
/**
 * Inventario estático P106 de controles visibles y acciones declaradas.
 * No navega ni ejecuta JavaScript: sirve de manifiesto previo al barrido E2E.
 */
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const webRoot = path.join(root, "web");
const output = path.join(root, "documentos", "arquitectura", "inventario_ui_plan_106.md");
const ignoredDirectories = new Set(["source", "uploads", "node_modules"]);
const controlPattern = /<(?<tag>button|a|input|select|textarea)\b(?<attributes>[^>]*)>(?<content>[\s\S]*?)<\/\k<tag>>|<input\b(?<inputAttributes>[^>]*)\/?\s*>/gi;

function walk(directory, files = []) {
  for (const item of fs.readdirSync(directory, { withFileTypes: true })) {
    const fullPath = path.join(directory, item.name);
    if (item.isDirectory() && !ignoredDirectories.has(item.name)) walk(fullPath, files);
    if (item.isFile() && item.name.endsWith(".html")) files.push(fullPath);
  }
  return files;
}

function attribute(attributes, name) {
  const match = new RegExp("\\b" + name + "\\s*=\\s*([\\\"'])(.*?)\\1", "i").exec(attributes || "");
  return match ? match[2].trim() : "";
}

function normalizedText(value) {
  return String(value || "").replace(/<[^>]*>/g, " ").replace(/&nbsp;/gi, " ").replace(/\s+/g, " ").trim();
}

function classify(tag, attributes, content) {
  const type = attribute(attributes, "type").toLowerCase();
  const role = attribute(attributes, "role").toLowerCase();
  const className = attribute(attributes, "class");
  const hasAction = /\bonclick\s*=|\bdata-(action|toggle|modal|target|tab|view|command)\s*=/i.test(attributes);
  if (tag === "button" || role === "button" || hasAction) return "accion";
  if (tag === "input" && ["button", "submit", "reset", "file", "checkbox", "radio"].includes(type)) return "accion";
  if (tag === "a" && (/(^|\s)(btn|button)(\s|$)/i.test(className) || attribute(attributes, "href") === "#")) return "accion";
  if (tag === "select" || tag === "textarea" || tag === "input") return "entrada";
  return "navegacion";
}

function extractControls(file) {
  const source = fs.readFileSync(file, "utf8");
  const records = [];
  let match;
  let number = 0;
  while ((match = controlPattern.exec(source)) !== null) {
    const tag = (match.groups.tag || "input").toLowerCase();
    const attributes = match.groups.attributes || match.groups.inputAttributes || "";
    const content = match.groups.content || "";
    const type = attribute(attributes, "type").toLowerCase();
    const label = normalizedText(content) || attribute(attributes, "value") || attribute(attributes, "aria-label") || attribute(attributes, "title") || attribute(attributes, "name") || attribute(attributes, "id") || "sin etiqueta";
    const category = classify(tag, attributes, content);
    if (category === "navegacion" && tag === "a") continue;
    number += 1;
    records.push({
      number,
      tag,
      type: type || "-",
      category,
      id: attribute(attributes, "id") || "-",
      label: label.slice(0, 120),
      dynamic: /\bonclick\s*=|\bdata-|{{|<%/i.test(attributes + content) ? "sí" : "no"
    });
  }
  return records;
}

const files = walk(webRoot).sort();
const pages = files.map((file) => ({ file: path.relative(root, file).replaceAll("\\", "/"), controls: extractControls(file) }));
const totals = pages.reduce((result, page) => {
  result.pages += 1;
  result.controls += page.controls.length;
  result.actions += page.controls.filter((control) => control.category === "accion").length;
  result.inputs += page.controls.filter((control) => control.category === "entrada").length;
  result.dynamic += page.controls.filter((control) => control.dynamic === "sí").length;
  return result;
}, { pages: 0, controls: 0, actions: 0, inputs: 0, dynamic: 0 });

const lines = [
  "# Inventario estático de interfaz - Plan 106",
  "",
  "Generado por `node tools/plan106_ui_inventory.mjs`. No editar manualmente.",
  "",
  "## Alcance",
  "",
  `- Páginas HTML: **${totals.pages}**`,
  `- Controles detectados: **${totals.controls}**`,
  `- Acciones a cubrir en E2E: **${totals.actions}**`,
  `- Entradas y selectores: **${totals.inputs}**`,
  `- Controles con marcador dinámico: **${totals.dynamic}**`,
  "- Estado: inventario estático previo; la cobertura funcional, visual, por permisos y de IA se registra en el runner E2E y la matriz P106.",
  "",
  "## Controles por página",
  ""
];

for (const page of pages) {
  if (!page.controls.length) continue;
  lines.push(`### \`${page.file}\` (${page.controls.length})`, "", "| # | Tipo | Clase | ID | Etiqueta | Dinámico |", "| ---: | --- | --- | --- | --- | --- |");
  for (const control of page.controls) {
    const safe = (value) => String(value).replaceAll("|", "\\|").replaceAll("`", "'");
    lines.push(`| ${control.number} | ${safe(control.tag + "/" + control.type)} | ${control.category} | ${safe(control.id)} | ${safe(control.label)} | ${control.dynamic} |`);
  }
  lines.push("");
}

fs.writeFileSync(output, lines.join("\n"), "utf8");
process.stdout.write(JSON.stringify({ output: path.relative(root, output).replaceAll("\\", "/"), ...totals }) + "\n");
