import {Sound} from './core.js';
document.body.classList.toggle('compact',new URLSearchParams(location.search).has('compact'));
const catalog=[
 {id:'simsong',name:'GTA SIMSong',tag:'SPRINGFIELD · MUNDO ABIERTO',help:'WASD: caminar. Arrastra o usa flechas para mirar. Espacio: saltar. Shift: correr. E: entrar, salir o conducir. F o clic: láser. V: tres cámaras. M: misiones. Derrota a los invasores y sus ovnis.',keys:[['←','a'],['↑','w'],['↓','s'],['→','d'],['Saltar',' '],['Usar','e'],['Correr','Shift'],['Láser','f'],['Cámara','v'],['Misiones','m']]},
 {id:'pacman',name:'Pac-Man',tag:'ARCADE · 10 NIVELES',help:'Flechas o WASD para moverte. Come las pastillas grandes para perseguir a los fantasmas. Completa diez niveles.',keys:[['←','ArrowLeft'],['↑','ArrowUp'],['↓','ArrowDown'],['→','ArrowRight']]},
 {id:'tetris',name:'Tetris',tag:'ENCUENTRA TU RITMO',help:'Flechas: mover y girar. Espacio: caída instantánea. C: reservar. Z: giro inverso.',keys:[['←','ArrowLeft'],['Girar','ArrowUp'],['↓','ArrowDown'],['→','ArrowRight'],['Caer',' '],['Reservar','c']]},
 {id:'buscaminas',name:'Buscaminas',tag:'UNA BUENA INTUICIÓN',help:'Toca para descubrir. Bandera o clic derecho para marcar. Primera apertura segura.',keys:[['⚑ Bandera','f']]},
 {id:'solitario',name:'Solitario',tag:'TU MOMENTO DE CALMA',help:'Toca la baraja. Selecciona una carta o secuencia y su destino. Alterna colores; solo reyes en columnas vacías. As a rey por palo en las bases.',keys:[['Auto','a'],['Deshacer','u']]},
 {id:'sorpresa',name:'Una puerta inesperada',tag:'EXPLORA OTRO MUNDO',help:'WASD: nadar. Arrastra o flechas: mirar. Espacio: subir. C: bajar. E: pulso explorador. Encuentra diez tesoros y repón oxígeno en la superficie.',keys:[['←','a'],['↑','w'],['↓','s'],['→','d'],['Subir',' '],['Bajar','c'],['Pulso','e']]}
];
const sound=new Sound(),rooms=new Map(),menu=document.getElementById('game-menu'),games=document.getElementById('games');
const status=document.getElementById('session-status');let active=null,current=null,last=0,selecting=false;
async function api(id,options={},records=false){
 if(options.method==='PUT'){const token=document.cookie.split(';').map(s=>s.trim()).find(s=>s.startsWith('pcs_csrf='));options.headers={...options.headers,'X-CSRF-Token':token?decodeURIComponent(token.slice(9)):''}}
 const r=await fetch(`/api/juegos?juego=${id}${records?'&action=records':''}`,{credentials:'same-origin',cache:'no-store',...options});
 if(!r.ok){const e=Error(r.status===401?'Inicia sesión en PCS para jugar y guardar.':await r.text());e.status=r.status;throw e}return r.json();
}
function message(room,text,error=false){room.message.textContent=text;room.message.classList.toggle('error',error)}
function overlay(room,title,subtitle,button='Continuar'){room.overlay.hidden=false;room.overlay.querySelector('strong').textContent=title;room.overlay.querySelector('span').textContent=subtitle;room.overlay.querySelector('button').textContent=button}
function refreshHUD(room){room.hud.children[0].textContent=`${Number(room.game.score).toLocaleString('es-CO')} pts`;room.hud.children[1].textContent=room.game.info}
async function save(room){
 if(!room.game||!room.loaded)return false;
 if(room.saving){room.saveAgain=true;return room.saving}
 message(room,'Guardando…');room.saving=(async()=>{try{do{room.saveAgain=false;const data=await api(room.item.id,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({version:room.version,puntaje:room.game.score,estado:room.game.save()})});room.version=data.version;await ranking(room)}while(room.saveAgain);message(room,`Guardado · ${new Date().toLocaleTimeString('es-CO')}`);return true}catch(e){message(room,e.message,true);if(e.status===409){pause(room,false);room.loaded=false;overlay(room,'Partida actualizada en otro equipo','Pulsa Cargar para recuperar la última versión.');room.overlay.querySelector('button').disabled=true}return false}finally{room.saving=null;room.saveAgain=false}})();return room.saving;
}
function pause(room,persist=true){
 if(!room?.game)return;room.running=false;room.game.pause();room.card.classList.remove('playing');room.releaseControls?.();if(active===room)active=null;
 room.pause.textContent='▶ Jugar';overlay(room,room.game.over?(room.game.s?.won?'¡Lo conseguiste!':'Partida terminada'):'Pausa',room.game.over?'Tu récord se conserva. Puedes empezar otra partida.':'Continúa cuando quieras.');room.overlay.querySelector('button').hidden=room.game.over;
 if(persist)void save(room);
}
async function play(room){if(!room.game||!room.loaded||room.game.over)return;if(active&&active!==room)pause(active);try{await sound.unlock();sound.effect('move')}catch{message(room,'Sonido no disponible en este navegador.',true)}room.game.active=true;room.game.keys={};active=room;room.running=true;room.card.classList.add('playing');room.overlay.hidden=true;room.pause.textContent='Ⅱ Pausar';room.canvas.focus({preventScroll:true})}
async function ranking(room){
 try{const {records}=await api(room.item.id,{},true);const list=room.card.querySelector('.record-list');list.replaceChildren();if(!records.length){const li=document.createElement('li');li.textContent='El primer récord puede ser tuyo.';list.append(li)}for(const [i,record]of records.slice(0,6).entries()){const li=document.createElement('li'),name=document.createElement('span'),score=document.createElement('b'),date=document.createElement('small');name.textContent=`${i===0?'♛':i+1} ${record.nombre}`;score.textContent=Number(record.puntaje).toLocaleString('es-CO');date.textContent=new Date(record.fecha).toLocaleString('es-CO');li.append(name,score,date);list.append(li)}}catch(e){room.card.querySelector('.record-list').textContent=e.message}
}
async function load(room,explicit=false){
 try{if(explicit)pause(room,false);if(room.saving)await room.saving;const data=await api(room.item.id);room.version=data.partida.version;room.loaded=true;const button=room.overlay.querySelector('button');button.disabled=false;button.hidden=false;
  if(data.partida.estado){await room.game.load(data.partida.estado);overlay(room,'Tu partida te espera','Retoma donde te quedaste.','Continuar partida');message(room,`Último guardado: ${new Date(data.partida.actualizado_en).toLocaleString('es-CO')}`)}else{if(explicit)await room.game.reset();overlay(room,'¿Listo para jugar?',room.item.id==='simsong'?'Defiende Springfield: 12 extraterrestres y 3 ovnis. Tienes 3 vidas; elige tus misiones y guarda tu avance.':'Activa los controles y el sonido.','Jugar');message(room,'Sin partida guardada')}
  room.game.render();refreshHUD(room);status.textContent=`${data.nombre} · Partidas privadas`;return true;
 }catch(e){room.loaded=false;overlay(room,'No se pudo cargar la partida',e.message,'Reintentar');message(room,e.message,true);if(e.status===401){status.replaceChildren();const link=document.createElement('a');link.href='/login.html';link.target='_top';link.textContent='Iniciar sesión en PCS';status.append(link)}return false}
}
function fit(room){if(!room?.game||room.card.hidden)return;if(room.item.id==='simsong'||room.item.id==='sorpresa'){room.game.resize();return}const bounds=room.canvas.parentElement.getBoundingClientRect(),scale=Math.min(bounds.width/room.canvas.width,bounds.height/room.canvas.height);room.canvas.style.width=(room.canvas.width*scale)+'px';room.canvas.style.height=(room.canvas.height*scale)+'px';room.game.render()}
async function showMenu(){if(current){pause(current,false);await save(current);current.card.hidden=true}current=null;games.hidden=true;menu.hidden=false;document.body.classList.remove('in-game');document.querySelector('h1').textContent='¿A qué jugamos?';history.replaceState(null,'',location.pathname+location.search);menu.querySelector('button')?.focus()}
function controls(room){
 const group=room.card.querySelector('.controls'),dpad=document.createElement('div'),actions=document.createElement('div');dpad.className='dpad';actions.className='action-pad';group.append(dpad,actions);const releases=[];
 for(const [index,[label,key]]of room.item.keys.entries()){
  const directional=room.item.keys.length>=4&&index<4,b=document.createElement('button');b.type='button';b.textContent=label;b.setAttribute('aria-label',label);b.dataset.key=key;b.className=directional?'direction dir-'+index:'action';(directional?dpad:actions).append(b);let repeat;
  b.addEventListener('pointerdown',e=>{e.preventDefault();if(!room.running)return;b.setPointerCapture(e.pointerId);room.game.key(key,true);b.classList.add('pressed');if(room.item.id==='tetris'&&['ArrowLeft','ArrowRight','ArrowDown'].includes(key))repeat=setInterval(()=>room.game.key(key,true),130);refreshHUD(room)});
  const release=()=>{clearInterval(repeat);room.game?.key(key,false);b.classList.remove('pressed')};b.onpointerup=release;b.onpointercancel=release;b.onlostpointercapture=release;releases.push(release);
 }if(!dpad.children.length)dpad.remove();room.releaseControls=()=>releases.forEach(f=>f());
}
function makeRoom(item){
 const card=document.createElement('section');card.id=item.id;card.className='game-card';card.innerHTML=`<div class="game-toolbar"><button data-action="menu" aria-label="Volver al menú">‹ Menú</button><h2>${item.name}</h2><button data-action="sound" aria-pressed="true" aria-label="Sonido">♫</button><button data-action="full" aria-label="Pantalla completa">⛶</button></div><div class="stage" data-game="${item.id}"><canvas tabindex="0" aria-label="${item.name}, área de juego"></canvas><div class="hud"><span>0 pts</span><span></span></div><div class="controls" aria-label="Mando de ${item.name}"></div><div class="overlay"><strong>Preparando juego…</strong><span>Un momento.</span><button type="button" disabled>Jugar</button></div></div><div class="game-actions"><button data-action="pause">▶ Jugar</button><button data-action="save">Guardar</button><button data-action="load">Cargar</button><button data-action="new">Nueva</button><button data-action="details" aria-expanded="false">Ayuda y récords</button></div><p class="save-status" role="status"></p><section class="game-details" hidden><p class="help">${item.help}</p><h3>Récords entre empresas</h3><ol class="record-list"></ol><button class="refresh-rank">Actualizar</button></section>`;
 games.append(card);const room={item,card,canvas:card.querySelector('canvas'),overlay:card.querySelector('.overlay'),hud:card.querySelector('.hud'),message:card.querySelector('.save-status'),pause:card.querySelector('[data-action=pause]'),running:false,loaded:false,version:0};rooms.set(item.id,room);
 room.overlay.querySelector('button').onclick=()=>room.loaded?play(room):load(room);
 card.querySelector('.refresh-rank').onclick=()=>ranking(room);
 card.addEventListener('click',async e=>{const button=e.target.closest('[data-action]'),action=button?.dataset.action;if(!action)return;try{
  if(action==='menu')await showMenu();
  if(action==='pause'){if(room.running)pause(room);else await play(room)}
  if(action==='save'){pause(room,false);await save(room)}
  if(action==='load'&&confirm('¿Cargar la última partida? Se reemplazan los avances sin guardar.'))await load(room,true);
  if(action==='new'&&room.game&&room.loaded&&confirm('¿Empezar una partida nueva? Tu récord se conserva.')){pause(room,false);if(room.saving)await room.saving;await room.game.reset();room.overlay.querySelector('button').hidden=false;await save(room);refreshHUD(room);await play(room)}
  if(action==='details'){const details=card.querySelector('.game-details');details.hidden=!details.hidden;button.setAttribute('aria-expanded',String(!details.hidden));if(!details.hidden)void ranking(room)}
  if(action==='sound'){sound.enabled=!sound.enabled;for(const r of rooms.values())r.card.querySelector('[data-action=sound]').setAttribute('aria-pressed',String(sound.enabled));if(sound.enabled){await sound.unlock();sound.effect('score')}}
  if(action==='full'){if(document.fullscreenElement)await document.exitFullscreen();else if(card.requestFullscreen)await card.requestFullscreen();else message(room,'Pantalla completa no disponible; puedes ampliar la ventana.',true)}
 }catch(err){message(room,err.message,true)}});
 controls(room);room.canvas.addEventListener('contextmenu',e=>e.preventDefault());let pointer;
 room.canvas.addEventListener('pointerdown',e=>{if(!room.running)return;room.canvas.setPointerCapture(e.pointerId);pointer={x:e.clientX,y:e.clientY};if(room.item.id==='simsong'&&e.pointerType==='mouse'&&e.button===0)room.game.key('f',true);if(room.game.click){room.game.click(room.game.point(e),e.button===2);refreshHUD(room)}});
 room.canvas.addEventListener('pointermove',e=>{if(!room.running||!pointer||!room.game.look)return;room.game.look((e.clientX-pointer.x)*.006,(e.clientY-pointer.y)*.004);pointer={x:e.clientX,y:e.clientY}});
 room.canvas.onpointerup=room.canvas.onpointercancel=room.canvas.onlostpointercapture=()=>{pointer=null;if(room.item.id==='simsong')room.game?.key('f',false)};new ResizeObserver(()=>fit(room)).observe(card.querySelector('.stage'));return room;
}
async function selectGame(item){
 if(selecting)return;selecting=true;try{
  if(current){pause(current,false);await save(current);current.card.hidden=true}
  let room=rooms.get(item.id);if(room?.failed){room.card.remove();rooms.delete(item.id);room=null}if(!room)room=makeRoom(item);current=room;menu.hidden=true;games.hidden=false;room.card.hidden=false;document.body.classList.add('in-game');document.querySelector('h1').textContent=item.name;history.replaceState(null,'','#'+item.id);
  if(!room.game){try{
   if(item.id==='pacman'){const {createPacman}=await import('./pacman/engine.js');room.canvas.width=540;room.canvas.height=390;room.game=await createPacman(room.canvas,sound)}
   else if(['tetris','buscaminas','solitario'].includes(item.id)){const classes=await import('./clasicos.js');room.game=new classes[{tetris:'Tetris',buscaminas:'Buscaminas',solitario:'Solitario'}[item.id]](room.canvas,sound)}
   else if(item.id==='simsong'){const {SpringfieldGame}=await import('./simsong/game.js');room.game=await SpringfieldGame.create(room.canvas,sound)}
   else{const {WorldGame}=await import('./mundos.js');room.game=new WorldGame(room.canvas,sound,'sorpresa')}
   await load(room);void ranking(room);
  }catch(e){message(room,e.message,true);overlay(room,'No se pudo preparar el juego',e.message);room.overlay.querySelector('button').hidden=true;room.failed=true}}fit(room);
 }finally{selecting=false}
}
for(const item of catalog){const b=document.createElement('button');b.className='game-choice';b.type='button';b.dataset.game=item.id;b.innerHTML=`<img src="/juegos/portadas/${item.id}.png" alt="Vista del juego ${item.name}" loading="lazy"><span class="choice-copy"><small>${item.tag}</small><strong>${item.name}</strong></span><span class="choice-play" aria-hidden="true">▶</span>`;b.onclick=()=>selectGame(item);menu.append(b)}
const allowed=['ArrowLeft','ArrowRight','ArrowUp','ArrowDown',' ','w','a','s','d','e','q','c','x','z','u','f','m','v','1','2','3','Shift'];
function normalizeKey(key){return key.length===1?key.toLowerCase():key}
// Keys are scoped to the game iframe; the business application keeps its shortcuts.
 document.addEventListener('keydown',e=>{if(!active||['INPUT','TEXTAREA'].includes(e.target.tagName))return;const key=normalizeKey(e.key);if(key==='Escape'||key==='p'){e.preventDefault();pause(active);return}if(allowed.includes(key)){e.preventDefault();if(!e.repeat||active.item.id==='tetris')active.game.key(key,true)}});
 document.addEventListener('keyup',e=>active?.game.key(normalizeKey(e.key),false));
 document.addEventListener('visibilitychange',()=>{if(document.hidden&&active)pause(active)});window.addEventListener('blur',()=>{if(active)pause(active)});
 window.addEventListener('message',e=>{if(e.origin===location.origin&&e.source===parent&&e.data?.type==='pcs-games:pause'&&current)pause(current)});
 window.addEventListener('beforeunload',e=>{if(active){e.preventDefault();e.returnValue=''}});setInterval(()=>{if(active)void save(active)},20000);
 function tick(now){const dt=Math.min(.05,(now-last)/1000||.016);last=now;if(active?.running){const room=active;try{room.game.update(dt);refreshHUD(room);if(room.game.over){sound.effect(room.game.s?.won?'win':'lose');pause(room)}}catch(e){pause(room,false);message(room,'El juego se detuvo. Carga tu última partida.',true);console.error(e)}}requestAnimationFrame(tick)}requestAnimationFrame(tick);
 window.addEventListener('resize',()=>fit(current));document.addEventListener('fullscreenchange',()=>requestAnimationFrame(()=>fit(current)));
 const initial=catalog.find(item=>'#'+item.id===location.hash);if(initial)void selectGame(initial);
