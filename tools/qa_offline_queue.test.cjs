const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../web/administrar_empresa/carrito_de_compras.html'), 'utf8').replace(/\r\n/g, '\n');
function extract(name) {
  const pattern = new RegExp('    (?:async )?function ' + name + '\\(');
  const match = pattern.exec(source);
  assert.ok(match, name);
  const tail = source.slice(match.index);
  const end = tail.slice(1).search(/\n    (?:async )?function /);
  assert.ok(end > 0, name + ' end');
  return tail.slice(0, end + 1);
}
function harness() {
  const storage = new Map();
  let lockTail = Promise.resolve();
  const context = vm.createContext({
    state:{empresaID:12,offlineVentas:{syncing:false}},
    navigator:{onLine:true,locks:{request:(_name,action)=>{const next=lockTail.then(action); lockTail=next.catch(()=>{}); return next;}}},
    localStorage:{getItem:key=>storage.get(key)||null,setItem:(key,value)=>storage.set(key,value),removeItem:key=>storage.delete(key)},
    currentOfflineOperatorEmail:()=> 'qa@example.invalid',
    normalize:value=>String(value||'').trim(),
    offlineQueueKey:()=> 'pcs-qa-12-operator',offlineLegacyQueueKey:()=> 'pcs-qa-12',
    isOfflineBillingEnabled:()=>true,
    fetchWithTimeout:async()=>({ok:true}), ensureOk:async response=>{if(!response.ok)throw new Error('network failed');},
    setStationPaymentPersistentMessage:()=>{},showOnlineFloatingNotice:()=>{},
    loadCarritos:async()=>{},loadItems:async()=>{},loadActiveCashRegisters:async()=>{}
  });
  for(const name of ['offlineRowBelongsToCurrentOperator','readOfflineQueueKey','writeOfflineQueueKey','loadOfflineQueue','saveOfflineQueue','withOfflineQueueLock','enqueueOfflineSale','syncOfflineSalesQueue']) vm.runInContext(extract(name), context);
  return {context,storage};
}
function row(key) {return {sync_key:key,empresa_id:12,usuario_email:'qa@example.invalid',estado_local:'pendiente'};}
test('quota failure is not reported as a saved sale', async()=>{
  const {context}=harness(); context.localStorage.setItem=()=>{throw new Error('QuotaExceededError');};
  await assert.rejects(context.enqueueOfflineSale(row('A')), /No se pudo guardar/);
});
test('unreadable queue is preserved, never replaced with empty data',async()=>{
  const {context,storage}=harness(); storage.set('pcs-qa-12-operator','BROKEN-JSON');
  await assert.rejects(context.enqueueOfflineSale(row('A')), /No se pudo leer/);
  assert.equal(storage.get('pcs-qa-12-operator'),'BROKEN-JSON');
});
test('unknown operator or another company cannot enter current queue',async()=>{
  const {context}=harness();
  for(const value of [{...row('A'),empresa_id:13},{...row('A'),usuario_email:''},{...row('A'),usuario_email:'other@example.invalid'}]) await assert.rejects(context.enqueueOfflineSale(value),/empresa y operador/);
  assert.equal(context.loadOfflineQueue().length,0);
});
test('32 parallel saves and duplicates preserve all distinct sales',async()=>{
  const {context}=harness();
  await Promise.all(Array.from({length:32},(_,i)=>context.enqueueOfflineSale(row('A'+i))));
  await context.enqueueOfflineSale(row('A0'));
  assert.equal(context.loadOfflineQueue().length,32);
});
test('legacy operator-key collisions never erase another operator sale',async()=>{
  const {context,storage}=harness();
  const foreign={...row('OTHER'),usuario_email:'other@example.invalid'};
  storage.set('pcs-qa-12-operator',JSON.stringify([foreign]));
  await context.enqueueOfflineSale(row('A'));
  assert.equal(context.loadOfflineQueue().length,1);
  assert.equal(JSON.parse(storage.get('pcs-qa-12-operator')).length,2);
});
test('synchronization cannot delete a sale captured while waiting for HTTP',async()=>{
  const {context}=harness(); await context.enqueueOfflineSale(row('A'));
  context.fetchWithTimeout=async()=>{await context.enqueueOfflineSale(row('B'));return {ok:true};};
  await context.syncOfflineSalesQueue();
  assert.deepEqual(Array.from(context.loadOfflineQueue(),item=>item.sync_key),['B']);
});
test('failed network request keeps the original idempotency key',async()=>{
  const {context}=harness(); await context.enqueueOfflineSale(row('A'));
  context.fetchWithTimeout=async()=>{throw new Error('offline');};
  await context.syncOfflineSalesQueue(); assert.equal(context.loadOfflineQueue()[0].sync_key,'A');
});
test('no unsafe offline fallback when cross-tab locks are unavailable',async()=>{
  const {context}=harness(); context.navigator.locks=undefined;
  await assert.rejects(context.enqueueOfflineSale(row('A')),/cola offline segura/);
});
test('offline receipt cannot precede durable local queue insertion',()=>{
  const body=extract('storeAndPrintOfflineSale');
  assert.ok(body.indexOf('await enqueueOfflineSale') < body.indexOf('await printOfflineSaleReceipt'));
  assert.match(body,/if \(!await enqueueOfflineSale\(payload\)\) throw/);
});
