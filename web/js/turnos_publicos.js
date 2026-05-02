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
  function setMsg(text, bad) { kioskMsg.textContent = text || ""; kioskMsg.style.color = bad ? "#ffb4b4" : "#b8d8ff"; }

  async function emitTicket(serviceId) {
    try {
      var item = await j(base + "&action=emitir_ticket", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ servicio_id: Number(serviceId) })
      });
      ticketCode.textContent = item.codigo_turno || "-";
      ticketInfo.textContent = (item.servicio_nombre || "Servicio") + " · Espera a que tu turno aparezca en pantalla.";
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

  loadServices();
})();
