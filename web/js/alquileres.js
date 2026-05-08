(function () {
  "use strict";

  var state = {
    empresaID: 0,
    tab: "dashboard",
    config: null,
    categorias: [],
    activos: [],
    tarifas: [],
    contratos: [],
    mantenimientos: [],
    ubicaciones: [],
    map: null,
    markers: []
  };

  function byId(id) { return document.getElementById(id); }
  function escapeHTML(value) { return String(value || "").replace(/[&<>\"']/g, function (c) { return ({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#39;"})[c]; }); }
  function formatMoney(value) {
    var currency = (state.config && state.config.moneda) || "COP";
    return new Intl.NumberFormat("es-CO", { style: "currency", currency: currency, maximumFractionDigits: 0 }).format(Number(value || 0));
  }
  function toDateTimeInput(value) { return String(value || "").replace(" ", "T").slice(0, 16); }
  function fromDateTimeInput(value) { return value ? String(value).replace("T", " ") + ":00" : ""; }
  function apiURL(action, extra) {
    var url = "/api/empresa/alquileres?empresa_id=" + encodeURIComponent(state.empresaID) + "&action=" + encodeURIComponent(action);
    if (extra) url += "&" + extra;
    return url;
  }
  function readJSON(response) {
    return response.text().then(function (text) {
      var payload = {};
      try { payload = text ? JSON.parse(text) : {}; } catch (_) { payload = { message: text || ("HTTP " + response.status) }; }
      if (!response.ok) throw new Error(payload.message || payload.error || ("HTTP " + response.status));
      return payload;
    });
  }
  function fetchJSON(action, extra) { return fetch(apiURL(action, extra), { credentials: "same-origin" }).then(readJSON); }
  function loadModulePart(label, loader) {
    return loader().then(function () {
      return null;
    }).catch(function (error) {
      return label + ": " + (error && error.message ? error.message : "no disponible");
    });
  }
  function sendJSON(action, payload) {
    return fetch(apiURL(action), {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload || {})
    }).then(readJSON);
  }
  function setMessage(text, isError) {
    var node = byId("rentalMsg"); if (!node) return;
    node.textContent = text || "";
    node.classList.toggle("is-hidden", !text);
    node.classList.toggle("error", !!text && !!isError);
    node.classList.toggle("success", !!text && !isError);
  }
  function showError(error) { setMessage(error && error.message ? error.message : "No se pudo completar la acción.", true); }

  function resolveEmpresaID() {
    if (window.__resolveEmpresaIdContext) return Number(window.__resolveEmpresaIdContext() || 0);
    try { return Number(new URLSearchParams(window.location.search).get("empresa_id") || 0); } catch (_) { return 0; }
  }

  function renderList(items, fields, actions) {
    if (!items || !items.length) return '<p class="form-help">Sin registros todavía.</p>';
    var head = fields.map(function (f) { return "<th>" + escapeHTML(f.label) + "</th>"; }).join("") + (actions ? "<th>Acciones</th>" : "");
    var rows = items.map(function (item) {
      return "<tr>" + fields.map(function (f) {
        var value = f.format ? f.format(item[f.key], item) : item[f.key];
        return "<td>" + escapeHTML(value) + "</td>";
      }).join("") + (actions ? "<td>" + actions(item) + "</td>" : "") + "</tr>";
    }).join("");
    return '<div class="table-responsive"><table class="data-table"><thead><tr>' + head + '</tr></thead><tbody>' + rows + '</tbody></table></div>';
  }

  function populateSelect(id, items, valueKey, labelBuilder, includeBlank) {
    var select = byId(id); if (!select) return;
    var current = String(select.value || "");
    var html = includeBlank ? ['<option value="">Sin seleccionar</option>'] : [];
    items.forEach(function (item) {
      html.push('<option value="' + escapeHTML(String(item[valueKey] || "")) + '">' + escapeHTML(labelBuilder(item)) + "</option>");
    });
    select.innerHTML = html.join("");
    if (current) select.value = current;
  }

  function ensureUniversalRentalTypeOptions() {
    var options = [
      ["objeto", "Objeto"],
      ["herramienta_electrica", "Herramienta electrica"],
      ["moto", "Moto"],
      ["mobiliario", "Mobiliario"],
      ["sonido_eventos", "Sonido / eventos"],
      ["tecnologia", "Tecnologia"]
    ];
    ["rentalCategoryType", "rentalAssetType"].forEach(function (id) {
      var select = byId(id);
      if (!select) return;
      var existing = {};
      Array.prototype.slice.call(select.options || []).forEach(function (option) {
        existing[String(option.value || "")] = true;
      });
      options.forEach(function (item) {
        if (existing[item[0]]) return;
        var option = document.createElement("option");
        option.value = item[0];
        option.textContent = item[1];
        select.appendChild(option);
      });
    });
  }

  function syncSelects() {
    populateSelect("rentalAssetCategory", state.categorias, "id", function (item) { return item.nombre; }, true);
    populateSelect("rentalRateCategory", state.categorias, "id", function (item) { return item.nombre; }, true);
    populateSelect("rentalContractAsset", state.activos, "id", function (item) { return item.nombre + " · " + (item.estado || ""); }, true);
    populateSelect("rentalMaintenanceAsset", state.activos, "id", function (item) { return item.nombre; }, true);
    populateSelect("rentalLocationAsset", state.activos, "id", function (item) { return item.nombre; }, true);
    populateSelect("rentalContractTariff", state.tarifas, "id", function (item) { return item.nombre; }, true);
    populateSelect("rentalLocationContract", state.contratos, "id", function (item) { return item.codigo + " · " + item.cliente_nombre; }, true);
  }

  function renderDashboard(data) {
    var cards = [
      ["Disponibles", data.activos_disponibles || 0],
      ["Alquilados", data.activos_alquilados || 0],
      ["Reservas pendientes", data.reservas_pendientes || 0],
      ["Vencidos", data.contratos_vencidos || 0],
      ["Devoluciones hoy", data.devoluciones_hoy || 0],
      ["Mantenimientos abiertos", data.mantenimientos_abiertos || 0],
      ["Ingresos del mes", formatMoney(data.ingresos_mes || 0)],
      ["Depósitos retenidos", formatMoney(data.depositos_retenidos || 0)],
      ["Utilización", ((data.utilizacion_promedio || 0).toFixed ? data.utilizacion_promedio.toFixed(1) : data.utilizacion_promedio) + "%"]
    ];
    byId("rentalDashboardCards").innerHTML = cards.map(function (item) {
      return '<article class="card"><p class="form-help" style="margin-bottom:6px;">' + escapeHTML(item[0]) + '</p><strong style="font-size:1.35rem;">' + escapeHTML(item[1]) + "</strong></article>";
    }).join("");
    byId("rentalDueContracts").innerHTML = renderList(data.proximos_vencimientos || [], [
      { key: "codigo", label: "Contrato" },
      { key: "cliente_nombre", label: "Cliente" },
      { key: "activo_nombre", label: "Activo" },
      { key: "fecha_fin_prevista", label: "Fin previsto" },
      { key: "estado", label: "Estado" }
    ]);
    byId("rentalRiskAssets").innerHTML = renderList(data.activos_en_riesgo || [], [
      { key: "nombre", label: "Activo" },
      { key: "tipo_activo", label: "Tipo" },
      { key: "estado", label: "Estado" },
      { key: "sede", label: "Sede" }
    ]);
    byId("rentalByLine").innerHTML = renderList(data.ingresos_por_linea || [], [
      { key: "etiqueta", label: "Línea" },
      { key: "cantidad", label: "Contratos" },
      { key: "monto", label: "Ingreso", format: function (v) { return formatMoney(v); } }
    ]);
    byId("rentalBySede").innerHTML = renderList(data.ingresos_por_sede || [], [
      { key: "etiqueta", label: "Sede" },
      { key: "cantidad", label: "Contratos" },
      { key: "monto", label: "Ingreso", format: function (v) { return formatMoney(v); } }
    ]);
  }

  function renderConfig() {
    var cfg = state.config || {};
    byId("rentalSystemName").value = cfg.nombre_sistema || "";
    byId("rentalCurrency").value = cfg.moneda || "COP";
    byId("rentalAlertHours").value = cfg.alertar_vencimiento_horas || 12;
    byId("rentalDepositBase").value = cfg.deposito_base_sugerido || 0;
    byId("rentalAllowReservations").checked = !!cfg.permitir_reservas;
    byId("rentalAllowGPS").checked = !!cfg.permitir_gps;
    byId("rentalRequireDeposit").checked = !!cfg.requerir_deposito;
    byId("rentalAllowMileage").checked = !!cfg.permitir_kilometraje;
    byId("rentalRequireChecklist").checked = !!cfg.requerir_checklist;
    byId("rentalAllowDelivery").checked = !!cfg.permitir_entrega_domicilio;
  }

  function renderTables() {
    byId("rentalCategoriesTable").innerHTML = renderList(state.categorias, [
      { key: "codigo", label: "Código" },
      { key: "nombre", label: "Categoría" },
      { key: "tipo_activo", label: "Tipo" },
      { key: "estado", label: "Estado" }
    ]);
    byId("rentalAssetsTable").innerHTML = renderList(state.activos, [
      { key: "codigo", label: "Código" },
      { key: "nombre", label: "Activo" },
      { key: "categoria_nombre", label: "Categoría" },
      { key: "sede", label: "Sede" },
      { key: "estado", label: "Estado" },
      { key: "deposito_sugerido", label: "Depósito", format: function (v) { return formatMoney(v); } }
    ]);
    byId("rentalRatesTable").innerHTML = renderList(state.tarifas, [
      { key: "codigo", label: "Código" },
      { key: "nombre", label: "Tarifa" },
      { key: "categoria_nombre", label: "Categoría" },
      { key: "modalidad_cobro", label: "Modalidad" },
      { key: "precio_dia", label: "Precio día", format: function (v) { return formatMoney(v); } },
      { key: "deposito_minimo", label: "Depósito", format: function (v) { return formatMoney(v); } }
    ]);
    byId("rentalContractsTable").innerHTML = renderList(state.contratos, [
      { key: "codigo", label: "Contrato" },
      { key: "cliente_nombre", label: "Cliente" },
      { key: "activo_nombre", label: "Activo" },
      { key: "fecha_inicio", label: "Inicio" },
      { key: "fecha_fin_prevista", label: "Fin" },
      { key: "estado", label: "Estado" },
      { key: "total", label: "Total", format: function (v) { return formatMoney(v); } },
      { key: "saldo_pendiente", label: "Saldo", format: function (v) { return formatMoney(v); } }
    ], function (item) {
      return '<button type="button" class="btn secondary small" data-contract-state="' + item.id + '" data-next-state="en_curso">Entregar</button> ' +
        '<button type="button" class="btn secondary small" data-contract-state="' + item.id + '" data-next-state="devuelto">Devolver</button> ' +
        '<button type="button" class="btn secondary small" data-contract-state="' + item.id + '" data-next-state="vencido">Vencido</button>';
    });
    byId("rentalMaintenanceTable").innerHTML = renderList(state.mantenimientos, [
      { key: "activo_nombre", label: "Activo" },
      { key: "tipo", label: "Tipo" },
      { key: "prioridad", label: "Prioridad" },
      { key: "estado", label: "Estado" },
      { key: "fecha_programada", label: "Programado" },
      { key: "costo_estimado", label: "Costo est.", format: function (v) { return formatMoney(v); } }
    ]);
    byId("rentalLocationsTable").innerHTML = renderList(state.ubicaciones, [
      { key: "activo_nombre", label: "Activo" },
      { key: "contrato_codigo", label: "Contrato" },
      { key: "latitud", label: "Latitud" },
      { key: "longitud", label: "Longitud" },
      { key: "fuente", label: "Fuente" },
      { key: "fecha_registro", label: "Fecha" }
    ]);
  }

  function ensureMap() {
    if (state.map || !window.L || !byId("rentalMap")) return;
    state.map = L.map("rentalMap").setView([4.711, -74.0721], 5);
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      maxZoom: 19,
      attribution: "&copy; OpenStreetMap"
    }).addTo(state.map);
  }

  function renderMap() {
    ensureMap();
    if (!state.map) return;
    state.markers.forEach(function (m) { try { state.map.removeLayer(m); } catch (_) {} });
    state.markers = [];
    var bounds = [];
    state.ubicaciones.forEach(function (item) {
      if (!item.latitud && !item.longitud) return;
      var marker = L.marker([Number(item.latitud), Number(item.longitud)]).addTo(state.map);
      marker.bindPopup("<strong>" + escapeHTML(item.activo_nombre || "Activo") + "</strong><br>" + escapeHTML(item.contrato_codigo || "Sin contrato") + "<br>" + escapeHTML(item.referencia || ""));
      state.markers.push(marker);
      bounds.push([Number(item.latitud), Number(item.longitud)]);
    });
    if (bounds.length) state.map.fitBounds(bounds, { padding: [24, 24] });
  }

  function refreshAll() {
    return Promise.all([
      loadModulePart("Dashboard", function () { return fetchJSON("dashboard").then(renderDashboard); }),
      loadModulePart("Configuración", function () { return fetchJSON("config").then(function (row) { state.config = row || {}; renderConfig(); }); }),
      loadModulePart("Categorias", function () { return fetchJSON("categorias").then(function (rows) { state.categorias = rows || []; }); }),
      loadModulePart("Activos", function () { return fetchJSON("activos").then(function (rows) { state.activos = rows || []; }); }),
      loadModulePart("Tarifas", function () { return fetchJSON("tarifas").then(function (rows) { state.tarifas = rows || []; }); }),
      loadModulePart("Contratos", function () { return fetchJSON("contratos").then(function (rows) { state.contratos = rows || []; }); }),
      loadModulePart("Mantenimientos", function () { return fetchJSON("mantenimientos").then(function (rows) { state.mantenimientos = rows || []; }); }),
      loadModulePart("Ubicaciones", function () { return fetchJSON("ubicaciones").then(function (rows) { state.ubicaciones = rows || []; }); })
    ]).then(function (errors) {
      syncSelects();
      renderTables();
      renderMap();
      errors = errors.filter(Boolean);
      if (errors.length) {
        setMessage("Módulo cargado parcialmente: " + errors.slice(0, 3).join(" | "), true);
      }
    });
  }

  function collectConfigPayload() {
    return {
      empresa_id: state.empresaID,
      nombre_sistema: byId("rentalSystemName").value,
      moneda: byId("rentalCurrency").value,
      alertar_vencimiento_horas: Number(byId("rentalAlertHours").value || 12),
      deposito_base_sugerido: Number(byId("rentalDepositBase").value || 0),
      permitir_reservas: byId("rentalAllowReservations").checked,
      permitir_gps: byId("rentalAllowGPS").checked,
      requerir_deposito: byId("rentalRequireDeposit").checked,
      permitir_kilometraje: byId("rentalAllowMileage").checked,
      requerir_checklist: byId("rentalRequireChecklist").checked,
      permitir_entrega_domicilio: byId("rentalAllowDelivery").checked
    };
  }
  function collectCategoryPayload() {
    return {
      id: Number(byId("rentalCategoryId").value || 0),
      empresa_id: state.empresaID,
      codigo: byId("rentalCategoryCode").value,
      nombre: byId("rentalCategoryName").value,
      tipo_activo: byId("rentalCategoryType").value,
      estado: "activo"
    };
  }
  function collectAssetPayload() {
    return {
      id: Number(byId("rentalAssetId").value || 0),
      empresa_id: state.empresaID,
      codigo: byId("rentalAssetCode").value,
      nombre: byId("rentalAssetName").value,
      categoria_id: Number(byId("rentalAssetCategory").value || 0),
      tipo_activo: byId("rentalAssetType").value,
      marca: byId("rentalAssetBrand").value,
      modelo: byId("rentalAssetModel").value,
      serie: byId("rentalAssetSerie").value,
      placa: byId("rentalAssetPlate").value,
      sede: byId("rentalAssetSede").value,
      estado: byId("rentalAssetStatus").value,
      valor_reposicion: Number(byId("rentalAssetReplacement").value || 0),
      costo_base_hora: Number(byId("rentalAssetBaseCost").value || 0),
      deposito_sugerido: Number(byId("rentalAssetDeposit").value || 0),
      usa_gps: byId("rentalAssetGPS").checked,
      requiere_checklist: byId("rentalAssetChecklist").checked,
      requiere_licencia: byId("rentalAssetLicense").checked
    };
  }
  function collectRatePayload() {
    return {
      id: Number(byId("rentalRateId").value || 0),
      empresa_id: state.empresaID,
      codigo: byId("rentalRateCode").value,
      nombre: byId("rentalRateName").value,
      categoria_id: Number(byId("rentalRateCategory").value || 0),
      modalidad_cobro: byId("rentalRateMode").value,
      precio_base: Number(byId("rentalRateBase").value || 0),
      precio_hora: Number(byId("rentalRateHour").value || 0),
      precio_dia: Number(byId("rentalRateDay").value || 0),
      precio_semana: Number(byId("rentalRateWeek").value || 0),
      precio_mes: Number(byId("rentalRateMonth").value || 0),
      kilometros_incluidos: Number(byId("rentalRateKm").value || 0),
      deposito_minimo: Number(byId("rentalRateDeposit").value || 0),
      estado: byId("rentalRateStatus").value
    };
  }
  function collectContractPayload() {
    return {
      id: Number(byId("rentalContractId").value || 0),
      empresa_id: state.empresaID,
      codigo: byId("rentalContractCode").value,
      tipo_registro: byId("rentalContractType").value,
      activo_id: Number(byId("rentalContractAsset").value || 0),
      cliente_nombre: byId("rentalContractClient").value,
      cliente_documento: byId("rentalContractDocument").value,
      cliente_telefono: byId("rentalContractPhone").value,
      cliente_email: byId("rentalContractEmail").value,
      tarifa_id: Number(byId("rentalContractTariff").value || 0),
      modalidad_cobro: byId("rentalContractMode").value,
      fecha_inicio: fromDateTimeInput(byId("rentalContractStart").value),
      fecha_fin_prevista: fromDateTimeInput(byId("rentalContractEnd").value),
      estado: byId("rentalContractStatus").value,
      dias_planeados: Number(byId("rentalContractDays").value || 0),
      horas_planeadas: Number(byId("rentalContractHours").value || 0),
      kilometros_incluidos: Number(byId("rentalContractKm").value || 0),
      valor_base: Number(byId("rentalContractBase").value || 0),
      deposito: Number(byId("rentalContractDeposit").value || 0),
      impuestos: Number(byId("rentalContractTaxes").value || 0),
      descuento: Number(byId("rentalContractDiscount").value || 0),
      origen_entrega: byId("rentalContractOrigin").value,
      destino_devolucion: byId("rentalContractReturn").value,
      requiere_garantia: byId("rentalContractGuarantee").checked,
      gps_tracking_activo: byId("rentalContractGPS").checked
    };
  }
  function collectMaintenancePayload() {
    return {
      id: Number(byId("rentalMaintenanceId").value || 0),
      empresa_id: state.empresaID,
      activo_id: Number(byId("rentalMaintenanceAsset").value || 0),
      tipo: byId("rentalMaintenanceType").value,
      prioridad: byId("rentalMaintenancePriority").value,
      estado: byId("rentalMaintenanceStatus").value,
      fecha_programada: byId("rentalMaintenanceDate").value,
      proveedor: byId("rentalMaintenanceProvider").value,
      costo_estimado: Number(byId("rentalMaintenanceEstimated").value || 0),
      costo_real: Number(byId("rentalMaintenanceReal").value || 0),
      descripcion: byId("rentalMaintenanceDescription").value
    };
  }
  function collectLocationPayload() {
    return {
      empresa_id: state.empresaID,
      activo_id: Number(byId("rentalLocationAsset").value || 0),
      contrato_id: Number(byId("rentalLocationContract").value || 0),
      fuente: byId("rentalLocationSource").value,
      latitud: Number(byId("rentalLocationLat").value || 0),
      longitud: Number(byId("rentalLocationLng").value || 0),
      referencia: byId("rentalLocationRef").value
    };
  }

  function setTab(tab) {
    state.tab = tab;
    Array.prototype.slice.call(document.querySelectorAll(".rental-tab")).forEach(function (section) {
      section.hidden = section.id !== ("rentalTab-" + tab);
    });
    if (tab === "mapa") setTimeout(renderMap, 120);
  }

  function wireEvents() {
    Array.prototype.slice.call(document.querySelectorAll("[data-rental-tab]")).forEach(function (btn) {
      btn.addEventListener("click", function () { setTab(btn.getAttribute("data-rental-tab")); });
    });
    byId("rentalConfigForm").addEventListener("submit", function (ev) {
      ev.preventDefault();
      sendJSON("config", collectConfigPayload()).then(function () { return refreshAll(); }).then(function () { setMessage("Configuración guardada.", false); }).catch(showError);
    });
    byId("rentalCategoryForm").addEventListener("submit", function (ev) {
      ev.preventDefault();
      sendJSON("categorias", collectCategoryPayload()).then(function () { return refreshAll(); }).then(function () { setMessage("Categoría guardada.", false); }).catch(showError);
    });
    byId("rentalAssetForm").addEventListener("submit", function (ev) {
      ev.preventDefault();
      sendJSON("activos", collectAssetPayload()).then(function () { return refreshAll(); }).then(function () { setMessage("Activo guardado.", false); }).catch(showError);
    });
    byId("rentalRateForm").addEventListener("submit", function (ev) {
      ev.preventDefault();
      sendJSON("tarifas", collectRatePayload()).then(function () { return refreshAll(); }).then(function () { setMessage("Tarifa guardada.", false); }).catch(showError);
    });
    byId("rentalContractForm").addEventListener("submit", function (ev) {
      ev.preventDefault();
      sendJSON("contratos", collectContractPayload()).then(function () { return refreshAll(); }).then(function () { setMessage("Contrato guardado.", false); }).catch(showError);
    });
    byId("rentalMaintenanceForm").addEventListener("submit", function (ev) {
      ev.preventDefault();
      sendJSON("mantenimientos", collectMaintenancePayload()).then(function () { return refreshAll(); }).then(function () { setMessage("Mantenimiento guardado.", false); }).catch(showError);
    });
    byId("rentalLocationForm").addEventListener("submit", function (ev) {
      ev.preventDefault();
      sendJSON("ubicaciones", collectLocationPayload()).then(function () { return refreshAll(); }).then(function () { setMessage("Ubicación registrada.", false); }).catch(showError);
    });
    byId("rentalSeedBtn").addEventListener("click", function () {
      sendJSON("seed_demo", {}).then(function () { return refreshAll(); }).then(function () { setMessage("Preconfiguración operativa aplicada.", false); }).catch(showError);
    });
    byId("rentalUseMyLocation").addEventListener("click", function () {
      if (!navigator.geolocation) {
        setMessage("Este navegador no permite geolocalización.", true);
        return;
      }
      navigator.geolocation.getCurrentPosition(function (position) {
        byId("rentalLocationLat").value = position.coords.latitude.toFixed(8);
        byId("rentalLocationLng").value = position.coords.longitude.toFixed(8);
        byId("rentalLocationRef").value = "GPS del operador";
      }, function () {
        setMessage("No se pudo obtener la ubicación actual.", true);
      }, { enableHighAccuracy: true, timeout: 10000 });
    });
    document.addEventListener("click", function (ev) {
      var btn = ev.target.closest("[data-contract-state]");
      if (btn) {
        sendJSON("cambiar_estado", {
          contrato_id: Number(btn.getAttribute("data-contract-state") || 0),
          estado: btn.getAttribute("data-next-state") || ""
        }).then(function () { return refreshAll(); }).then(function () { setMessage("Estado actualizado.", false); }).catch(showError);
      }
    });
  }

  function init() {
    state.empresaID = resolveEmpresaID();
    if (!state.empresaID) {
      setMessage("No se encontró empresa_id en el contexto.", true);
      return;
    }
    ensureUniversalRentalTypeOptions();
    wireEvents();
    refreshAll().catch(showError);
  }

  init();
})();
