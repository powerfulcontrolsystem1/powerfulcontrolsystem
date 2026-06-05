(function () {
  "use strict";

  var state = {
    empresaId: "",
    camaras: [],
    catalogo: [],
    visores: []
  };

  function byId(id) { return document.getElementById(id); }
  function esc(value) {
    return String(value == null ? "" : value).replace(/[&<>"']/g, function (ch) {
      return {"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[ch];
    });
  }
  function cleanNum(value) {
    var n = Number(value || 0);
    return Number.isFinite(n) ? Math.max(0, Math.trunc(n)) : 0;
  }
  function resolveEmpresaId() {
    var params = new URLSearchParams(window.location.search || "");
    var id = params.get("empresa_id") || params.get("id") || "";
    if (!id && window.__resolveEmpresaIdContext) {
      try { id = window.__resolveEmpresaIdContext() || ""; } catch (e) { id = ""; }
    }
    if (!id) {
      ["active_empresa_id", "empresa_id", "admin_empresa_id"].some(function (key) {
        try { id = sessionStorage.getItem(key) || localStorage.getItem(key) || ""; } catch (e) { id = ""; }
        return !!id;
      });
    }
    state.empresaId = String(id || "").replace(/\D+/g, "");
    return state.empresaId;
  }
  function api(action, options) {
    var url = "/api/empresa/camaras?empresa_id=" + encodeURIComponent(state.empresaId);
    if (action) {
      var actionParts = String(action).split("&");
      url += "&action=" + encodeURIComponent(actionParts.shift() || "");
      if (actionParts.length) url += "&" + actionParts.join("&");
    }
    return fetch(url, options || {}).then(async function (res) {
      var text = await res.text();
      var data = {};
      try { data = text ? JSON.parse(text) : {}; } catch (e) { data = { ok: false, error: text }; }
      if (!res.ok) throw new Error(data.error || text || "Error " + res.status);
      return data;
    });
  }
  function post(action, payload) {
    return api(action, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload || {})
    });
  }
  function setStatus(message, isError) {
    var el = byId("camaraStatus");
    if (!el) return;
    el.hidden = !message;
    el.textContent = message || "";
    el.classList.toggle("is-error", !!isError);
  }
  function fillSelect(el, options, valueKey, labelKey) {
    if (!el) return;
    el.innerHTML = (options || []).map(function (item) {
      var value = typeof valueKey === "function" ? valueKey(item) : item[valueKey];
      var label = typeof labelKey === "function" ? labelKey(item) : item[labelKey];
      return '<option value="' + esc(value) + '">' + esc(label) + "</option>";
    }).join("");
  }
  function load() {
    if (!resolveEmpresaId()) {
      setStatus("No se detecto la empresa activa.", true);
      return Promise.resolve();
    }
    setStatus("Cargando camaras...", false);
    return api("dashboard").then(function (data) {
      state.camaras = Array.isArray(data.camaras) ? data.camaras : [];
      state.catalogo = Array.isArray(data.catalogo) ? data.catalogo : [];
      state.visores = Array.isArray(data.visores) ? data.visores : [];
      fillSelect(byId("camaraProtocolo"), state.catalogo, "clave", "nombre");
      fillSelect(byId("camaraVisor"), state.visores, "clave", "nombre");
      renderKPIs(data.kpis || {});
      renderCamaras();
      renderCatalogo();
      setStatus("", false);
    }).catch(function (err) {
      setStatus(err.message || "No se pudo cargar el modulo de camaras.", true);
    });
  }
  function renderKPIs(kpis) {
    var total = kpis.total != null ? kpis.total : state.camaras.length;
    var activas = kpis.activas != null ? kpis.activas : state.camaras.filter(function (c) { return c.activa; }).length;
    var estaciones = kpis.con_estacion != null ? kpis.con_estacion : 0;
    var web = kpis.visores_web != null ? kpis.visores_web : 0;
    if (byId("camKpiTotal")) byId("camKpiTotal").textContent = total;
    if (byId("camKpiActivas")) byId("camKpiActivas").textContent = activas;
    if (byId("camKpiEstaciones")) byId("camKpiEstaciones").textContent = estaciones;
    if (byId("camKpiWeb")) byId("camKpiWeb").textContent = web;
  }
  function safeBrowserURL(value) {
    var raw = String(value || "").trim();
    var lower = raw.toLowerCase();
    if (!raw || lower.indexOf("javascript:") === 0 || lower.indexOf("data:") === 0) return "";
    if (lower.indexOf("http://") === 0 || lower.indexOf("https://") === 0 || raw.indexOf("/") === 0) return raw;
    return "";
  }
  function previewHtml(cam) {
    var visor = String(cam.visor_tipo || "auto").toLowerCase();
    var protocolo = String(cam.protocolo_origen || "").toLowerCase();
    var embed = safeBrowserURL(cam.url_embed);
    var snapshot = safeBrowserURL(cam.url_snapshot);
    var stream = safeBrowserURL(cam.url_stream);
    if (visor === "iframe" && embed) return '<iframe src="' + esc(embed) + '" loading="lazy" referrerpolicy="no-referrer"></iframe>';
    if ((visor === "mjpeg" || protocolo === "mjpeg") && (snapshot || stream)) return '<img src="' + esc(snapshot || stream) + '" alt="' + esc(cam.nombre || "Camara") + '">';
    if ((visor === "hls" || protocolo === "hls") && (stream || embed)) return '<video controls muted autoplay playsinline src="' + esc(stream || embed) + '"></video>';
    if ((visor === "webrtc" || protocolo === "webrtc") && embed) return '<iframe src="' + esc(embed) + '" loading="lazy" referrerpolicy="no-referrer"></iframe>';
    if (embed) return '<iframe src="' + esc(embed) + '" loading="lazy" referrerpolicy="no-referrer"></iframe>';
    if (snapshot) return '<img src="' + esc(snapshot) + '" alt="' + esc(cam.nombre || "Camara") + '">';
    return '<div><strong>' + esc(String(protocolo || "RTSP").toUpperCase()) + '</strong><br>Configure URL HLS, WebRTC, MJPEG o iframe para vista en navegador.</div>';
  }
  function renderCamaras() {
    var grid = byId("camaraGrid");
    if (!grid) return;
    if (!state.camaras.length) {
      grid.innerHTML = '<article class="camara-card"><p>No hay camaras registradas. Agrega la primera camara o DVR de esta empresa.</p></article>';
      return;
    }
    grid.innerHTML = state.camaras.map(function (cam) {
      var proto = String(cam.protocolo_origen || "rtsp").toUpperCase();
      var estado = cam.activa ? "Activa" : "Inactiva";
      return '<article class="camara-card" data-camara-id="' + esc(cam.id) + '">' +
        '<div class="camara-preview">' + previewHtml(cam) + '</div>' +
        '<div><strong>' + esc(cam.nombre) + '</strong><br><span class="muted">' + esc(cam.ubicacion || cam.dvr_nombre || "Sin ubicacion") + '</span></div>' +
        '<div class="camara-meta">' +
        '<span>Protocolo</span><strong>' + esc(proto) + '</strong>' +
        '<span>Canal</span><strong>' + esc(cam.canal || "-") + '</strong>' +
        '<span>Estacion</span><strong>' + esc(cam.estacion_id || "-") + '</strong>' +
        '<span>Estado</span><strong>' + esc(estado) + '</strong>' +
        '</div>' +
        '<div class="camara-actions">' +
        '<button class="btn secondary" type="button" data-edit="' + esc(cam.id) + '">Editar</button>' +
        '<button class="btn danger" type="button" data-delete="' + esc(cam.id) + '">Desactivar</button>' +
        '</div>' +
        '</article>';
    }).join("");
  }
  function renderCatalogo() {
    var grid = byId("camaraTechGrid");
    if (!grid) return;
    grid.innerHTML = (state.catalogo || []).map(function (item) {
      return '<article class="camara-tech">' +
        '<h3>' + esc(item.nombre || item.clave) + '</h3>' +
        '<p>' + esc(item.uso || "") + '</p>' +
        '<p><strong>Visores:</strong> ' + esc((item.visores || []).join(", ")) + '</p>' +
        '<p>' + esc(item.observacion || "") + '</p>' +
        '</article>';
    }).join("");
  }
  function collect() {
    return {
      id: cleanNum(byId("camaraId") && byId("camaraId").value),
      nombre: byId("camaraNombre") ? byId("camaraNombre").value.trim() : "",
      ubicacion: byId("camaraUbicacion") ? byId("camaraUbicacion").value.trim() : "",
      dvr_nombre: byId("camaraDvrNombre") ? byId("camaraDvrNombre").value.trim() : "",
      dvr_host: byId("camaraDvrHost") ? byId("camaraDvrHost").value.trim() : "",
      canal: byId("camaraCanal") ? byId("camaraCanal").value.trim() : "",
      fabricante: byId("camaraFabricante") ? byId("camaraFabricante").value.trim() : "",
      modelo: byId("camaraModelo") ? byId("camaraModelo").value.trim() : "",
      protocolo_origen: byId("camaraProtocolo") ? byId("camaraProtocolo").value : "rtsp",
      visor_tipo: byId("camaraVisor") ? byId("camaraVisor").value : "auto",
      estacion_id: cleanNum(byId("camaraEstacionId") && byId("camaraEstacionId").value),
      url_stream: byId("camaraUrlStream") ? byId("camaraUrlStream").value.trim() : "",
      url_embed: byId("camaraUrlEmbed") ? byId("camaraUrlEmbed").value.trim() : "",
      url_snapshot: byId("camaraUrlSnapshot") ? byId("camaraUrlSnapshot").value.trim() : "",
      usuario_ref: byId("camaraUsuarioRef") ? byId("camaraUsuarioRef").value.trim() : "",
      password_ref: byId("camaraPasswordRef") ? byId("camaraPasswordRef").value.trim() : "",
      orden: cleanNum(byId("camaraOrden") && byId("camaraOrden").value),
      cargar_en_estaciones: !!(byId("camaraCargarEstaciones") && byId("camaraCargarEstaciones").checked),
      activa: !!(byId("camaraActiva") && byId("camaraActiva").checked),
      estado: byId("camaraActiva") && byId("camaraActiva").checked ? "activo" : "inactivo",
      observaciones: byId("camaraObservaciones") ? byId("camaraObservaciones").value.trim() : ""
    };
  }
  function fillForm(cam) {
    cam = cam || {};
    if (byId("camaraId")) byId("camaraId").value = cam.id || "";
    if (byId("camaraNombre")) byId("camaraNombre").value = cam.nombre || "";
    if (byId("camaraUbicacion")) byId("camaraUbicacion").value = cam.ubicacion || "";
    if (byId("camaraDvrNombre")) byId("camaraDvrNombre").value = cam.dvr_nombre || "";
    if (byId("camaraDvrHost")) byId("camaraDvrHost").value = cam.dvr_host || "";
    if (byId("camaraCanal")) byId("camaraCanal").value = cam.canal || "";
    if (byId("camaraFabricante")) byId("camaraFabricante").value = cam.fabricante || "";
    if (byId("camaraModelo")) byId("camaraModelo").value = cam.modelo || "";
    if (byId("camaraProtocolo")) byId("camaraProtocolo").value = cam.protocolo_origen || "rtsp";
    if (byId("camaraVisor")) byId("camaraVisor").value = cam.visor_tipo || "auto";
    if (byId("camaraEstacionId")) byId("camaraEstacionId").value = cam.estacion_id || 0;
    if (byId("camaraUrlStream")) byId("camaraUrlStream").value = cam.url_stream || "";
    if (byId("camaraUrlEmbed")) byId("camaraUrlEmbed").value = cam.url_embed || "";
    if (byId("camaraUrlSnapshot")) byId("camaraUrlSnapshot").value = cam.url_snapshot || "";
    if (byId("camaraUsuarioRef")) byId("camaraUsuarioRef").value = cam.usuario_ref || "";
    if (byId("camaraPasswordRef")) byId("camaraPasswordRef").value = cam.password_ref || "";
    if (byId("camaraOrden")) byId("camaraOrden").value = cam.orden || 0;
    if (byId("camaraCargarEstaciones")) byId("camaraCargarEstaciones").checked = cam.cargar_en_estaciones !== false;
    if (byId("camaraActiva")) byId("camaraActiva").checked = cam.activa !== false;
    if (byId("camaraObservaciones")) byId("camaraObservaciones").value = cam.observaciones || "";
  }
  function wire() {
    var form = byId("camaraForm");
    if (form) {
      form.addEventListener("submit", function (ev) {
        ev.preventDefault();
        var camara = collect();
        setStatus("Guardando camara...", false);
        post("camara", { camara: camara }).then(function () {
          fillForm({});
          return load();
        }).then(function () {
          setStatus("Camara guardada correctamente.", false);
        }).catch(function (err) {
          setStatus(err.message || "No se pudo guardar la camara.", true);
        });
      });
    }
    if (byId("btnCamaraLimpiar")) byId("btnCamaraLimpiar").addEventListener("click", function () { fillForm({}); });
    if (byId("btnCamarasRefresh")) byId("btnCamarasRefresh").addEventListener("click", load);
    var grid = byId("camaraGrid");
    if (grid) {
      grid.addEventListener("click", function (ev) {
        var edit = ev.target && ev.target.closest("[data-edit]");
        var del = ev.target && ev.target.closest("[data-delete]");
        if (edit) {
          var id = cleanNum(edit.getAttribute("data-edit"));
          var cam = state.camaras.find(function (item) { return cleanNum(item.id) === id; });
          if (cam) {
            fillForm(cam);
            window.scrollTo({ top: 0, behavior: "smooth" });
          }
        }
        if (del) {
          var delID = cleanNum(del.getAttribute("data-delete"));
          if (!delID || !window.confirm("Desactivar esta camara?")) return;
          setStatus("Desactivando camara...", false);
          api("camara&id=" + encodeURIComponent(delID), { method: "DELETE" }).then(load).catch(function (err) {
            setStatus(err.message || "No se pudo desactivar la camara.", true);
          });
        }
      });
    }
  }
  document.addEventListener("DOMContentLoaded", function () {
    wire();
    load();
  });
})();
