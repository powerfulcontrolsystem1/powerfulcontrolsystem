(function () {
  var state = {
    items: [],
    total: 0,
    limit: 40,
    offset: 0,
    scope: "-",
    configuredScope: "principal",
    downloadPrefix: "auditoria_global"
  };

  function $(id) {
    return document.getElementById(id);
  }

  function esc(value) {
    return String(value == null ? "" : value).replace(/[&<>"']/g, function (match) {
      return {
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        '"': "&quot;",
        "'": "&#39;"
      }[match];
    });
  }

  function setMsg(text, isError) {
    var el = $("msg");
    if (!el) return;
    el.textContent = text || "";
    el.classList.toggle("error", !!isError);
    el.classList.toggle("success", !!text && !isError);
  }

  function queryParams() {
    var params = new URLSearchParams();
    params.set("scope", state.configuredScope || "principal");
    params.set("limit", String(state.limit));
    params.set("offset", String(state.offset));
    [
      ["desde", "desde"],
      ["hasta", "hasta"],
      ["modulo", "modulo"],
      ["resultado", "resultado"],
      ["usuario", "usuario"],
      ["search", "search"]
    ].forEach(function (pair) {
      var el = $(pair[0]);
      var value = el ? String(el.value || "").trim() : "";
      if (value) params.set(pair[1], value);
    });
    var empresaID = $("empresaId") ? String($("empresaId").value || "").trim() : "";
    if (/^\d+$/.test(empresaID)) params.set("empresa_id", empresaID);
    return params;
  }

  async function fetchJSON(url) {
    var res = await fetch(url, { credentials: "same-origin" });
    var raw = await res.text();
    var data = null;
    try { data = raw ? JSON.parse(raw) : null; } catch (e) { data = null; }
    if (!res.ok) {
      throw new Error((data && (data.error || data.message)) || raw || ("HTTP " + res.status));
    }
    return data;
  }

  function renderKPIs() {
    var users = {};
    var errors = 0;
    state.items.forEach(function (item) {
      if (Number(item.codigo_http || 0) >= 400 || String(item.resultado || "").toLowerCase() !== "ok") {
        errors += 1;
      }
      if (item.usuario_creador) users[String(item.usuario_creador).toLowerCase()] = true;
    });
    $("kpiTotal").textContent = String(state.total || 0);
    $("kpiErrores").textContent = String(errors);
    $("kpiUsuarios").textContent = String(Object.keys(users).length);
    $("kpiScope").textContent = state.scope === "global" ? "Global" : "Mi alcance";
  }

  function renderTable() {
    var tbody = $("tablaAuditoriaGlobal").querySelector("tbody");
    if (!state.items.length) {
      tbody.innerHTML = '<tr><td colspan="8">Sin movimientos para los filtros seleccionados.</td></tr>';
    } else {
      tbody.innerHTML = state.items.map(function (item, index) {
        var statusClass = Number(item.codigo_http || 0) >= 400 ? "audit-bad" : "audit-ok";
        return "<tr>" +
          "<td>" + esc(item.fecha_evento || item.fecha_creacion || "-") + "</td>" +
          "<td>" + esc(item.usuario_creador || "-") + "</td>" +
          "<td>" + esc(item.modulo || "-") + "</td>" +
          "<td>" + esc(item.accion || "-") + "</td>" +
          "<td>" + esc(item.empresa_id ? ("#" + item.empresa_id) : "-") + "</td>" +
          '<td><span class="' + statusClass + '">' + esc((item.resultado || "ok") + " " + (item.codigo_http || "")) + "</span></td>" +
          "<td>" + esc(item.endpoint || "-") + "</td>" +
          '<td><button class="btn secondary small" type="button" data-detail-index="' + index + '">Ver</button></td>' +
          "</tr>";
      }).join("");
    }
    var start = state.total ? state.offset + 1 : 0;
    var end = Math.min(state.offset + state.items.length, state.total);
    $("pagerInfo").textContent = "Mostrando " + start + "-" + end + " de " + state.total;
    $("prevPage").disabled = state.offset <= 0;
    $("nextPage").disabled = state.offset + state.limit >= state.total;
  }

  async function load(reset) {
    if (reset) state.offset = 0;
    setMsg("Consultando auditoría...", false);
    try {
      var data = await fetchJSON("/super/api/auditoria?" + queryParams().toString());
      state.items = Array.isArray(data.items) ? data.items : [];
      state.total = Number(data.total || 0);
      state.scope = data.scope || "-";
      renderKPIs();
      renderTable();
      setMsg("Auditoría actualizada.", false);
    } catch (err) {
      state.items = [];
      state.total = 0;
      renderKPIs();
      renderTable();
      setMsg(err.message || "No se pudo consultar la auditoría.", true);
    }
  }

  function exportRows(format) {
    var rows = state.items.slice();
    if (!rows.length) {
      setMsg("No hay filas para exportar.", true);
      return;
    }
    var content = "";
    var type = "application/json;charset=utf-8";
    var ext = "json";
    if (format === "csv") {
      ext = "csv";
      type = "text/csv;charset=utf-8";
      var headers = ["fecha_evento", "usuario_creador", "principal_email", "empresa_id", "modulo", "accion", "resultado", "codigo_http", "endpoint", "request_id"];
      content = headers.join(",") + "\n" + rows.map(function (row) {
        return headers.map(function (h) {
          return '"' + String(row[h] == null ? "" : row[h]).replace(/"/g, '""') + '"';
        }).join(",");
      }).join("\n");
    } else {
      content = JSON.stringify(rows, null, 2);
    }
    var blob = new Blob([content], { type: type });
    var url = URL.createObjectURL(blob);
    var a = document.createElement("a");
    a.href = url;
    a.download = (state.downloadPrefix || "auditoria_global") + "_" + new Date().toISOString().slice(0, 10) + "." + ext;
    document.body.appendChild(a);
    a.click();
    a.remove();
    setTimeout(function () { URL.revokeObjectURL(url); }, 500);
  }

  function showDetail(index) {
    var item = state.items[index];
    if (!item) return;
    $("detalleJson").textContent = JSON.stringify(item, null, 2);
    var dialog = $("detalleDialog");
    if (dialog && dialog.showModal) {
      dialog.showModal();
    }
  }

  document.addEventListener("DOMContentLoaded", function () {
    state.configuredScope = document.body.dataset.auditScope || "principal";
    state.downloadPrefix = document.body.dataset.auditDownloadPrefix || (state.configuredScope === "super_panel" ? "auditoria_super_admin" : "auditoria_global");
    var moduloDefault = document.body.dataset.auditDefaultModulo || "";
    if (moduloDefault && $("modulo")) $("modulo").value = moduloDefault;
    var btnAgregarAdmin = $("btnAgregarAdmin");
    if (btnAgregarAdmin) {
      btnAgregarAdmin.addEventListener("click", function () {
        window.location.href = "/super/administradores.html?scope=principal";
      });
    }
    $("btnActualizar").addEventListener("click", function () { load(false); });
    $("btnCSV").addEventListener("click", function () { exportRows("csv"); });
    $("btnJSON").addEventListener("click", function () { exportRows("json"); });
    $("filtros").addEventListener("submit", function (ev) {
      ev.preventDefault();
      load(true);
    });
    $("prevPage").addEventListener("click", function () {
      state.offset = Math.max(0, state.offset - state.limit);
      load(false);
    });
    $("nextPage").addEventListener("click", function () {
      if (state.offset + state.limit < state.total) {
        state.offset += state.limit;
        load(false);
      }
    });
    $("cerrarDetalle").addEventListener("click", function () {
      var dialog = $("detalleDialog");
      if (dialog && dialog.close) dialog.close();
    });
    $("tablaAuditoriaGlobal").addEventListener("click", function (ev) {
      var btn = ev.target.closest && ev.target.closest("[data-detail-index]");
      if (!btn) return;
      showDetail(Number(btn.getAttribute("data-detail-index")));
    });
    load(true);
  });
})();
