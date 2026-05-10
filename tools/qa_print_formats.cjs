#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");
const vm = require("vm");
const { chromium } = require("playwright");

const ROOT = path.resolve(__dirname, "..");
const OUT_DIR = process.env.PCS_QA_PRINT_OUT_DIR || path.join(
  ROOT,
  "test_runs",
  "qa_print_formats_" + new Date().toISOString().replace(/[:.]/g, "-").slice(0, 19)
);
const PRINT_JS = path.join(ROOT, "web", "js", "print_documents.js");

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}

function slug(value) {
  return String(value || "documento")
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/[^\w.-]+/g, "_")
    .replace(/^_+|_+$/g, "")
    .toLowerCase()
    .slice(0, 90) || "documento";
}

function money(value) {
  return "$" + Number(value || 0).toLocaleString("es-CO");
}

function loadPCSPrint() {
  const source = fs.readFileSync(PRINT_JS, "utf8");
  const sandbox = {
    window: {
      open: () => null
    }
  };
  vm.runInNewContext(source, sandbox, { filename: PRINT_JS });
  if (!sandbox.window.PCSPrint || typeof sandbox.window.PCSPrint.buildDocument !== "function") {
    throw new Error("No se pudo cargar window.PCSPrint desde web/js/print_documents.js");
  }
  return sandbox.window.PCSPrint;
}

function baseRows() {
  return [
    ["Hospedaje habitacion 203", "1", money(120000), money(120000)],
    ["Servicio restaurante y minibar", "2", money(18500), money(37000)],
    ["Descuento convenio corporativo", "1", "-" + money(7000), "-" + money(7000)]
  ];
}

function documentOptions(kind, format) {
  const today = "2026-05-10 15:30";
  const common = {
    format,
    autoPrint: false,
    closeAfterPrint: false,
    company: "Motel Calipso",
    subtitle: "NIT 900123456-7 | Regimen comun | Cra 45 # 12-34, Colombia",
    footer: "Documento generado por Powerful Control System",
    tableHeaders: ["Detalle", "Cant.", "Unit.", "Total"],
    rows: baseRows(),
    signatures: format === "carta"
  };
  const byKind = {
    factura: {
      title: "Factura electronica de venta",
      kind: "factura",
      badge: format === "pos" ? "POS" : "Carta",
      meta: [
        { label: "Codigo", value: "FE-MC-000382" },
        { label: "Numero legal", value: "SETP990000382" },
        { label: "Fecha", value: today },
        { label: "Cliente", value: "Cliente final pruebas visuales" },
        { label: "Ambiente", value: "Habilitacion" },
        { label: "Estado", value: "Emitida" }
      ],
      totalLabel: "Total documento",
      totalValue: money(150000),
      note: "CUFE/CUDE: 9f48d172d4c9a5df084a1de74b2bb0a73f790ce1e6c2af0b58b4bff9b2a1d8c3. Resolucion y condiciones legales visibles en la impresion."
    },
    recibo: {
      title: "Recibo de venta",
      kind: "recibo",
      badge: format === "pos" ? "POS" : "Carta",
      meta: [
        { label: "Codigo", value: "RC-MC-000941" },
        { label: "Fecha", value: today },
        { label: "Caja", value: "CAJA-PRINCIPAL" },
        { label: "Metodo", value: "Efectivo" },
        { label: "Cliente", value: "Consumidor final" }
      ],
      totalLabel: "Pagado",
      totalValue: money(150000),
      note: "Gracias por su compra. Conserve este recibo para cualquier aclaracion."
    },
    comprobante_ingreso: {
      title: "Comprobante de ingreso",
      kind: "comprobante",
      badge: format === "pos" ? "POS" : "Carta",
      tableHeaders: ["Campo", "Detalle"],
      rows: [
        ["Empresa", { value: "Motel Calipso", number: false }],
        ["Codigo", { value: "ING-20260510-015", number: false }],
        ["Categoria", { value: "Operacion / Caja", number: false }],
        ["Concepto", { value: "Pago de hospedaje y servicios", number: false }],
        ["Metodo de pago", { value: "Transferencia bancaria", number: false }],
        ["Referencia", { value: "TRX-558801234567890", number: false }]
      ],
      totalLabel: "Total neto",
      totalValue: money(150000),
      note: "Comprobante adjunto: soporte_transferencia_55880.pdf"
    },
    comprobante_egreso: {
      title: "Comprobante de egreso",
      kind: "comprobante",
      badge: format === "pos" ? "POS" : "Carta",
      tableHeaders: ["Campo", "Detalle"],
      rows: [
        ["Empresa", { value: "Motel Calipso", number: false }],
        ["Codigo", { value: "EGR-20260510-006", number: false }],
        ["Categoria", { value: "Proveedores", number: false }],
        ["Concepto", { value: "Compra de insumos de limpieza", number: false }],
        ["Tercero", { value: "Proveedor Calipso SAS", number: false }],
        ["Comprobante", { value: "Factura FV-88210", number: false }]
      ],
      totalLabel: "Total neto",
      totalValue: money(98500),
      note: "Retencion calculada y observaciones listas para firma."
    },
    orden: {
      title: "Orden de servicio",
      kind: "orden",
      badge: format === "pos" ? "POS" : "Carta",
      meta: [
        { label: "Orden", value: "ORD-20260510-044" },
        { label: "Destino", value: "Habitacion 203" },
        { label: "Area", value: "Restaurante" },
        { label: "Prioridad", value: "Normal" }
      ],
      totalLabel: "Items",
      totalValue: "3",
      note: "Entregar en recepcion y confirmar con numero de habitacion."
    },
    corte_caja: {
      title: "Corte de caja",
      kind: "comprobante",
      badge: format === "pos" ? "POS" : "Carta",
      tableHeaders: ["Campo", "Valor"],
      rows: [
        ["Caja", { value: "CAJA-PRINCIPAL", number: false }],
        ["Turno", { value: "Tarde", number: false }],
        ["Usuario", { value: "powerfulcontrolsystem@gmail.com", number: false }],
        ["Ventas", { value: money(840000), number: false }],
        ["Anulaciones", { value: money(0), number: false }],
        ["Diferencia", { value: money(0), number: false }]
      ],
      totalLabel: "Caja fisica",
      totalValue: money(840000),
      note: "Reporte listo para guardar cierre. Incluye resumen, pagos y auditoria."
    },
    parqueadero: {
      title: "Recibo parqueadero",
      kind: "recibo",
      badge: format === "pos" ? "POS" : "Carta",
      tableHeaders: ["Campo", "Detalle"],
      rows: [
        ["Ticket", { value: "PK-000184", number: false }],
        ["Placa", { value: "ABC123", number: false }],
        ["Entrada", { value: "2026-05-10 12:15", number: false }],
        ["Salida", { value: "2026-05-10 15:30", number: false }],
        ["Minutos", { value: "195", number: false }],
        ["QR", { value: "TOKEN-PK-000184-VALIDACION-SALIDA", number: false }]
      ],
      totalLabel: "Total",
      totalValue: money(18000),
      note: "Recibo de salida con QR de validacion."
    },
    turno_atencion: {
      title: "Ticket de turno",
      kind: "recibo",
      badge: format === "pos" ? "POS" : "Carta",
      meta: [
        { label: "Turno", value: "T-000" },
        { label: "Servicio", value: "Recepcion" },
        { label: "Puesto", value: "Modulo 1" },
        { label: "Fecha", value: today }
      ],
      totalLabel: "Estado",
      totalValue: "Esperando llamado",
      note: "Espera tu llamado en pantalla."
    }
  };
  return Object.assign({}, common, byKind[kind]);
}

function turnoTicketHTML() {
  return "<!doctype html><html><head><meta charset=\"utf-8\"><title>Turno T-000</title><style>" +
    "@page{size:80mm auto;margin:6mm}body{font-family:Arial,sans-serif;margin:0;color:#111}.ticket{width:72mm;margin:0 auto;text-align:center}.brand{font-size:12px;text-transform:uppercase;font-weight:800;letter-spacing:.08em}.code{font-size:42px;font-weight:900;margin:10px 0}.row{border-top:1px dashed #999;padding:8px 0;font-size:13px}.muted{color:#555;font-size:11px}.screen{margin-top:10px;font-weight:700}@media print{.no-print{display:none}}button{margin-top:14px;padding:10px 14px;border:1px solid #222;background:#fff;border-radius:8px;cursor:pointer}" +
    "</style></head><body><main class=\"ticket\"><div class=\"brand\">Turnos de atencion</div><div class=\"code\">T-000</div><div class=\"row\"><strong>Recepcion</strong></div><div class=\"row\">Cliente: Prueba visual</div><div class=\"row\">Puesto: Modulo 1</div><div class=\"row muted\">2026-05-10 15:30</div><div class=\"screen\">Espera tu llamado en pantalla</div><button class=\"no-print\" onclick=\"window.print()\">Imprimir</button></main><script>setTimeout(function(){window.print()},250)</script></body></html>";
}

function parqueaderoTicketHTML() {
  return "<!doctype html><html><head><meta charset=\"utf-8\"><title>Recibo parqueadero PK-000184</title><style>" +
    "@page{size:80mm auto;margin:4mm}html,body{margin:0 auto;background:#fff;color:#111827;font-family:Arial,sans-serif}.parking-receipt{width:72mm;margin:0 auto;padding:0;border:0;background:white;color:#111827;font-family:Arial,sans-serif}.parking-receipt h3{margin:0 0 8px;text-align:center;color:#111827}.parking-receipt-row{display:flex;justify-content:space-between;gap:10px;padding:5px 0;border-bottom:1px solid #e5e7eb;font-size:13px}.parking-receipt-row strong{overflow-wrap:anywhere;text-align:right}.parking-receipt-total{font-size:20px;font-weight:900;text-align:right;margin-top:10px}.parking-qr{display:grid;place-items:center;margin:14px auto 4px;min-height:96px;font-size:11px;word-break:break-all;text-align:center}" +
    "</style></head><body><div id=\"parkingReceipt\" class=\"parking-receipt\"><h3>Parqueadero</h3><div class=\"parking-receipt-row\"><span>Ticket</span><strong>PK-000184</strong></div><div class=\"parking-receipt-row\"><span>Placa</span><strong>ABC123</strong></div><div class=\"parking-receipt-row\"><span>Entrada</span><strong>2026-05-10 12:15</strong></div><div class=\"parking-receipt-row\"><span>Salida</span><strong>2026-05-10 15:30</strong></div><div class=\"parking-receipt-row\"><span>Minutos</span><strong>195</strong></div><div class=\"parking-receipt-row\"><span>Subtotal</span><strong>$15.126</strong></div><div class=\"parking-receipt-row\"><span>Impuestos</span><strong>$2.874</strong></div><div class=\"parking-receipt-total\">$18.000</div><div class=\"parking-qr\">TOKEN-PK-000184-VALIDACION-SALIDA</div></div><script>setTimeout(function(){window.print()},250)</script></body></html>";
}

async function measure(page) {
  return page.evaluate(() => {
    const doc = document.documentElement;
    const body = document.body;
    const viewportWidth = doc.clientWidth || innerWidth;
    const bodyWidth = Math.max(body.scrollWidth, doc.scrollWidth);
    const visibleText = (body.innerText || "").trim();
    const badNodes = [];
    for (const el of Array.from(body.querySelectorAll("*"))) {
      const rect = el.getBoundingClientRect();
      const style = getComputedStyle(el);
      if (rect.width <= 0 || rect.height <= 0 || style.display === "none" || style.visibility === "hidden") continue;
      const horizontalOverflow = el.scrollWidth > el.clientWidth + 3 && style.overflowX === "visible";
      const outsideViewport = rect.right > viewportWidth + 4;
      if (horizontalOverflow || outsideViewport) {
        badNodes.push({
          tag: el.tagName.toLowerCase(),
          className: String(el.className || "").slice(0, 80),
          text: (el.innerText || el.textContent || "").trim().replace(/\s+/g, " ").slice(0, 120),
          clientWidth: el.clientWidth,
          scrollWidth: el.scrollWidth,
          right: Math.round(rect.right),
          viewportWidth
        });
      }
      if (badNodes.length >= 8) break;
    }
    return {
      title: document.title,
      textLength: visibleText.length,
      viewportWidth,
      bodyWidth,
      horizontalOverflow: bodyWidth > viewportWidth + 8,
      badNodes
    };
  });
}

async function verifyAutoPrint(browser, html) {
  const page = await browser.newPage({ viewport: { width: 420, height: 640 } });
  const instrumented = html.replace(/<head([^>]*)>/i, '<head$1><script>window.__pcsPrintCalls=0;window.__pcsCloseCalls=0;window.print=function(){window.__pcsPrintCalls+=1};window.close=function(){window.__pcsCloseCalls+=1};</script>');
  await page.setContent(instrumented, { waitUntil: "load" });
  await page.waitForTimeout(500);
  const calls = await page.evaluate(() => ({ print: window.__pcsPrintCalls || 0, close: window.__pcsCloseCalls || 0 }));
  await page.close();
  return calls;
}

async function renderCase(browser, item, screenshotsDir, pdfDir, htmlDir) {
  const isPos = item.format === "pos";
  const viewport = isPos ? { width: 340, height: 900 } : { width: 1100, height: 1500 };
  const name = slug(item.name);
  const htmlPath = path.join(htmlDir, name + ".html");
  const screenshotPath = path.join(screenshotsDir, name + ".png");
  const pdfPath = path.join(pdfDir, name + ".pdf");
  fs.writeFileSync(htmlPath, item.html, "utf8");

  const page = await browser.newPage({ viewport });
  const consoleErrors = [];
  page.on("console", (msg) => {
    if (["error", "warning"].includes(msg.type())) consoleErrors.push(msg.text().slice(0, 500));
  });

  await page.setContent(item.html, { waitUntil: "load" });
  await page.emulateMedia({ media: "print" });
  await page.screenshot({ path: screenshotPath, fullPage: true });
  const scrollHeight = await page.evaluate(() => Math.max(document.body.scrollHeight, document.documentElement.scrollHeight));
  if (isPos) {
    const heightMm = Math.min(600, Math.max(120, Math.ceil(scrollHeight * 0.264583 + 12)));
    await page.pdf({ path: pdfPath, width: "80mm", height: heightMm + "mm", printBackground: true });
  } else {
    await page.pdf({ path: pdfPath, format: "Letter", printBackground: true, margin: { top: "12mm", right: "12mm", bottom: "12mm", left: "12mm" } });
  }
  const metrics = await measure(page);
  const stats = fs.statSync(screenshotPath);
  await page.close();

  return {
    name: item.name,
    format: item.format,
    kind: item.kind,
    html: path.relative(ROOT, htmlPath).replace(/\\/g, "/"),
    screenshot: path.relative(ROOT, screenshotPath).replace(/\\/g, "/"),
    pdf: path.relative(ROOT, pdfPath).replace(/\\/g, "/"),
    screenshotBytes: stats.size,
    consoleErrors,
    metrics,
    status: (!metrics.horizontalOverflow && metrics.textLength > 30 && stats.size > 5000 && metrics.badNodes.length === 0 && consoleErrors.length === 0) ? "ok" : "review"
  };
}

function markdownReport(results, printCalls) {
  const lines = [];
  const failures = results.filter((r) => r.status !== "ok");
  lines.push("# QA visual de impresion");
  lines.push("");
  lines.push("- Fecha: " + new Date().toISOString());
  lines.push("- Motor comun: `web/js/print_documents.js`");
  lines.push("- Casos renderizados: " + results.length);
  lines.push("- Casos OK: " + results.filter((r) => r.status === "ok").length);
  lines.push("- Casos a revisar: " + failures.length);
  lines.push("");
  lines.push("## Validacion de llamada a impresion");
  for (const call of printCalls) {
    lines.push("- " + call.name + ": `window.print()` llamado " + call.calls.print + " vez/veces.");
  }
  lines.push("");
  lines.push("## Evidencia");
  for (const r of results) {
    lines.push("- " + (r.status === "ok" ? "OK" : "REVISAR") + " | " + r.name + " | " + r.kind + " | " + r.format + " | captura: `" + r.screenshot + "` | PDF: `" + r.pdf + "`");
    if (r.metrics.horizontalOverflow || r.metrics.badNodes.length || r.consoleErrors.length) {
      lines.push("  - bodyWidth=" + r.metrics.bodyWidth + " viewportWidth=" + r.metrics.viewportWidth + " erroresConsola=" + r.consoleErrors.length);
      for (const n of r.metrics.badNodes) {
        lines.push("  - Nodo con posible desborde: " + n.tag + "." + n.className + " | " + n.text);
      }
    }
  }
  return lines.join("\n") + "\n";
}

async function main() {
  ensureDir(OUT_DIR);
  const screenshotsDir = path.join(OUT_DIR, "screenshots");
  const pdfDir = path.join(OUT_DIR, "pdf");
  const htmlDir = path.join(OUT_DIR, "html");
  ensureDir(screenshotsDir);
  ensureDir(pdfDir);
  ensureDir(htmlDir);

  const PCSPrint = loadPCSPrint();
  const documentKinds = ["factura", "recibo", "comprobante_ingreso", "comprobante_egreso", "orden", "corte_caja", "parqueadero", "turno_atencion"];
  const cases = [];
  for (const kind of documentKinds) {
    for (const format of ["carta", "pos"]) {
      const options = documentOptions(kind, format);
      cases.push({
        name: options.title + " - " + format,
        kind,
        format,
        html: PCSPrint.buildDocument(options)
      });
    }
  }
  cases.push({ name: "Ticket real turnos atencion - pos", kind: "turno_atencion_real", format: "pos", html: turnoTicketHTML() });
  cases.push({ name: "Recibo real parqueadero - pos", kind: "parqueadero_real", format: "pos", html: parqueaderoTicketHTML() });

  const browser = await chromium.launch({ headless: true });
  const results = [];
  for (const item of cases) {
    results.push(await renderCase(browser, item, screenshotsDir, pdfDir, htmlDir));
  }
  const printCalls = [
    { name: "PCSPrint autoPrint factura POS", calls: await verifyAutoPrint(browser, PCSPrint.buildDocument(Object.assign({}, documentOptions("factura", "pos"), { autoPrint: true }))) },
    { name: "Ticket turnos autoPrint", calls: await verifyAutoPrint(browser, turnoTicketHTML()) },
    { name: "Recibo parqueadero autoPrint", calls: await verifyAutoPrint(browser, parqueaderoTicketHTML()) }
  ];
  await browser.close();

  const summary = {
    outDir: OUT_DIR,
    generatedAt: new Date().toISOString(),
    results,
    printCalls
  };
  fs.writeFileSync(path.join(OUT_DIR, "results.json"), JSON.stringify(summary, null, 2), "utf8");
  fs.writeFileSync(path.join(OUT_DIR, "reporte.md"), markdownReport(results, printCalls), "utf8");

  const printFailures = printCalls.filter((item) => !item.calls || item.calls.print < 1);
  const failures = results.filter((r) => r.status !== "ok" || r.metrics.horizontalOverflow || r.metrics.badNodes.length || r.consoleErrors.length);
  console.log(JSON.stringify({
    outDir: OUT_DIR,
    total: results.length,
    ok: results.length - failures.length,
    review: failures.length,
    printFailures: printFailures.length,
    printCalls
  }, null, 2));
  if (failures.length || printFailures.length) process.exitCode = 1;
}

main().catch((err) => {
  console.error(err && err.stack ? err.stack : err);
  process.exit(1);
});
