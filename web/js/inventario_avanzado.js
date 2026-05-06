(function(){
  "use strict";

  var qs = new URLSearchParams(window.location.search);
  var empresaId = qs.get("empresa_id") || localStorage.getItem("empresa_id") || "";
  var api = "/api/empresa/inventario_avanzado";

  function el(id){ return document.getElementById(id); }
  function val(id){ var n = el(id); return n ? n.value.trim() : ""; }
  function num(id){ var n = Number(val(id)); return Number.isFinite(n) ? n : 0; }
  function today(){ return new Date().toISOString().slice(0,10); }
  function plusDays(days){ var d = new Date(); d.setDate(d.getDate()+days); return d.toISOString().slice(0,10); }
  function money(v){ return new Intl.NumberFormat("es-CO",{style:"currency",currency:"COP",maximumFractionDigits:0}).format(Number(v)||0); }
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
  function post(action, payload){
    payload = payload || {};
    payload.action = action;
    payload.empresa_id = Number(empresaId);
    return fetch(url(action), {method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(payload)})
      .then(function(r){ if(!r.ok){ return r.text().then(function(t){ throw new Error(t || "Error"); }); } return r.json(); });
  }

  function loadAll(){
    if (!empresaId) {
      setMsg("Selecciona una empresa para operar inventario avanzado.", "error");
      return Promise.resolve();
    }
    return fetch(url("dashboard")).then(function(r){ return r.json(); }).then(function(d){
      el("kpiLotes").textContent = d.lotes_activos || 0;
      el("kpiReservas").textContent = d.reservas_activas || 0;
      el("kpiVencer").textContent = (d.lotes_por_vencer || 0) + "/" + (d.lotes_vencidos || 0);
      el("kpiValor").textContent = money(d.valor_disponible || 0);
      renderLotes(d.ultimos_lotes || []);
      renderValor(d.valorizacion || []);
      return loadReservas();
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
    return fetch(url("reservas")).then(function(r){ return r.json(); }).then(function(rows){
      var body = el("reservasBody");
      body.innerHTML = "";
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
      });
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
      el("serialLote").value = r.id;
      el("resLote").value = r.id;
      return loadAll();
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function saveSerial(){
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
      el("resSerial").value = r.id;
      return loadAll();
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function saveReserva(){
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
      el("confirmReservaID").value = r.id;
      return loadAll();
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function confirmReserva(){
    return post("confirmar_reserva", {reserva_id:num("confirmReservaID")}).then(function(){
      setMsg("Reserva confirmada y salida registrada", "success");
      return loadAll();
    }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function seed(){
    return post("seed_demo", {}).then(function(r){
      setMsg("Demo cargada #" + r.id, "success");
      return loadAll();
    }).catch(function(err){ setMsg(err.message, "error"); });
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
  [["btnRefresh",loadAll],["btnSeed",seed],["btnSaveLote",saveLote],["btnSaveSerial",saveSerial],["btnSaveReserva",saveReserva],["btnConfirmReserva",confirmReserva]].forEach(function(pair){
    var node = el(pair[0]);
    if (node) node.addEventListener("click", pair[1]);
  });

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
