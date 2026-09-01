(function(){
  "use strict";

  var qs = new URLSearchParams(window.location.search);
  var empresaId = qs.get("empresa_id") || localStorage.getItem("empresa_id") || "";
  var api = "/api/empresa/compras_avanzadas";
  var proveedores = [];
  var productos = [];
  var bodegas = [];
  var detalleItems = [];
  var pageMutationKey = String(Date.now()) + "-" + Math.random().toString(36).slice(2);

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
  function publicError(status, raw){
    var text = String(raw || "").trim();
    if (text && text.charAt(0) === "{") {
      try {
        var data = JSON.parse(text);
        if (data && typeof data.error === "string") text = data.error.trim();
      } catch (_err) {}
    }
    if (!text || text.indexOf("<") !== -1 || text.length > 240) {
      if (status === 401) return "La sesion expiro. Inicia sesion nuevamente.";
      if (status === 403) return "No tienes permiso para operar compras en esta empresa.";
      if (status === 404) return "La informacion solicitada no esta disponible.";
      if (status === 409) return "La operacion ya fue procesada o entra en conflicto con el estado actual.";
      return "No se pudo completar la operacion de compras.";
    }
    return text;
  }
  function readResponse(r){
    return r.text().then(function(raw){
      if (!r.ok) throw new Error(publicError(r.status, raw));
      if (!raw.trim()) return {};
      try { return JSON.parse(raw); } catch (_err) { throw new Error("La respuesta del servidor no es valida."); }
    });
  }
  function idempotencyKey(action, reference){
    var safe = String(reference || pageMutationKey).trim().replace(/[^A-Za-z0-9_.-]+/g, "-");
    var key = "pcs.purchases." + String(empresaId || "0") + "." + action + "." + safe;
    return key.slice(0, 200);
  }
  function post(action, payload, reference){
    payload = payload || {};
    payload.action = action;
    payload.empresa_id = Number(empresaId);
    return fetch(url(action), {
      method: "POST",
      headers: {"Content-Type":"application/json", "Idempotency-Key":idempotencyKey(action, reference)},
      body: JSON.stringify(payload)
    }).then(readResponse);
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

  function productById(id){
    var n = Number(id || 0);
    for (var i = 0; i < productos.length; i += 1) {
      if (Number(productos[i].id) === n) return productos[i];
    }
    return null;
  }
  function productLabel(p){
    var code = p.sku || p.codigo_barras || ("ID-" + p.id);
    return (p.nombre || "Producto") + " (" + code + ")";
  }
  function renderProductoSelects(){
    var options = ['<option value="">Seleccione producto activo</option>'];
    productos.forEach(function(p){
      if (String(p.estado || "activo").toLowerCase() !== "activo") return;
      options.push('<option value="' + escapeHtml(p.id) + '">' + escapeHtml(productLabel(p)) + '</option>');
    });
    document.querySelectorAll(".producto-select").forEach(function(node){
      var current = node.value;
      node.innerHTML = options.join("");
      if (current && productById(current)) node.value = current;
      if (options.length === 1) node.innerHTML = '<option value="">Primero cree un producto</option>';
    });
  }
  function loadProductos(){
    if (!empresaId) return Promise.resolve();
    return fetch("/api/empresa/productos?empresa_id=" + encodeURIComponent(empresaId) + "&estado=activo&limit=500", {credentials:"same-origin"})
      .then(readResponse)
      .then(function(rows){ productos = Array.isArray(rows) ? rows : []; renderProductoSelects(); });
  }
  function renderBodegas(){
    var options = ['<option value="">Seleccione bodega activa</option>'];
    bodegas.forEach(function(b){
      if (String(b.estado || "activo").toLowerCase() !== "activo") return;
      options.push('<option value="' + escapeHtml(b.id) + '">' + escapeHtml((b.nombre || "Bodega") + " (" + (b.codigo || ("ID-" + b.id)) + ")") + '</option>');
    });
    document.querySelectorAll(".bodega-select").forEach(function(node){
      var current = node.value;
      node.innerHTML = options.join("");
      if (current) node.value = current;
    });
  }
  function loadBodegas(){
    if (!empresaId) return Promise.resolve();
    return fetch("/api/empresa/bodegas?empresa_id=" + encodeURIComponent(empresaId), {credentials:"same-origin"})
      .then(readResponse)
      .then(function(rows){ bodegas = Array.isArray(rows) ? rows : []; renderBodegas(); });
  }

  function loadDashboard(){
    if (!empresaId) {
      setMsg("Selecciona una empresa para operar compras avanzadas.", "error");
      return Promise.resolve();
    }
    return fetch(url("dashboard")).then(readResponse).then(function(d){
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
    return fetch(url("detalle", {id:id})).then(readResponse).then(function(d){
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
    detalleItems = Array.isArray(d.items) ? d.items : [];
    detalleItems.forEach(function(x){
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
    var itemSelect = el("recItemID");
    itemSelect.innerHTML = '<option value="">Seleccione item pendiente</option>';
    detalleItems.forEach(function(x){
      var pendiente = Math.max(0, Number(x.cantidad_solicitada || 0) - Number(x.cantidad_recibida || 0));
      if (pendiente <= 0) return;
      var option = document.createElement("option");
      option.value = String(x.id || "");
      option.textContent = (x.producto_nombre || "Item") + " - pendiente " + pendiente;
      itemSelect.appendChild(option);
    });
    if (itemSelect.options.length > 1) {
      itemSelect.selectedIndex = 1;
      fillReceptionItem();
    } else {
      fillReceptionItem();
    }
  }

  function fillReceptionItem(){
    var itemID = num("recItemID");
    var selected = null;
    for (var i = 0; i < detalleItems.length; i += 1) {
      if (Number(detalleItems[i].id) === itemID) { selected = detalleItems[i]; break; }
    }
    el("recProducto").value = selected && selected.producto_id ? String(selected.producto_id) : "";
    el("recOrdenada").value = selected ? Number(selected.cantidad_solicitada || 0) : "";
    el("recRecibida").value = selected ? Math.max(0, Number(selected.cantidad_solicitada || 0) - Number(selected.cantidad_recibida || 0)) : "";
    el("recCosto").value = selected ? Number(selected.costo_estimado || 0) : "";
  }

  function saveRequisicion(){
    var items = [];
    [["1"],["2"]].forEach(function(s){
      var idx = s[0];
      var product = productById(val("itemNombre" + idx));
      if (product) {
        items.push({producto_id:Number(product.id),producto_nombre:product.nombre || "",cantidad_solicitada:num("itemCant"+idx),costo_estimado:num("itemCosto"+idx) || Number(product.costo || 0),unidad:product.unidad_medida || "und",proveedor_sugerido:providerNameFromSelect("itemProv"+idx)});
      }
    });
    if (!items.length || items.some(function(item){ return item.cantidad_solicitada <= 0; })) {
      setMsg("Selecciona al menos un producto y una cantidad mayor a cero.", "error");
      return Promise.resolve();
    }
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
    }}, val("reqCodigo")).then(function(r){
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
    }}, String(num("cotReqID")) + "." + val("cotNumero")).then(function(r){
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
    }}, String(num("aprReqID")) + "." + String(num("aprCotID")) + "." + val("aprDecision")).then(function(r){
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
    if (!num("recReqID") || !num("recItemID") || !num("recBodega") || num("recRecibida") <= 0) {
      setMsg("Selecciona requisicion, item, bodega y una cantidad recibida mayor a cero.", "error");
      return Promise.resolve();
    }
    return post("recepcion", {recepcion:{
      requisicion_id:num("recReqID"),
      bodega_id:num("recBodega"),
      cotizacion_id:num("recCotID"),
      proveedor_id:proveedorID,
      proveedor_nombre:proveedorNombre,
      documento:val("recDocumento"),
      fecha_recepcion:val("recFecha") || today(),
      items:[{
        requisicion_item_id:num("recItemID"),
        producto_nombre:val("recProducto"),
        cantidad_ordenada:num("recOrdenada"),
        cantidad_recibida:num("recRecibida"),
        costo_unitario:num("recCosto"),
        lote:val("recLote"),
        estado_calidad:"aprobado"
      }]
    }}, String(num("recReqID")) + "." + val("recDocumento")).then(function(r){
      setMsg("Recepcion guardada #" + r.id, "success");
      return loadDetalle(num("recReqID")).then(loadDashboard);
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function seed(){
    return post("seed_demo", {}, pageMutationKey).then(function(r){
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
  [["btnRefresh",loadDashboard],["btnSaveReq",saveRequisicion],["btnSaveCot",saveCotizacion],["btnSaveApr",saveAprobacion],["btnSaveRec",saveRecepcion]].forEach(function(pair){
    var node = el(pair[0]);
    if (node) node.addEventListener("click", pair[1]);
  });

  ["reqFecha","cotFecha","recFecha"].forEach(function(id){ if (el(id)) el(id).value = today(); });
  el("recItemID").addEventListener("change", fillReceptionItem);
  ["1","2"].forEach(function(idx){
    el("itemNombre" + idx).addEventListener("change", function(){
      var product = productById(val("itemNombre" + idx));
      if (product && !num("itemCosto" + idx)) el("itemCosto" + idx).value = Number(product.costo || 0);
    });
  });
  el("reqCodigo").value = "REQ-" + Date.now().toString().slice(-6);
  el("cotNumero").value = "COT-" + Date.now().toString().slice(-6);
  el("recDocumento").value = "REC-" + Date.now().toString().slice(-6);
  var proveedoresLink = el("btnProveedores");
  if (proveedoresLink && empresaId) {
    proveedoresLink.href = "/administrar_empresa/administrar_productos.html?view=proveedores&empresa_id=" + encodeURIComponent(empresaId);
  }
  Promise.all([loadProveedores(), loadProductos(), loadBodegas()]).then(loadDashboard).catch(function(err){
    setMsg(err.message || "No se pudieron cargar los catalogos de compras", "error");
    return loadDashboard();
  });
})();
