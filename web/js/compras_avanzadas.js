(function(){
  "use strict";

  var qs = new URLSearchParams(window.location.search);
  var empresaId = qs.get("empresa_id") || localStorage.getItem("empresa_id") || "";
  var api = "/api/empresa/compras_avanzadas";
  var proveedores = [];

  function el(id){ return document.getElementById(id); }
  function val(id){ var node = el(id); return node ? node.value.trim() : ""; }
  function num(id){ var n = Number(val(id)); return Number.isFinite(n) ? n : 0; }
  function money(v){ return new Intl.NumberFormat("es-CO",{style:"currency",currency:"COP",maximumFractionDigits:0}).format(Number(v)||0); }
  function today(){ return new Date().toISOString().slice(0,10); }
  function escapeHtml(text){
    var div = document.createElement("div");
    div.textContent = text == null ? "" : String(text);
    return div.innerHTML;
  }
  function setMsg(text, cls){
    var node = el("msg");
    if (!node) return;
    node.className = "cav-msg" + (cls ? " " + cls : "");
    node.textContent = text || "";
  }
  function url(action, extra){
    var p = new URLSearchParams(extra || {});
    p.set("empresa_id", empresaId);
    if (action) p.set("action", action);
    return api + "?" + p.toString();
  }
  function post(action, payload){
    payload = payload || {};
    payload.action = action;
    payload.empresa_id = Number(empresaId);
    return fetch(url(action), {
      method: "POST",
      headers: {"Content-Type":"application/json"},
      body: JSON.stringify(payload)
    }).then(function(r){
      if(!r.ok){ return r.text().then(function(t){ throw new Error(t || "Error"); }); }
      return r.json();
    });
  }
  function providerLabel(p){
    var extra = p.codigo || p.documento || ("ID-" + p.id);
    return (p.nombre || "Proveedor") + " (" + extra + ")";
  }
  function providerById(id){
    var n = Number(id || 0);
    for (var i = 0; i < proveedores.length; i += 1) {
      if (Number(proveedores[i].id) === n) return proveedores[i];
    }
    return null;
  }
  function providerNameFromSelect(id){
    var provider = providerById(val(id));
    return provider ? String(provider.nombre || "").trim() : "";
  }
  function setProviderSelectByNameOrID(selectID, providerID, providerName){
    var node = el(selectID);
    if (!node) return;
    var id = Number(providerID || 0);
    if (id > 0 && providerById(id)) {
      node.value = String(id);
      return;
    }
    var normalized = String(providerName || "").trim().toLowerCase();
    if (!normalized) return;
    for (var i = 0; i < proveedores.length; i += 1) {
      if (String(proveedores[i].nombre || "").trim().toLowerCase() === normalized) {
        node.value = String(proveedores[i].id);
        return;
      }
    }
  }
  function renderProveedorSelects(){
    var options = ['<option value="">Seleccione proveedor creado</option>'];
    proveedores.forEach(function(p){
      if (String(p.estado || "activo").toLowerCase() !== "activo") return;
      options.push('<option value="' + escapeHtml(p.id) + '">' + escapeHtml(providerLabel(p)) + '</option>');
    });
    document.querySelectorAll(".proveedor-select").forEach(function(node){
      var current = node.value;
      node.innerHTML = options.join("");
      if (current && providerById(current)) node.value = current;
      if (options.length === 1) {
        node.innerHTML = '<option value="">Primero cree un proveedor</option>';
      }
    });
  }
  function loadProveedores(){
    if (!empresaId) {
      renderProveedorSelects();
      return Promise.resolve();
    }
    return fetch("/api/empresa/proveedores?empresa_id=" + encodeURIComponent(empresaId) + "&include_inactive=1", {credentials:"same-origin"})
      .then(function(r){
        if(!r.ok){ return r.text().then(function(t){ throw new Error(t || "No se pudieron cargar proveedores"); }); }
        return r.json();
      })
      .then(function(rows){
        proveedores = Array.isArray(rows) ? rows : [];
        renderProveedorSelects();
      });
  }

  function loadDashboard(){
    if (!empresaId) {
      setMsg("Selecciona una empresa para operar compras avanzadas.", "error");
      return Promise.resolve();
    }
    return fetch(url("dashboard")).then(function(r){ return r.json(); }).then(function(d){
      el("kpiAbiertas").textContent = d.requisiciones_abiertas || 0;
      el("kpiAprobacion").textContent = d.requisiciones_pendientes_aprobacion || 0;
      el("kpiCotizaciones").textContent = d.cotizaciones_en_evaluacion || 0;
      el("kpiValor").textContent = money(d.valor_pendiente_aprobacion || 0);
      renderRequisiciones(d.ultimas_requisiciones || []);
    }).catch(function(err){ setMsg(err.message || "No se pudo cargar compras avanzadas", "error"); });
  }

  function renderRequisiciones(rows){
    var body = el("reqBody");
    if (!body) return;
    body.innerHTML = "";
    rows.forEach(function(r){
      var tr = document.createElement("tr");
      tr.innerHTML = "<td></td><td></td><td></td><td></td><td></td><td></td><td><button class='btn secondary small' type='button'>Ver</button></td>";
      tr.children[0].textContent = r.id;
      tr.children[1].textContent = r.codigo || "";
      tr.children[2].textContent = r.area || "";
      tr.children[3].textContent = r.prioridad || "";
      tr.children[4].textContent = money(r.total_estimado || 0);
      tr.children[5].textContent = r.estado_flujo || "";
      tr.querySelector("button").addEventListener("click", function(){ loadDetalle(r.id); });
      body.appendChild(tr);
    });
  }

  function loadDetalle(id){
    return fetch(url("detalle", {id:id})).then(function(r){ return r.json(); }).then(function(d){
      el("cotReqID").value = d.id || "";
      el("aprReqID").value = d.id || "";
      el("recReqID").value = d.id || "";
      renderDetalle(d);
    }).catch(function(err){ setMsg(err.message || "No se pudo cargar detalle", "error"); });
  }

  function renderDetalle(d){
    var body = el("detailBody");
    body.innerHTML = "";
    function add(tipo, ref, nombre, cantidad, valor, estado){
      var tr = document.createElement("tr");
      tr.innerHTML = "<td></td><td></td><td></td><td></td><td></td><td></td>";
      tr.children[0].textContent = tipo;
      tr.children[1].textContent = ref;
      tr.children[2].textContent = nombre;
      tr.children[3].textContent = cantidad || "";
      tr.children[4].textContent = valor ? money(valor) : "";
      tr.children[5].textContent = estado || "";
      body.appendChild(tr);
    }
    (d.items || []).forEach(function(x){
      add("Item", x.id, x.producto_nombre, (x.cantidad_recibida || 0) + " / " + (x.cantidad_solicitada || 0), x.costo_estimado, x.estado);
    });
    (d.cotizaciones || []).forEach(function(x){
      add("Cotizacion", x.id + " - " + x.numero, x.proveedor_nombre, x.tiempo_entrega_dias + " dias", x.total, x.estado);
      if (x.estado === "seleccionada") {
        el("aprCotID").value = x.id;
        el("recCotID").value = x.id;
        setProviderSelectByNameOrID("recProveedor", x.proveedor_id, x.proveedor_nombre);
      }
    });
    (d.aprobaciones || []).forEach(function(x){
      add("Aprobacion", x.id, x.aprobador, "Nivel " + x.nivel, x.monto_autorizado, x.decision);
    });
    (d.recepciones || []).forEach(function(x){
      add("Recepcion", x.id + " - " + x.documento, x.proveedor_nombre, x.estado_recepcion, 0, x.fecha_recepcion);
    });
    if ((d.items || []).length) {
      var first = d.items[0];
      el("recItemID").value = first.id || "";
      el("recProducto").value = first.producto_nombre || "";
      el("recOrdenada").value = first.cantidad_solicitada || "";
      el("recCosto").value = first.costo_estimado || "";
    }
  }

  function saveRequisicion(){
    var items = [];
    [["1"],["2"]].forEach(function(s){
      var idx = s[0];
      var name = val("itemNombre" + idx);
      if (name) {
        items.push({producto_nombre:name,cantidad_solicitada:num("itemCant"+idx),costo_estimado:num("itemCosto"+idx),unidad:"und",proveedor_sugerido:providerNameFromSelect("itemProv"+idx)});
      }
    });
    return post("requisicion", {requisicion:{
      codigo: val("reqCodigo"),
      solicitante: val("reqSolicitante"),
      area: val("reqArea"),
      centro_costo: val("reqCentroCosto"),
      prioridad: val("reqPrioridad"),
      fecha_solicitud: val("reqFecha") || today(),
      fecha_necesidad: val("reqNecesidad"),
      estado_flujo: val("reqEstado") || "solicitada",
      justificacion: val("reqJustificacion"),
      items: items
    }}).then(function(r){
      setMsg("Requisicion guardada #" + r.id, "success");
      el("cotReqID").value = r.id;
      el("aprReqID").value = r.id;
      el("recReqID").value = r.id;
      return loadDashboard();
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function saveCotizacion(){
    var proveedorID = num("cotProveedor");
    var proveedorNombre = providerNameFromSelect("cotProveedor");
    if (!proveedorID || !proveedorNombre) {
      setMsg("Selecciona un proveedor creado para guardar la cotizacion.", "error");
      return Promise.resolve();
    }
    return post("cotizacion", {cotizacion:{
      requisicion_id:num("cotReqID"),
      proveedor_id:proveedorID,
      proveedor_nombre:proveedorNombre,
      numero:val("cotNumero"),
      fecha_cotizacion:val("cotFecha") || today(),
      validez_hasta:val("cotValidez"),
      tiempo_entrega_dias:num("cotEntrega"),
      subtotal:num("cotSubtotal"),
      impuestos:num("cotImpuestos"),
      condiciones_pago:val("cotCondiciones"),
      estado:"evaluacion"
    }}).then(function(r){
      setMsg("Cotizacion guardada #" + r.id, "success");
      el("aprCotID").value = r.id;
      el("recCotID").value = r.id;
      return loadDetalle(num("cotReqID")).then(loadDashboard);
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function saveAprobacion(){
    return post("aprobar", {aprobacion:{
      requisicion_id:num("aprReqID"),
      cotizacion_id:num("aprCotID"),
      decision:val("aprDecision"),
      monto_autorizado:num("aprMonto"),
      comentario:val("aprComentario")
    }}).then(function(r){
      setMsg("Decision registrada #" + r.id, "success");
      return loadDetalle(num("aprReqID")).then(loadDashboard);
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function saveRecepcion(){
    var proveedorID = num("recProveedor");
    var proveedorNombre = providerNameFromSelect("recProveedor");
    if (!proveedorID || !proveedorNombre) {
      setMsg("Selecciona un proveedor creado para guardar la recepcion.", "error");
      return Promise.resolve();
    }
    return post("recepcion", {recepcion:{
      requisicion_id:num("recReqID"),
      cotizacion_id:num("recCotID"),
      proveedor_id:proveedorID,
      proveedor_nombre:proveedorNombre,
      documento:val("recDocumento"),
      fecha_recepcion:val("recFecha") || today(),
      estado_recepcion:val("recEstado"),
      items:[{
        requisicion_item_id:num("recItemID"),
        producto_nombre:val("recProducto"),
        cantidad_ordenada:num("recOrdenada"),
        cantidad_recibida:num("recRecibida"),
        costo_unitario:num("recCosto"),
        estado_calidad:"aprobado"
      }]
    }}).then(function(r){
      setMsg("Recepcion guardada #" + r.id, "success");
      return loadDetalle(num("recReqID")).then(loadDashboard);
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function seed(){
    return post("seed_demo", {}).then(function(r){
      setMsg("Demo cargada #" + r.id, "success");
      return loadDashboard().then(function(){ return loadDetalle(r.id); });
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  document.querySelectorAll(".cav-tab").forEach(function(btn){
    btn.addEventListener("click", function(){
      document.querySelectorAll(".cav-tab").forEach(function(x){ x.classList.remove("is-active"); });
      document.querySelectorAll(".cav-panel").forEach(function(x){ x.classList.remove("is-active"); });
      btn.classList.add("is-active");
      var panel = el(btn.getAttribute("data-panel"));
      if (panel) panel.classList.add("is-active");
    });
  });
  [["btnRefresh",loadDashboard],["btnSeed",seed],["btnSaveReq",saveRequisicion],["btnSaveCot",saveCotizacion],["btnSaveApr",saveAprobacion],["btnSaveRec",saveRecepcion]].forEach(function(pair){
    var node = el(pair[0]);
    if (node) node.addEventListener("click", pair[1]);
  });

  ["reqFecha","cotFecha","recFecha"].forEach(function(id){ if (el(id)) el(id).value = today(); });
  el("reqCodigo").value = "REQ-" + Date.now().toString().slice(-6);
  el("cotNumero").value = "COT-" + Date.now().toString().slice(-6);
  el("recDocumento").value = "REC-" + Date.now().toString().slice(-6);
  var proveedoresLink = el("btnProveedores");
  if (proveedoresLink && empresaId) {
    proveedoresLink.href = "/administrar_empresa/administrar_productos.html?view=proveedores&empresa_id=" + encodeURIComponent(empresaId);
  }
  loadProveedores().then(loadDashboard).catch(function(err){
    setMsg(err.message || "No se pudieron cargar proveedores", "error");
    return loadDashboard();
  });
})();
