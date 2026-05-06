(function(){
  "use strict";
  var qs = new URLSearchParams(window.location.search);
  var empresaId = qs.get("empresa_id") || localStorage.getItem("empresa_id") || "";
  var api = "/api/empresa/crm_avanzado";

  function el(id){ return document.getElementById(id); }
  function val(id){ var n = el(id); return n ? n.value.trim() : ""; }
  function num(id){ var n = Number(val(id)); return Number.isFinite(n) ? n : 0; }
  function month(){ return new Date().toISOString().slice(0,7); }
  function money(v){ return new Intl.NumberFormat("es-CO",{style:"currency",currency:"COP",maximumFractionDigits:0}).format(Number(v)||0); }
  function setMsg(text, cls){ var n = el("msg"); if(!n) return; n.className = "cva-msg" + (cls ? " " + cls : ""); n.textContent = text || ""; }
  function url(action, extra){ var p = new URLSearchParams(extra || {}); p.set("empresa_id", empresaId); if(action) p.set("action", action); return api + "?" + p.toString(); }
  function post(action, payload){
    payload = payload || {};
    payload.action = action;
    payload.empresa_id = Number(empresaId);
    return fetch(url(action), {method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(payload)})
      .then(function(r){ if(!r.ok){ return r.text().then(function(t){ throw new Error(t || "Error"); }); } return r.json(); });
  }

  function load(){
    if(!empresaId){ setMsg("Selecciona una empresa para operar CRM avanzado.", "error"); return Promise.resolve(); }
    var periodo = val("periodo") || month();
    return fetch(url("dashboard", {periodo: periodo})).then(function(r){ return r.json(); }).then(function(d){
      el("kpiPipeline").textContent = money(d.valor_pipeline || 0);
      el("kpiForecast").textContent = money(d.forecast_ponderado || 0);
      el("kpiCotizaciones").textContent = (d.cotizaciones_abiertas || 0) + " / " + money(d.cotizaciones_valor || 0);
      el("kpiMeta").textContent = (d.cumplimiento_meta_pct || 0) + "%";
      renderEmbudo(d.embudo || []);
      renderScores(d.top_leads || []);
      renderAgenda(d.agenda || []);
      renderMetas(d.metas || []);
    }).catch(function(err){ setMsg(err.message || "No se pudo cargar CRM avanzado", "error"); });
  }

  function renderEmbudo(rows){
    var body = el("embudoBody"); body.innerHTML = "";
    rows.forEach(function(x){
      var tr = document.createElement("tr");
      tr.innerHTML = "<td></td><td></td><td></td><td></td><td></td>";
      tr.children[0].textContent = x.estado || "";
      tr.children[1].textContent = x.leads || 0;
      tr.children[2].textContent = money(x.valor || 0);
      tr.children[3].textContent = money(x.forecast || 0);
      tr.children[4].textContent = (x.probabilidad_promedio || 0) + "%";
      body.appendChild(tr);
    });
  }

  function renderScores(rows){
    var body = el("scoresBody"); body.innerHTML = "";
    rows.forEach(function(x){
      var tr = document.createElement("tr");
      tr.innerHTML = "<td></td><td></td><td></td><td></td><td></td><td></td><td><button class='btn secondary small' type='button'>Cotizar</button></td>";
      tr.children[0].textContent = x.id;
      tr.children[1].textContent = (x.codigo || "") + " " + (x.nombre || x.empresa_origen || "");
      tr.children[2].textContent = x.estado_lead || "";
      tr.children[3].textContent = money(x.valor_potencial || 0);
      tr.children[4].textContent = x.score || 0;
      tr.children[5].textContent = x.recomendacion || "";
      tr.querySelector("button").addEventListener("click", function(){ el("leadConvertir").value = x.id || ""; });
      body.appendChild(tr);
    });
  }

  function renderAgenda(rows){
    var body = el("agendaBody"); body.innerHTML = "";
    rows.forEach(function(x){
      var tr = document.createElement("tr");
      tr.innerHTML = "<td></td><td></td><td></td><td></td><td></td><td></td>";
      tr.children[0].textContent = x.tipo || "";
      tr.children[1].textContent = x.referencia || "";
      tr.children[2].textContent = x.nombre || "";
      tr.children[3].textContent = x.responsable || "";
      tr.children[4].textContent = x.fecha || "";
      tr.children[5].textContent = x.estado || "";
      body.appendChild(tr);
    });
  }

  function renderMetas(rows){
    var body = el("metasBody"); body.innerHTML = "";
    rows.forEach(function(x){
      var tr = document.createElement("tr");
      tr.innerHTML = "<td></td><td></td><td></td><td></td><td></td><td></td>";
      tr.children[0].textContent = x.periodo || "";
      tr.children[1].textContent = x.propietario || "";
      tr.children[2].textContent = x.canal || "";
      tr.children[3].textContent = money(x.meta_valor || 0);
      tr.children[4].textContent = x.meta_leads || 0;
      tr.children[5].textContent = (x.meta_conversion_pct || 0) + "%";
      body.appendChild(tr);
    });
  }

  function saveMeta(){
    return post("meta", {meta:{
      periodo: val("metaPeriodo") || month(),
      propietario: val("metaPropietario"),
      canal: val("metaCanal"),
      meta_valor: num("metaValor"),
      meta_leads: num("metaLeads"),
      meta_conversion_pct: num("metaConv"),
      estado: "activo"
    }}).then(function(r){ setMsg("Meta guardada #" + r.id, "success"); return load(); }).catch(function(err){ setMsg(err.message, "error"); });
  }

  function convertLead(){
    return post("cotizacion_desde_lead", {lead_id:num("leadConvertir"), codigo:val("cotCodigo")})
      .then(function(r){ setMsg("Cotizacion creada #" + r.id, "success"); return load(); })
      .catch(function(err){ setMsg(err.message, "error"); });
  }

  function seed(){
    return post("seed_demo", {}).then(function(r){ setMsg("Demo CRM creada #" + r.id, "success"); return load(); }).catch(function(err){ setMsg(err.message, "error"); });
  }

  [["btnRefresh",load],["btnSaveMeta",saveMeta],["btnConvertLead",convertLead],["btnSeed",seed]].forEach(function(pair){ var n = el(pair[0]); if(n) n.addEventListener("click", pair[1]); });
  el("periodo").value = month();
  el("metaPeriodo").value = month();
  el("metaValor").value = "3000000";
  el("metaLeads").value = "15";
  el("metaConv").value = "25";
  el("cotCodigo").value = "COT-CRM-" + Date.now().toString().slice(-6);
  load();
})();
