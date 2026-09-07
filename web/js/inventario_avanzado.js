(function(){
  "use strict";

  var qs = new URLSearchParams(window.location.search);
  var empresaId = qs.get("empresa_id") || localStorage.getItem("empresa_id") || "";
  var api = "/api/empresa/inventario_avanzado";
  var productos = [];
  var bodegas = [];
  var lotes = [];
  var seriales = [];

  function el(id){ return document.getElementById(id); }
  function val(id){ var n = el(id); return n ? n.value.trim() : ""; }
  function num(id){ var n = Number(val(id)); return Number.isFinite(n) ? n : 0; }
  function today(){ return new Date().toISOString().slice(0,10); }
  function plusDays(days){ var d = new Date(); d.setDate(d.getDate()+days); return d.toISOString().slice(0,10); }
  function money(v){ return new Intl.NumberFormat("es-CO",{style:"currency",currency:"COP",maximumFractionDigits:0}).format(Number(v)||0); }
  function idempotencyKey(action, payload){
    var data = payload || {};
    var ref = (data.reserva && data.reserva.origen_ref) || (data.lote && data.lote.lote_codigo) ||
      (data.serial && data.serial.serial) || data.reserva_id || (Date.now() + "-" + Math.random().toString(36).slice(2,12));
    return ["pcs","inventory",empresaId,"advanced",action,String(ref)].join("-").replace(/[^A-Za-z0-9._-]/g,"-").slice(0,200);
  }
  function setMsg(text, cls){
    var node = el("msg");
    if (!node) return;
    node.className = "iav-msg" + (cls ? " " + cls : "");
    node.textContent = text || "";
  }
  function url(action, extra){
    var p = new URLSearchParams(extra || {});
    p.set("empresa_id", empresaId);
    if (action) p.set("action", action);
    return api + "?" + p.toString();
  }
  function readJSONResponse(response){
    return response.text().then(function(text){
      var data = null;
      if (text) {
        try { data = JSON.parse(text); } catch (_) { data = null; }
      }
      if (!response.ok) {
        var message = data && typeof data.error === "string" ? data.error : "No se pudo completar la operación.";
        throw new Error(message);
      }
      return data || {};
    });
  }

  function replaceSelectOptions(selector, placeholder, rows, labelFor, allowed){
    document.querySelectorAll(selector).forEach(function(node){
      var current = node.value;
      node.innerHTML = "";
      var empty = document.createElement("option");
      empty.value = "";
      empty.textContent = placeholder;
      node.appendChild(empty);
      (rows || []).forEach(function(row){
        if (allowed && !allowed(row)) return;
        var option = document.createElement("option");
        option.value = String(row.id || "");
        option.textContent = labelFor(row);
        node.appendChild(option);
      });
      if (current && Array.prototype.some.call(node.options, function(option){ return option.value === current; })) {
        node.value = current;
      }
    });
  }

  function loadProductos(){
    return fetch("/api/empresa/productos?empresa_id=" + encodeURIComponent(empresaId) + "&estado=activo&limit=500")
      .then(readJSONResponse).then(function(rows){
        productos = Array.isArray(rows) ? rows : [];
        replaceSelectOptions(".iav-product-select", "Seleccione producto activo", productos, function(p){
          return (p.nombre || "Producto") + " (" + (p.sku || p.codigo_barras || ("ID-" + p.id)) + ")";
        });
      });
  }

  function loadBodegas(){
    return fetch("/api/empresa/bodegas?empresa_id=" + encodeURIComponent(empresaId))
      .then(readJSONResponse).then(function(rows){
        bodegas = Array.isArray(rows) ? rows : [];
        replaceSelectOptions(".iav-bodega-select", "Seleccione bodega activa", bodegas, function(b){
          return (b.nombre || "Bodega") + " (" + (b.codigo || ("ID-" + b.id)) + ")";
        }, function(b){ return String(b.estado || "activo").toLowerCase() === "activo"; });
      });
  }

  function loadLotes(){
    return fetch(url("lotes", {estado:"activo"})).then(readJSONResponse).then(function(rows){
      lotes = Array.isArray(rows) ? rows : [];
      replaceSelectOptions(".iav-lote-select", "Seleccione lote activo", lotes, function(lote){
        return (lote.lote_codigo || ("Lote #" + lote.id)) + " - " + (lote.producto_nombre || ("Producto #" + lote.producto_id)) + " / " + (lote.bodega_nombre || ("Bodega #" + lote.bodega_id)) + " (libre " + (lote.cantidad_libre || 0) + ")";
      }, function(lote){ return Number(lote.cantidad_libre || 0) > 0; });
    });
  }

  function loadSeriales(){
    return fetch(url("seriales")).then(readJSONResponse).then(function(rows){
      seriales = Array.isArray(rows) ? rows : [];
      replaceSelectOptions(".iav-serial-select", "Sin serial", seriales, function(serial){
        return (serial.serial || ("Serial #" + serial.id)) + " - " + (serial.producto_nombre || ("Producto #" + serial.producto_id));
      }, function(serial){ return String(serial.estado_inventario || "").toLowerCase() === "disponible"; });
    });
  }
  function post(action, payload){
    payload = payload || {};
    payload.action = action;
    payload.empresa_id = Number(empresaId);
    return fetch(url(action), {method:"POST",headers:{"Content-Type":"application/json","Idempotency-Key":idempotencyKey(action,payload)},body:JSON.stringify(payload)})
      .then(readJSONResponse);
  }

  function loadAll(){
    if (!empresaId) {
      setMsg("Selecciona una empresa para operar inventario avanzado.", "error");
      return Promise.resolve();
    }
    var dashboard = fetch(url("dashboard")).then(readJSONResponse);
    return Promise.all([dashboard, loadProductos(), loadBodegas(), loadLotes(), loadSeriales(), loadReservas()]).then(function(values){
      var d = values[0] || {};
      el("kpiLotes").textContent = d.lotes_activos || 0;
      el("kpiReservas").textContent = d.reservas_activas || 0;
      el("kpiVencer").textContent = (d.lotes_por_vencer || 0) + "/" + (d.lotes_vencidos || 0);
      el("kpiValor").textContent = money(d.valor_disponible || 0);
      renderLotes(d.ultimos_lotes || []);
      renderValor(d.valorizacion || []);
    }).catch(function(err){ setMsg(err.message || "No se pudo cargar inventario avanzado", "error"); });
  }

  function renderLotes(rows){
    var body = el("lotesBody");
    body.innerHTML = "";
    rows.forEach(function(x){
      var tr = document.createElement("tr");
      tr.innerHTML = "<td></td><td></td><td></td><td></td><td></td><td></td><td></td><td></td><td><button class='btn secondary small' type='button'>Usar</button></td>";
      tr.children[0].textContent = x.id;
      tr.children[1].textContent = x.producto_nombre || x.producto_id;
      tr.children[2].textContent = x.bodega_nombre || x.bodega_id;
      tr.children[3].textContent = x.lote_codigo || "";
      tr.children[4].textContent = x.cantidad_disponible || 0;
      tr.children[5].textContent = x.cantidad_reservada || 0;
      tr.children[6].textContent = (x.fecha_vencimiento || "") + " " + (x.estado_vencimiento || "");
      tr.children[7].textContent = money(x.valor_disponible || 0);
      tr.querySelector("button").addEventListener("click", function(){
        ["serialLote","resLote"].forEach(function(id){ el(id).value = x.id || ""; });
        ["serialProducto","resProducto"].forEach(function(id){ el(id).value = x.producto_id || ""; });
        ["serialBodega","resBodega"].forEach(function(id){ el(id).value = x.bodega_id || ""; });
        el("resCantidad").value = "1";
      });
      body.appendChild(tr);
    });
  }

  function loadReservas(){
    return fetch(url("reservas")).then(readJSONResponse).then(function(rows){
      var body = el("reservasBody");
      body.innerHTML = "";
      var confirmSelect = el("confirmReservaID");
      var currentReserva = confirmSelect.value;
      confirmSelect.innerHTML = '<option value="">Seleccione reserva activa</option>';
      (rows || []).forEach(function(x){
        var tr = document.createElement("tr");
        tr.innerHTML = "<td></td><td></td><td></td><td></td><td></td><td></td><td></td>";
        tr.children[0].textContent = x.id;
        tr.children[1].textContent = x.producto_nombre || x.producto_id;
        tr.children[2].textContent = x.lote_codigo || x.lote_id || "";
        tr.children[3].textContent = x.serial || x.serial_id || "";
        tr.children[4].textContent = x.cantidad || 0;
        tr.children[5].textContent = x.cliente_nombre || "";
        tr.children[6].textContent = x.estado || "";
        tr.addEventListener("click", function(){ el("confirmReservaID").value = x.id || ""; });
        body.appendChild(tr);
        if (String(x.estado || "").toLowerCase() === "activa") {
          var option = document.createElement("option");
          option.value = String(x.id || "");
          option.textContent = "#" + x.id + " - " + (x.producto_nombre || ("Producto #" + x.producto_id)) + " / " + (x.cliente_nombre || "Sin cliente");
          confirmSelect.appendChild(option);
        }
      });
      if (currentReserva && Array.prototype.some.call(confirmSelect.options, function(option){ return option.value === currentReserva; })) confirmSelect.value = currentReserva;
    });
  }

  function renderValor(rows){
    var body = el("valorBody");
    body.innerHTML = "";
    rows.forEach(function(x){
      var tr = document.createElement("tr");
      tr.innerHTML = "<td></td><td></td><td></td><td></td><td></td><td></td><td></td>";
      tr.children[0].textContent = x.producto_nombre || x.producto_id;
      tr.children[1].textContent = x.bodega_nombre || x.bodega_id;
      tr.children[2].textContent = x.cantidad_disponible || 0;
      tr.children[3].textContent = x.cantidad_reservada || 0;
      tr.children[4].textContent = x.cantidad_libre || 0;
      tr.children[5].textContent = money(x.costo_promedio || 0);
      tr.children[6].textContent = money(x.valor_disponible || 0);
      body.appendChild(tr);
    });
  }

  function saveLote(){
    if (!num("loteProducto") || !num("loteBodega") || !val("loteCodigo") || num("loteCantidad") <= 0) {
      setMsg("Selecciona producto y bodega, e indica lote y cantidad mayor a cero.", "error");
      return Promise.resolve();
    }
    return post("lote", {lote:{
      producto_id:num("loteProducto"),
      bodega_id:num("loteBodega"),
      lote_codigo:val("loteCodigo"),
      fecha_fabricacion:val("loteFabricacion"),
      fecha_vencimiento:val("loteVence"),
      cantidad_inicial:num("loteCantidad"),
      costo_unitario:num("loteCosto"),
      estado_calidad:val("loteCalidad"),
      proveedor:val("loteProveedor"),
      documento_ref:val("loteDocumento"),
      ubicacion_interna:val("loteUbicacion")
    }}).then(function(r){
      setMsg("Lote guardado #" + r.id, "success");
      return loadAll().then(function(){
        el("serialLote").value = r.id;
        el("resLote").value = r.id;
        applyLoteSelection("serialLote", "serialProducto", "serialBodega");
        applyLoteSelection("resLote", "resProducto", "resBodega");
      });
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function saveSerial(){
    if (!num("serialProducto") || !num("serialBodega") || !val("serialCodigo")) {
      setMsg("Selecciona producto y bodega, e indica el serial.", "error");
      return Promise.resolve();
    }
    return post("serial", {serial:{
      lote_id:num("serialLote"),
      producto_id:num("serialProducto"),
      bodega_id:num("serialBodega"),
      serial:val("serialCodigo"),
      estado_inventario:val("serialEstado"),
      fecha_ingreso:val("serialIngreso"),
      garantia_hasta:val("serialGarantia")
    }}).then(function(r){
      setMsg("Serial guardado #" + r.id, "success");
      return loadAll().then(function(){
        el("resSerial").value = r.id;
        applySerialSelection();
      });
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function saveReserva(){
    if (!num("resProducto") || !num("resBodega") || num("resCantidad") <= 0 || !val("resRef")) {
      setMsg("Selecciona producto y bodega, e indica cantidad y referencia.", "error");
      return Promise.resolve();
    }
    return post("reserva", {reserva:{
      producto_id:num("resProducto"),
      bodega_id:num("resBodega"),
      lote_id:num("resLote"),
      serial_id:num("resSerial"),
      cantidad:num("resCantidad"),
      origen_modulo:val("resModulo"),
      origen_ref:val("resRef"),
      cliente_nombre:val("resCliente"),
      fecha_expira:val("resExpira")
    }}).then(function(r){
      setMsg("Reserva creada #" + r.id, "success");
      return loadAll().then(function(){ el("confirmReservaID").value = r.id; });
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function confirmReserva(){
    if (!num("confirmReservaID")) {
      setMsg("Selecciona una reserva activa para confirmar la salida.", "error");
      return Promise.resolve();
    }
    return post("confirmar_reserva", {reserva_id:num("confirmReservaID")}).then(function(){
      setMsg("Reserva confirmada y salida registrada", "success");
      return loadAll();
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function loteByID(id){
    for (var i = 0; i < lotes.length; i += 1) if (Number(lotes[i].id) === Number(id || 0)) return lotes[i];
    return null;
  }

  function serialByID(id){
    for (var i = 0; i < seriales.length; i += 1) if (Number(seriales[i].id) === Number(id || 0)) return seriales[i];
    return null;
  }

  function applyLoteSelection(loteSelectID, productoSelectID, bodegaSelectID){
    var lote = loteByID(val(loteSelectID));
    if (!lote) return;
    el(productoSelectID).value = String(lote.producto_id || "");
    el(bodegaSelectID).value = String(lote.bodega_id || "");
  }

  function applySerialSelection(){
    var serial = serialByID(val("resSerial"));
    if (!serial) return;
    el("resProducto").value = String(serial.producto_id || "");
    el("resBodega").value = String(serial.bodega_id || "");
    if (serial.lote_id) el("resLote").value = String(serial.lote_id);
    el("resCantidad").value = "1";
  }

  document.querySelectorAll(".iav-tab").forEach(function(btn){
    btn.addEventListener("click", function(){
      document.querySelectorAll(".iav-tab").forEach(function(x){ x.classList.remove("is-active"); });
      document.querySelectorAll(".iav-panel").forEach(function(x){ x.classList.remove("is-active"); });
      btn.classList.add("is-active");
      var panel = el(btn.getAttribute("data-panel"));
      if (panel) panel.classList.add("is-active");
    });
  });
  [["btnRefresh",loadAll],["btnSaveLote",saveLote],["btnSaveSerial",saveSerial],["btnSaveReserva",saveReserva],["btnConfirmReserva",confirmReserva]].forEach(function(pair){
    var node = el(pair[0]);
    if (node) node.addEventListener("click", pair[1]);
  });
  el("serialLote").addEventListener("change", function(){ applyLoteSelection("serialLote", "serialProducto", "serialBodega"); });
  el("resLote").addEventListener("change", function(){ applyLoteSelection("resLote", "resProducto", "resBodega"); });
  el("resSerial").addEventListener("change", applySerialSelection);

  el("loteCodigo").value = "LOT-" + Date.now().toString().slice(-6);
  el("serialCodigo").value = "SER-" + Date.now().toString().slice(-6);
  el("resRef").value = "RSV-" + Date.now().toString().slice(-6);
  el("loteFabricacion").value = plusDays(-30);
  el("loteVence").value = plusDays(90);
  el("serialIngreso").value = today();
  el("serialGarantia").value = plusDays(365);
  el("resExpira").value = plusDays(2);
  loadAll();
})();
