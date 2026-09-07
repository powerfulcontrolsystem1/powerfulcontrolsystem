(function () {
  "use strict";
  function q(name) { return (new URLSearchParams(window.location.search)).get(name) || ""; }
  function esc(v) { return String(v || "").replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;"); }
  async function j(url, opts) {
    var res = await fetch(url, Object.assign({ credentials: "same-origin" }, opts || {}));
    var text = await res.text();
    var data = {};
    try { data = text ? JSON.parse(text) : {}; } catch (e) { data = { raw: text }; }
    if (!res.ok) throw new Error(data.error || data.message || data.raw || ("HTTP " + res.status));
    return data;
  }
  var empresaId = q("empresa_id") || q("id");
  var base = "/api/public/turnos_atencion?empresa_id=" + encodeURIComponent(empresaId);
  var servicesGrid = document.getElementById("servicesGrid");
  var kioskMsg = document.getElementById("kioskMsg");
  var ticketCode = document.getElementById("ticketCode");
  var ticketInfo = document.getElementById("ticketInfo");
  var printTicketBtn = document.getElementById("printTicketBtn");
  var lastTicket = null;
  function setMsg(text, bad) { kioskMsg.textContent = text || ""; kioskMsg.style.color = bad ? "#ffb4b4" : "#b8d8ff"; }

  if (!empresaId) {
    servicesGrid.innerHTML = '<div class="form-help">Abre este kiosco desde el enlace publico de una empresa.</div>';
    setMsg("Falta empresa_id en el enlace publico.", true);
    if (printTicketBtn) printTicketBtn.disabled = true;
    return;
  }

  function printTicket(item) {
    item = item || {};
    var code = item.codigo_turno || "-";
    var service = item.servicio_nombre || "Servicio";
    var date = item.fecha_emision || new Date().toLocaleString("es-CO");
    var html = '<!doctype html><html><head><meta charset="utf-8"><title>Turno ' + esc(code) + '</title><style>' +
      '@page{size:80mm auto;margin:6mm}body{font-family:Arial,sans-serif;margin:0;color:#111}.ticket{width:72mm;margin:0 auto;text-align:center}.brand{font-size:12px;text-transform:uppercase;font-weight:800;letter-spacing:.08em}.code{font-size:42px;font-weight:900;margin:10px 0}.row{border-top:1px dashed #999;padding:8px 0;font-size:13px}.muted{color:#555;font-size:11px}.screen{margin-top:10px;font-weight:700}@media print{.no-print{display:none}}button{margin-top:14px;padding:10px 14px;border:1px solid #222;background:#fff;border-radius:8px;cursor:pointer}' +
      '</style></head><body><main class="ticket"><div class="brand">Turnos de atencion</div><div class="code">' + esc(code) + '</div><div class="row"><strong>' + esc(service) + '</strong></div><div class="row muted">' + esc(date) + '</div><div class="screen">Espera tu llamado en pantalla</div><button class="no-print" onclick="window.print()">Imprimir</button></main><script>setTimeout(function(){window.print()},250)<\/script></body></html>';
    var win = window.open("", "pcs_turno_print", "width=420,height=640");
    if (!win) {
      setMsg("El navegador bloqueo la impresion. Permite ventanas emergentes.", true);
      return;
    }
    win.document.open();
    win.document.write(html);
    win.document.close();
    win.focus();
  }

  async function emitTicket(serviceId) {
    try {
      var item = await j(base + "&action=emitir_ticket", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ servicio_id: Number(serviceId) })
      });
      lastTicket = item;
      ticketCode.textContent = item.codigo_turno || "-";
      ticketInfo.textContent = (item.servicio_nombre || "Servicio") + " · Espera a que tu turno aparezca en pantalla.";
      if (printTicketBtn) printTicketBtn.disabled = false;
      setMsg("Tu ticket fue emitido correctamente.");
    } catch (e) { setMsg(e.message, true); }
  }

  async function loadServices() {
    try {
      var items = await j(base + "&action=servicios");
      servicesGrid.innerHTML = items.length ? items.map(function (x) {
        return '<button class="service-btn" data-id="' + x.id + '" style="border-color:' + esc(x.color || "#2563eb") + ';"><span>' + esc(x.nombre) + '</span><small>' + esc(x.descripcion || ("Toma el ticket para " + x.nombre.toLowerCase())) + '</small></button>';
      }).join("") : '<div class="form-help">No hay servicios disponibles en este momento.</div>';
    } catch (e) {
      setMsg(e.message, true);
    }
  }

  servicesGrid.addEventListener("click", function (ev) {
    var btn = ev.target.closest(".service-btn");
    if (!btn) return;
    emitTicket(btn.dataset.id);
  });
  if (printTicketBtn) {
    printTicketBtn.addEventListener("click", function () { printTicket(lastTicket); });
  }

  loadServices();
})();
