// Local UI contract: real shared drawer/assets; deterministic APIs and dictation.
// No provider, business database or external mutation is used.
const fs = require('node:fs');
const path = require('node:path');
const http = require('node:http');
const assert = require('node:assert/strict');
const { chromium } = require(path.join(process.env.USERPROFILE, '.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules/playwright'));
const root = path.resolve(__dirname, '..');
const out = path.join(root, 'test_runs/ai_chat_20260906');
fs.mkdirSync(out, { recursive: true });
let history = [];
let submits = 0;
const confirmKeys=[];
const html = `<!doctype html><html lang="es"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="stylesheet" href="/estilos.css"><title>PCS · Verificación local del chat</title><body><main style="padding:32px"><h1>Panel empresarial</h1><p>Verificación local · Agente PCS</p><p>Ventas · Inventario · Compras · Reportes</p></main><script>window.__pcsAutoInjectChatShell=true;window.__pcsEnterpriseAIChromeAllowed=true;window.SpeechRecognition=class{start(){this.onstart?.();setTimeout(()=>{const r=[{transcript:'Agrega dos cervezas a la habitación uno'}];r.isFinal=true;this.onresult?.({resultIndex:0,results:[r]});this.onend?.();},30)}stop(){this.onend?.()}abort(){this.onend?.()}};</script><script src="/js/ai_chat_drawer.js"></script></body></html>`;
const server = http.createServer(async (req, res) => {
  const url = new URL(req.url, 'http://localhost');
  if (url.pathname === '/administrar_empresa/qa_chat.html') { res.setHeader('Content-Type', 'text/html; charset=utf-8'); res.end(html); return; }
  if (url.pathname.startsWith('/api/') || url.pathname === '/me') {
    let body = ''; for await (const c of req) body += c;
    let data = { ok: true, chat_enabled: true, ai_chat_enabled: true, robot_enabled: false };
    if (url.pathname.endsWith('/modelos')) data = { ok: true, modelo_preferido: 'openai:gpt-5.6-terra', modelos: [{id:'openai:gpt-5.6-terra',display_name:'Modelo global · PCS',usage:{daily_used:0,daily_limit:100,daily_remaining:100}}] };
    if (url.pathname.endsWith('/historial')) data = {ok:true,history_scope:url.searchParams.get('scope')||'usuario',can_view_company_history:true,items:history};
    if (url.pathname.endsWith('/consultar')) {
      submits++;
      const input=JSON.parse(body);
      assert.equal(input.agent_id,'agente_pcs'); assert.equal(input.modo_agente,true);
      data={ok:true,model_id:'openai:gpt-5.6-terra',conversation_id:input.conversation_id,respuesta:'Para agregar un consumo:\n\n1. Abre **Estaciones** y selecciona la habitación.\n2. Busca el producto e indica la cantidad.\n3. Revisa la cuenta y confirma.\n\nAgregar productos no cobra ni cierra la cuenta.'};
      if (!/expl[ií]ca/i.test(input.pregunta)) data.enterprise_proposals=[{proposal_id:'proposal_test',plan_hash:'test_hash',tool_name:'sales.add_station_product',resumen:'Agregar 2 × Cerveza a Habitación 1. Precio unitario: 5.000 COP. No cobra ni cierra la cuenta.',risk_level:'medium'}];
      history.unshift({id:submits,conversation_id:input.conversation_id,usuario_creador:'usuario-prueba',pregunta:input.pregunta,respuesta:data.respuesta,fecha_consulta:'2026-09-06 12:00:00'});
    }
    if(url.searchParams.get('action')==='confirm') { confirmKeys.push(JSON.parse(body).idempotency_key);data={ok:true,status:'completed',result:{verified:true}};if(confirmKeys.length===1){res.statusCode=409;data={ok:false,error:'Confirmación temporalmente no disponible'};} }
    res.setHeader('Content-Type','application/json');res.end(JSON.stringify(data));return;
  }
  const target=path.resolve(root,'web','.'+url.pathname);
  if (!target.startsWith(path.join(root,'web')+path.sep)||!fs.existsSync(target)||!fs.statSync(target).isFile()){res.statusCode=404;res.end();return;}
  res.setHeader('Content-Type',target.endsWith('.css')?'text/css':target.endsWith('.js')?'text/javascript':target.endsWith('.svg')?'image/svg+xml':'image/png');
  fs.createReadStream(target).pipe(res);
});
(async()=>{
  await new Promise(resolve=>server.listen(0,'127.0.0.1',resolve));
  const browser=await chromium.launch({headless:true,executablePath:'C:/Program Files/Google/Chrome/Application/chrome.exe'});
  const page=await browser.newPage({viewport:{width:1440,height:1000}});
  const errors=[];page.on('pageerror',e=>{errors.push(e.message);process.stderr.write(e.stack+'\n')});
  await page.goto(`http://127.0.0.1:${server.address().port}/administrar_empresa/qa_chat.html?empresa_id=12`);
  await page.locator('#openAIDrawer').click();
  await page.locator('#aiChatDrawer.open').waitFor();
  await page.locator('#aiChatHelpPrompt').click();
  assert.match(await page.locator('#aiChatInput').inputValue(),/paso a paso/);
  await page.locator('#aiChatInput').fill('Explícame paso a paso cómo agregar un consumo a una habitación');
  await page.locator('#aiChatForm button[type=submit]').click();
  await page.locator('.ai-chat-message.assistant').filter({hasText:'Abre Estaciones'}).waitFor({timeout:8000}).catch(async e=>{process.stdout.write(await page.locator('#aiChatDrawer').innerText());await page.screenshot({path:path.join(out,'failure.png')});throw e;});
  assert.equal(await page.locator('.ai-enterprise-card').count(),0);
  const before=submits;
  await page.locator('#aiChatMicBtn').click();
  await page.waitForFunction(()=>document.querySelector('#aiChatInput').value.includes('dos cervezas'));
  assert.equal(submits,before,'Dictation must not submit automatically');
  for (const theme of ['dark','light']) {
    await page.evaluate(t=>{document.documentElement.classList.toggle('theme-light',t==='light');document.documentElement.dataset.theme=t},theme);
    for (const size of [{width:1440,height:1000,name:'desktop'},{width:390,height:844,name:'mobile'}]) {
      await page.setViewportSize(size);
      await page.screenshot({path:path.join(out,`${theme}_${size.name}.png`)});
      const geometry=await page.evaluate(()=>{const el=document.querySelector('#aiChatDrawer');const r=el.getBoundingClientRect();return {right:r.right,bottom:r.bottom,width:innerWidth,height:innerHeight,overflow:document.documentElement.scrollWidth>innerWidth}});
      assert.ok(!geometry.overflow && geometry.right<=geometry.width+1 && geometry.bottom<=geometry.height+1,JSON.stringify(geometry));
    }
  }
  await page.locator('#aiChatHistoryBtn').click();
  await page.locator('.ai-chat-history-item').first().click();
  await page.locator('#aiChatInput').fill('Agrega dos cervezas a la habitación uno');
  await page.locator('#aiChatForm button[type=submit]').click();
  await page.locator('[data-enterprise-confirm]').waitFor();
  await page.locator('[data-enterprise-confirm]').scrollIntoViewIfNeeded();
  await page.screenshot({path:path.join(out,'proposal_mobile.png')});
  await page.locator('[data-enterprise-confirm]').click();
  await page.locator('#aiChatNotice').filter({hasText:'Confirmación temporalmente no disponible'}).waitFor();
  await page.locator('[data-enterprise-confirm]').click();
  await page.getByText('Acción completada y verificada.').waitFor();
  assert.equal(confirmKeys.length,2);assert.equal(confirmKeys[0],confirmKeys[1],'Retry must preserve idempotency key');
  await page.locator('#aiChatAttachment').setInputFiles({name:'carta.csv',mimeType:'text/csv',buffer:Buffer.from('producto,precio\nCerveza,5000')});
  await page.locator('#aiChatAttachmentName').filter({hasText:'carta.csv'}).waitFor();
  await page.locator('#aiChatClearAttachment').click();
  await page.locator('#closeAIDrawer').click();
  await page.locator('#openAIDrawer').click();
  await page.locator('#aiChatDrawer.open').waitFor();
  assert.deepEqual(errors,[]);
  process.stdout.write(JSON.stringify({ok:true,submits,themes:2,viewports:2,dictation:'simulated; editable; no autosend',history:true,proposal:true,consoleErrors:errors,screenshots:out})+'\n');
  await browser.close();server.close();
})().catch(e=>{process.stderr.write(e.stack+'\n');server.close();process.exit(1)});
