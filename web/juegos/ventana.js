// Non-modal, same-origin game companion. It never reads the business iframe.
let widget, frame, returnFocus, minimized = false;
function notifyPause(){frame?.contentWindow?.postMessage({type:'pcs-games:pause'},location.origin)}
function clampPosition(){
  if(!widget||widget.hidden)return;
  const r=widget.getBoundingClientRect();
  widget.style.left=Math.max(4,Math.min(r.left,innerWidth-r.width-4))+'px';
  widget.style.top=Math.max(4,Math.min(r.top,innerHeight-Math.min(r.height,innerHeight-8)-4))+'px';
}
export function openGamesWidget(launcher){
  returnFocus=launcher;
  if(!widget){
    const css=document.createElement('link');css.rel='stylesheet';css.href='/juegos/ventana.css';
    widget=document.createElement('section');widget.id='pcs-games-window';widget.setAttribute('role','dialog');widget.setAttribute('aria-modal','false');widget.setAttribute('aria-label','Juegos PCS, ventana flotante');
    widget.innerHTML='<header class="pcs-games-drag"><span class="pcs-games-emblem">✦</span><strong>PCS <small>PLAY</small></strong><span class="pcs-games-caption">Tu consola personal</span><button type="button" data-window="minimize" aria-label="Minimizar juegos">−</button><button type="button" data-window="close" aria-label="Cerrar juegos">×</button></header><iframe title="Menú y juegos PCS" src="/juegos.html?compact=1" allow="fullscreen; gamepad" allowfullscreen></iframe>';
    widget.style.position='fixed';widget.style.visibility='hidden';
    document.body.append(widget);frame=widget.querySelector('iframe');
    css.onload=()=>{widget.style.left=Math.max(6,innerWidth-Math.min(660,innerWidth-12)-24)+'px';widget.style.top='70px';widget.style.visibility='';clampPosition()};document.head.append(css);
    widget.style.left=Math.max(6,innerWidth-Math.min(660,innerWidth-12)-24)+'px';widget.style.top='70px';
    widget.querySelector('[data-window=close]').onclick=()=>{notifyPause();widget.hidden=true;returnFocus?.focus()};
    widget.querySelector('[data-window=minimize]').onclick=()=>{notifyPause();minimized=!minimized;widget.classList.toggle('minimized',minimized);widget.querySelector('[data-window=minimize]').setAttribute('aria-label',minimized?'Restaurar juegos':'Minimizar juegos');clampPosition()};
    const handle=widget.querySelector('header');let drag;
    handle.addEventListener('pointerdown',e=>{if(e.target.closest('button'))return;const r=widget.getBoundingClientRect();drag={x:e.clientX-r.left,y:e.clientY-r.top};handle.setPointerCapture(e.pointerId);widget.classList.add('dragging');e.preventDefault()});
    handle.addEventListener('pointermove',e=>{if(!drag)return;widget.style.left=(e.clientX-drag.x)+'px';widget.style.top=(e.clientY-drag.y)+'px';clampPosition()});
    const release=()=>{drag=null;widget.classList.remove('dragging')};handle.onpointerup=release;handle.onpointercancel=release;handle.onlostpointercapture=release;
    window.addEventListener('resize',clampPosition);
    new ResizeObserver(clampPosition).observe(widget);
    window.addEventListener('message',e=>{if(e.origin!==location.origin||e.source!==frame.contentWindow)return;if(e.data?.type==='pcs-games:close'){notifyPause();widget.hidden=true;returnFocus?.focus()}});
  }
  widget.hidden=false;minimized=false;widget.classList.remove('minimized');clampPosition();
  widget.querySelector('[data-window=minimize]').setAttribute('aria-label','Minimizar juegos');frame.focus();
}
