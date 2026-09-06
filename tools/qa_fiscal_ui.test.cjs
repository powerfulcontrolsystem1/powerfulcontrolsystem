const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const dir = path.join(__dirname, '../web/administrar_empresa');
function read(page) {return fs.readFileSync(path.join(dir,page), 'utf8').replace(/\r\n/g,'\n');}
function extract(source,name) {
  const start = source.indexOf('    function '+name+'(');
  assert.ok(start>=0,name);
  const end = source.indexOf('\n    function ', start+1);
  assert.ok(end>start,name);
  return source.slice(start,end);
}
test('modified fiscal and offline pages have valid inline JavaScript',()=>{
  for (const page of ['carrito_de_compras.html','facturacion_electronica.html','facturacion_electronica_pruebas_dian.html']) {
    for (const match of read(page).matchAll(/<script\b([^>]*)>([\s\S]*?)<\/script>/gi)) {
      if (!/\bsrc\s*=|application\/ld\+json/i.test(match[1])) new vm.Script(match[2],{filename:page});
    }
  }
});
test('calendar authorization dates do not shift with browser timezone',()=>{
  const source=read('facturacion_electronica_pruebas_dian.html');
  const context=vm.createContext({txt:v=>String(v||'').trim()});
  vm.runInContext(extract(source,'portalDateValue')+extract(source,'asDateTimeLocal'),context);
  assert.equal(context.portalDateValue('2026-06-17'),'17/6/2026');
  assert.equal(context.portalDateValue('2028-06-17'),'17/6/2028');
  assert.equal(context.asDateTimeLocal('2019-03-14'),'2019-03-14T00:00');
  assert.equal(context.portalDateValue(''),'No registrado');
});
test('same range with distinct counters warns instead of implying production completeness',()=>{
  const fields={dian_prefijo:'1PCS',adv_prefijo_factura:'1PCS',dian_resolucion_numero:'QA',adv_resolucion_numero:'QA',dian_consecutivo_actual:'12',adv_proximo_consecutivo:'9'};
  const notice={style:{}};
  const context=vm.createContext({document:{getElementById:id=>id==='fiscalNumberingConsistency'?notice:{value:fields[id]}}});
  vm.runInContext(extract(read('facturacion_electronica.html'),'renderFiscalNumberingConsistency'),context);
  context.renderFiscalNumberingConsistency();
  assert.equal(notice.hidden,false); assert.match(notice.textContent,/12.*9/);
  fields.adv_proximo_consecutivo='12'; context.renderFiscalNumberingConsistency();
  assert.equal(notice.hidden,true);
});
test('Colombia contingency UI is explicit, tenant scoped and type-03 fail closed',()=>{
  const source=read('facturacion_electronica.html');
  for (const marker of [
    'El comprobante offline no es una factura de contingencia.',
    '/api/empresa/facturacion_electronica/contingencias?empresa_id=',
    'ACTIVAR CONTINGENCIA',
    'RECUPERAR SERVICIO',
    'CERRAR CONTINGENCIA',
    'REGISTRAR TALONARIO',
    'transmisi&oacute;n tipo 03 permanece pendiente',
    'Array.isArray(data && data.documentos)',
    'cell.textContent = value'
  ]) assert.ok(source.includes(marker),marker);
});
