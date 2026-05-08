(function () {
  "use strict";

  function getQueryParam(name) {
    return new URLSearchParams(window.location.search || "").get(name) || "";
  }

  function parsePositiveInt(raw) {
    var n = Number(String(raw || "").trim());
    if (!Number.isFinite(n)) return 0;
    n = Math.trunc(n);
    return n > 0 ? n : 0;
  }

  function getEmpresaId() {
    var direct = parsePositiveInt(getQueryParam("empresa_id") || getQueryParam("id"));
    if (direct > 0) return String(direct);
    var keys = ["active_empresa_id", "empresa_id", "admin_empresa_id"];
    var stores = [];
    try { stores.push(window.sessionStorage); } catch (error) {}
    try { stores.push(window.localStorage); } catch (error) {}
    for (var s = 0; s < stores.length; s += 1) {
      for (var i = 0; i < keys.length; i += 1) {
        var val = parsePositiveInt(stores[s] && stores[s].getItem(keys[i]));
        if (val > 0) return String(val);
      }
    }
    return "";
  }

  function isoDate(date) {
    var y = date.getFullYear();
    var m = String(date.getMonth() + 1).padStart(2, "0");
    var d = String(date.getDate()).padStart(2, "0");
    return y + "-" + m + "-" + d;
  }

  function addDays(date, days) {
    var copy = new Date(date.getTime());
    copy.setDate(copy.getDate() + days);
    return copy;
  }

  function esc(value) {
    return String(value == null ? "" : value)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  var empresaId = getEmpresaId();
  var els = {
    desde: document.getElementById("desde"),
    hasta: document.getElementById("hasta"),
    btnActualizar: document.getElementById("btnActualizar"),
    usuarioResumen: document.getElementById("usuarioResumen"),
    mensaje: document.getElementById("mensaje"),
    lista: document.getElementById("listaHorarios"),
    kpiTurnos: document.getElementById("kpiTurnos"),
    kpiHoy: document.getElementById("kpiHoy"),
    kpiProximos: document.getElementById("kpiProximos"),
    kpiHoras: document.getElementById("kpiHoras")
  };

  function setMessage(text, isError) {
    if (!els.mensaje) return;
    els.mensaje.textContent = text || "";
    els.mensaje.style.color = isError ? "var(--danger)" : "var(--muted)";
  }

  async function fetchJSON(url) {
    var res = await fetch(url, { credentials: "same-origin" });
    var text = await res.text();
    var data = {};
    if (text) {
      try { data = JSON.parse(text); } catch (error) { data = { raw: text }; }
    }
    if (!res.ok) {
      throw new Error((data && (data.message || data.error || data.raw)) || ("HTTP " + res.status));
    }
    return data;
  }

  function renderKpis(resumen) {
    resumen = resumen || {};
    els.kpiTurnos.textContent = String(resumen.turnos || 0);
    els.kpiHoy.textContent = String(resumen.turnos_hoy || 0);
    els.kpiProximos.textContent = String(resumen.proximos || 0);
    els.kpiHoras.textContent = String(resumen.horas || 0);
  }

  function renderUser(usuario) {
    usuario = usuario || {};
    var nombre = usuario.nombre || usuario.email || "Usuario";
    var rol = usuario.rol ? " · " + usuario.rol : "";
    els.usuarioResumen.textContent = nombre + rol + ". Consulta aqui tus turnos publicados por la empresa.";
  }

  function renderItems(items) {
    items = Array.isArray(items) ? items : [];
    if (!items.length) {
      els.lista.innerHTML = '<div class="mi-horario-empty">No tienes turnos publicados en el rango seleccionado.</div>';
      return;
    }
    els.lista.innerHTML = items.map(function (item) {
      var color = item.color || "var(--accent)";
      return '<article class="mi-horario-turno" style="border-left-color:' + esc(color) + '">' +
        '<header><div><h3>' + esc(item.turno_nombre || item.tipo_turno || "Turno") + '</h3><time>' + esc(item.fecha || "") + '</time></div>' +
        '<span class="mi-horario-chip">' + esc(item.estado || "publicado") + '</span></header>' +
        '<div class="mi-horario-meta">' +
        '<span class="mi-horario-chip">' + esc(item.hora_inicio || "") + ' - ' + esc(item.hora_fin || "") + '</span>' +
        '<span class="mi-horario-chip">' + esc(item.horas_programadas || 0) + ' h</span>' +
        '<span class="mi-horario-chip">Descanso ' + esc(item.descanso_minutos || 0) + ' min</span>' +
        '</div>' +
        '<p>' + esc((item.area || "Sin area") + " / " + (item.sede || "Sin sede")) + '</p>' +
        (item.observaciones ? '<p>' + esc(item.observaciones) + '</p>' : '') +
        '</article>';
    }).join("");
  }

  async function load() {
    if (!empresaId) {
      setMessage("No se pudo resolver la empresa activa.", true);
      return;
    }
    setMessage("Cargando horario publicado...");
    var url = "/api/empresa/mi_horario?empresa_id=" + encodeURIComponent(empresaId) +
      "&desde=" + encodeURIComponent(els.desde.value || "") +
      "&hasta=" + encodeURIComponent(els.hasta.value || "");
    try {
      var data = await fetchJSON(url);
      renderUser(data.usuario);
      renderKpis(data.resumen);
      renderItems(data.items);
      setMessage(data.resumen && data.resumen.actualizado_en ? "Actualizado: " + data.resumen.actualizado_en : "");
    } catch (error) {
      renderKpis({});
      renderItems([]);
      setMessage(error.message || "No se pudo cargar tu horario.", true);
    }
  }

  function init() {
    var today = new Date();
    els.desde.value = isoDate(today);
    els.hasta.value = isoDate(addDays(today, 14));
    els.btnActualizar.addEventListener("click", load);
    load();
  }

  init();
})();
