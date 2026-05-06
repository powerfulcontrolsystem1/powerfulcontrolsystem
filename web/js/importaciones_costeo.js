(function(){
  var state={empresaID:getEmpresaID(),importaciones:[],detalle:null};
  var ids="btnRefresh btnSeed btnDistribuir btnSaveImportacion btnSaveItem btnSaveCosto kpiAbiertas kpiCerradas kpiNacionalizacion kpiTotal importacionesBody itemsBody costosBody msg impCodigo impProveedor impPais impIncoterm impMoneda impTRM impFecha impRef itemImpID itemNombre itemSKU itemCantidad itemCosto itemPeso itemVol costImpID costTipo costConcepto costBase costValor costCuenta distribuirID".split(" ");
  var els={};ids.forEach(function(id){els[id]=document.getElementById(id);});
  function getEmpresaID(){try{var p=new URLSearchParams(location.search||"");var id=p.get("empresa_id")||p.get("id");if(id)return id;}catch(e){}try{if(parent&&parent.__resolveEmpresaIdContext)return String(parent.__resolveEmpresaIdContext()||"");}catch(e){}return "";}
  function api(action,extra){return "/api/empresa/importaciones_costeo?empresa_id="+encodeURIComponent(state.empresaID)+"&action="+encodeURIComponent(action)+(extra||"");}
  async function req(url,opt){var res=await fetch(url,Object.assign({credentials:"same-origin"},opt||{}));var text=await res.text();var data={};try{data=text?JSON.parse(text):{};}catch(e){data={error:text};}if(!res.ok)throw new Error(data.error||text||("HTTP "+res.status));return data;}
  function esc(v){return String(v==null?"":v).replace(/[&<>"']/g,function(c){return {"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#39;"}[c];});}
  function money(v){try{return new Intl.NumberFormat("es-CO",{style:"currency",currency:"COP",maximumFractionDigits:0}).format(Number(v||0));}catch(e){return "$"+Number(v||0).toFixed(0);}}
  function msg(t,c){els.msg.textContent=t||"";els.msg.className="imp-msg "+(c||"");}
  function today(){return new Date().toISOString().slice(0,10);}
  async function load(){
    if(!state.empresaID){msg("Falta empresa_id en la URL.","error");return;}
    els.impFecha.value=els.impFecha.value||today();
    var d=await req(api("dashboard"));
    state.importaciones=d.ultimas_importaciones||[];
    els.kpiAbiertas.textContent=d.importaciones_abiertas||0;
    els.kpiCerradas.textContent=d.importaciones_cerradas||0;
    els.kpiNacionalizacion.textContent=money(d.costos_pendientes_cop||0);
    els.kpiTotal.textContent=money(d.costo_total_cop||0);
    renderImportaciones();
    if(state.importaciones[0]) await loadDetalle(state.importaciones[0].id);
  }
  function renderImportaciones(){
    els.importacionesBody.innerHTML=state.importaciones.map(function(x){return "<tr data-id='"+esc(x.id)+"'><td>"+esc(x.id)+"</td><td><strong>"+esc(x.codigo)+"</strong></td><td>"+esc(x.proveedor)+"</td><td>"+esc(x.incoterm)+"</td><td>"+money(x.subtotal_cop)+"</td><td>"+money(x.costos_nacionalizacion_cop)+"</td><td>"+money(x.costo_total_cop)+"</td><td>"+esc(x.estado)+"</td></tr>";}).join("")||'<tr><td colspan="8">Sin importaciones.</td></tr>';
    Array.prototype.forEach.call(els.importacionesBody.querySelectorAll("tr[data-id]"),function(row){row.onclick=function(){loadDetalle(Number(row.getAttribute("data-id"))).catch(function(e){msg(e.message,"error");});};});
  }
  async function loadDetalle(id){
    state.detalle=await req(api("detalle","&id="+encodeURIComponent(id)));
    els.itemImpID.value=id;els.costImpID.value=id;els.distribuirID.value=id;
    var items=state.detalle.items||[], costos=state.detalle.costos||[];
    els.itemsBody.innerHTML=items.map(function(x){return "<tr><td>"+esc(x.producto_nombre)+"<br><span class='form-help'>"+esc(x.sku||"")+"</span></td><td>"+esc(x.cantidad)+" "+esc(x.unidad)+"</td><td>"+money(x.costo_base_cop)+"</td><td>"+money(x.costo_distribuido_cop)+"</td><td><strong>"+money(x.costo_unitario_final_cop)+"</strong></td></tr>";}).join("")||'<tr><td colspan="5">Sin items.</td></tr>';
    els.costosBody.innerHTML=costos.map(function(x){return "<tr><td>"+esc(x.tipo)+"</td><td>"+esc(x.concepto)+"</td><td>"+esc(x.base_distribucion)+"</td><td>"+money(x.valor_cop)+"</td><td>"+esc(x.cuenta_contable||"")+"</td></tr>";}).join("")||'<tr><td colspan="5">Sin costos.</td></tr>';
  }
  async function post(action,payload){await req(api(action),{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(payload||{})});await load();}
  els.btnRefresh.onclick=function(){load().then(function(){msg("Datos actualizados.","success");}).catch(function(e){msg(e.message,"error");});};
  els.btnSeed.onclick=function(){post("seed_demo",{}).then(function(){msg("Demo cargado.","success");}).catch(function(e){msg(e.message,"error");});};
  els.btnSaveImportacion.onclick=function(){post("importacion",{codigo:els.impCodigo.value,proveedor:els.impProveedor.value,pais_origen:els.impPais.value,incoterm:els.impIncoterm.value,moneda_origen:els.impMoneda.value,trm:Number(els.impTRM.value||1),fecha_documento:els.impFecha.value,documento_referencia:els.impRef.value,estado:"en_transito"}).then(function(){msg("Importacion guardada.","success");}).catch(function(e){msg(e.message,"error");});};
  els.btnSaveItem.onclick=function(){post("item",{importacion_id:Number(els.itemImpID.value||0),producto_nombre:els.itemNombre.value,sku:els.itemSKU.value,cantidad:Number(els.itemCantidad.value||0),costo_unitario_origen:Number(els.itemCosto.value||0),peso_kg:Number(els.itemPeso.value||0),volumen_m3:Number(els.itemVol.value||0)}).then(function(){msg("Item agregado.","success");}).catch(function(e){msg(e.message,"error");});};
  els.btnSaveCosto.onclick=function(){post("costo",{importacion_id:Number(els.costImpID.value||0),tipo:els.costTipo.value,concepto:els.costConcepto.value,base_distribucion:els.costBase.value,valor_cop:Number(els.costValor.value||0),cuenta_contable:els.costCuenta.value}).then(function(){msg("Costo agregado.","success");}).catch(function(e){msg(e.message,"error");});};
  els.btnDistribuir.onclick=function(){post("distribuir",{importacion_id:Number(els.distribuirID.value||0)}).then(function(){msg("Costos distribuidos.","success");}).catch(function(e){msg(e.message,"error");});};
  load().catch(function(e){msg(e.message,"error");});
})();
