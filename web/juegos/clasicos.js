import {BaseGame,copy} from './core.js';
const colors=['','#f2cf65','#7ed8d0','#8eb9f0','#ad9ddd','#edaa76','#8ecb98','#ed91a0'];
const pieces=[[[1,1],[1,1]],[[0,0,0,0],[2,2,2,2],[0,0,0,0],[0,0,0,0]],[[3,0,0],[3,3,3],[0,0,0]],[[0,4,0],[4,4,4],[0,0,0]],[[0,0,5],[5,5,5],[0,0,0]],[[0,6,6],[6,6,0],[0,0,0]],[[7,7,0],[0,7,7],[0,0,0]]];
export class Tetris extends BaseGame {
 constructor(c,a){super(c,a,'tetris');c.width=440;c.height=560;this.reset()}
 reset(){this.s={score:0,over:false,board:Array.from({length:20},()=>Array(10).fill(0)),bag:[],next:[],hold:null,held:false,lines:0,level:1,timer:0};for(let i=0;i<4;i++)this.s.next.push(this.drawPiece());this.spawn();this.render()}
 drawPiece(){if(!this.s.bag.length){this.s.bag=[0,1,2,3,4,5,6];for(let i=6;i>0;i--){let j=Math.floor(Math.random()*(i+1));[this.s.bag[i],this.s.bag[j]]=[this.s.bag[j],this.s.bag[i]]}}return this.s.bag.pop()}
 spawn(id){const s=this.s;s.id=id??s.next.shift();if(id===undefined)s.next.push(this.drawPiece());s.piece=copy(pieces[s.id]);s.x=3;s.y=0;s.timer=0;if(this.hit(s.x,s.y,s.piece))s.over=true;}
 hit(x,y,m){return m.some((row,dy)=>row.some((v,dx)=>v&&(x+dx<0||x+dx>=10||y+dy>=20||(y+dy>=0&&this.s.board[y+dy][x+dx]))))}
 lock(){const s=this.s;for(let y=0;y<s.piece.length;y++)for(let x=0;x<s.piece[y].length;x++)if(s.piece[y][x]){if(s.y+y<0){s.over=true;return}s.board[s.y+y][s.x+x]=s.piece[y][x]}
  const remain=s.board.filter(row=>row.some(v=>!v)),n=20-remain.length;while(remain.length<20)remain.unshift(Array(10).fill(0));s.board=remain;s.score+=[0,100,300,500,800][n]*s.level;s.lines+=n;s.level=1+Math.floor(s.lines/10);s.held=false;this.sound.effect(n?'score':'move');this.spawn();}
 drop(){if(!this.hit(this.s.x,this.s.y+1,this.s.piece)){this.s.y++;return true}this.lock();return false}
 action(k){if(this.over)return;const s=this.s;if(k==='ArrowLeft'||k==='ArrowRight'){const dx=k==='ArrowLeft'?-1:1;if(!this.hit(s.x+dx,s.y,s.piece))s.x+=dx}
  else if(k==='ArrowDown'){if(this.drop())s.score++}
  else if(k==='ArrowUp'||k==='x'||k==='z'){let m=s.piece[0].map((_,i)=>s.piece.map(row=>row[i]).reverse());if(k==='z')m=s.piece[0].map((_,i)=>s.piece.map(row=>row[row.length-1-i]));for(const [dx,dy] of [[0,0],[-1,0],[1,0],[-2,0],[2,0],[0,-1],[0,-2]])if(!this.hit(s.x+dx,s.y+dy,m)){s.piece=m;s.x+=dx;s.y+=dy;break}}
  else if(k===' '){let n=0;while(!this.hit(s.x,s.y+1,s.piece)){s.y++;n++}s.score+=n*2;this.lock()}
  else if(k==='c'&&!s.held){let old=s.hold;s.hold=s.id;this.spawn(old??undefined);s.held=true}this.render();}
 update(dt){if(this.over)return;this.s.timer+=dt;if(this.s.timer>Math.max(.09,.8*Math.pow(.82,this.s.level-1))){this.s.timer=0;this.drop()}this.render()}
 get info(){return `Nivel ${this.s.level} · ${this.s.lines} líneas`}
 render(){const c=this.canvas.getContext('2d'),s=this.s;c.fillStyle='#142f37';c.fillRect(0,0,440,560);const block=(x,y,v,alpha=1,size=24)=>{c.globalAlpha=alpha;c.fillStyle=colors[v];c.fillRect(x+1,y+1,size-2,size-2);c.fillStyle='#ffffff30';c.fillRect(x+3,y+3,size-6,4);c.globalAlpha=1};
  c.fillStyle='#0b232a';c.fillRect(18,50,240,480);for(let y=0;y<20;y++)for(let x=0;x<10;x++){c.strokeStyle='#ffffff09';c.strokeRect(18+x*24,50+y*24,24,24);if(s.board[y][x])block(18+x*24,50+y*24,s.board[y][x])}
  let gy=s.y;while(!this.hit(s.x,gy+1,s.piece))gy++;s.piece.forEach((row,y)=>row.forEach((v,x)=>{if(v){block(18+(s.x+x)*24,50+(gy+y)*24,v,.2);block(18+(s.x+x)*24,50+(s.y+y)*24,v)}}));
  c.fillStyle='#dceecb';c.font='600 13px system-ui';c.fillText('SIGUIENTES',284,74);s.next.slice(0,3).forEach((id,i)=>pieces[id].forEach((row,y)=>row.forEach((v,x)=>{if(v)block(284+x*22,92+i*82+y*22,v,1,22)})));
  c.fillStyle='#dceecb';c.fillText('RESERVA · C',284,372);if(s.hold!==null)pieces[s.hold].forEach((row,y)=>row.forEach((v,x)=>{if(v)block(284+x*22,390+y*22,v,1,22)}));c.fillStyle='#9ab9b0';c.font='12px system-ui';c.fillText(`NIVEL ${s.level}`,284,491);c.fillText(`${s.lines} LÍNEAS`,284,513);
 }
}

export class Buscaminas extends BaseGame {
 constructor(c,a){super(c,a,'buscaminas');c.width=720;c.height=600;this.reset()}
 reset(){this.s={score:0,over:false,won:false,started:false,time:0,flags:[],open:[],mines:[],flagMode:false};this.render()}
 neighbors(i){const x=i%12,y=Math.floor(i/12),a=[];for(let dy=-1;dy<=1;dy++)for(let dx=-1;dx<=1;dx++)if((dx||dy)&&x+dx>=0&&x+dx<12&&y+dy>=0&&y+dy<10)a.push(i+dx+dy*12);return a}
 action(k){if(k==='f'){this.s.flagMode=!this.s.flagMode;this.sound.effect('flag');this.render()}}
 click(p,right=false){if(this.over)return;const x=Math.floor((p.x-24)/56),y=Math.floor((p.y-26)/56);if(x<0||x>=12||y<0||y>=10)return;const i=x+y*12,s=this.s;
  if(right||s.flagMode){if(s.open.includes(i))return;s.flags=s.flags.includes(i)?s.flags.filter(v=>v!==i):s.flags.length<18?[...s.flags,i]:s.flags;this.sound.effect('flag');this.render();return}
  if(s.flags.includes(i))return;if(!s.started){const safe=[i,...this.neighbors(i)],pool=Array.from({length:120},(_,j)=>j).filter(j=>!safe.includes(j));for(let j=pool.length-1;j>0;j--){const n=Math.floor(Math.random()*(j+1));[pool[j],pool[n]]=[pool[n],pool[j]]}s.mines=pool.slice(0,18);s.started=true}
  if(s.open.includes(i)){const neighbors=this.neighbors(i);if(neighbors.filter(n=>s.flags.includes(n)).length===neighbors.filter(n=>s.mines.includes(n)).length)for(const n of neighbors)if(!s.flags.includes(n))this.reveal(n)}else this.reveal(i);
  if(!s.over&&s.open.length===102){s.over=true;s.won=true;s.score+=Math.max(0,5000-Math.floor(s.time)*5);this.sound.effect('win')}this.render();}
 reveal(i){const s=this.s;if(s.over||s.open.includes(i)||s.flags.includes(i))return;if(s.mines.includes(i)){s.over=true;s.exploded=i;this.sound.effect('lose');return}const stack=[i];while(stack.length){const n=stack.pop();if(s.open.includes(n)||s.flags.includes(n))continue;s.open.push(n);s.score+=10;const ns=this.neighbors(n);if(!ns.some(k=>s.mines.includes(k)))stack.push(...ns.filter(k=>!s.open.includes(k)))}this.sound.effect('score')}
 update(dt){if(this.s.started&&!this.over)this.s.time+=dt}
 get info(){return `${18-this.s.flags.length} minas · ${Math.floor(this.s.time)} s${this.s.flagMode?' · Bandera activa':''}`}
 render(){const c=this.canvas.getContext('2d'),s=this.s;c.fillStyle='#e5eddd';c.fillRect(0,0,720,600);const inks=['','#37699b','#438455','#c55243','#8361a5','#a87831','#328482','#594263','#414c36'];c.textAlign='center';c.textBaseline='middle';
  for(let i=0;i<120;i++){const x=24+(i%12)*56,y=26+Math.floor(i/12)*56,open=s.open.includes(i);c.fillStyle=open?'#f6f7ec':'#b3c8a1';c.beginPath();c.roundRect(x+2,y+2,51,51,7);c.fill();if(!open){c.fillStyle='#d4dfc8';c.fillRect(x+8,y+7,37,3)}
   if(s.flags.includes(i)){c.fillStyle='#b34930';c.font='29px system-ui';c.fillText('⚑',x+28,y+29)}else if(s.over&&!s.won&&s.mines.includes(i)){c.fillStyle=i===s.exploded?'#d34d37':'#4d6450';c.beginPath();c.arc(x+28,y+27,11,0,Math.PI*2);c.fill();for(let a=0;a<8;a++){c.strokeStyle=c.fillStyle;c.beginPath();c.moveTo(x+28+Math.cos(a*Math.PI/4)*9,y+27+Math.sin(a*Math.PI/4)*9);c.lineTo(x+28+Math.cos(a*Math.PI/4)*17,y+27+Math.sin(a*Math.PI/4)*17);c.stroke()}}
   else if(open){let n=this.neighbors(i).filter(j=>s.mines.includes(j)).length;if(n){c.fillStyle=inks[n];c.font='600 25px system-ui';c.fillText(n,x+28,y+29)}}
  }
 }
}

const suits=['♠','♥','♣','♦'];const red=c=>c.suit%2===1;const rank=c=>['','A','2','3','4','5','6','7','8','9','10','J','Q','K'][c.rank];
export class Solitario extends BaseGame {
 constructor(c,a){super(c,a,'solitario');c.width=840;c.height=650;this.selected=null;this.history=[];this.reset()}
 reset(){let deck=Array.from({length:52},(_,i)=>({rank:i%13+1,suit:Math.floor(i/13),up:false}));for(let i=51;i>0;i--){let j=Math.floor(Math.random()*(i+1));[deck[i],deck[j]]=[deck[j],deck[i]]}const cols=Array.from({length:7},()=>[]);for(let i=0;i<7;i++){for(let j=0;j<=i;j++)cols[i].push(deck.pop());cols[i].at(-1).up=true}this.s={score:0,over:false,won:false,stock:deck,waste:[],foundation:[[],[],[],[]],cols,moves:0,time:0,revealed:0};this.history=[];this.selected=null;this.render()}
 remember(){this.history.push(copy(this.s));if(this.history.length>40)this.history.shift()}
 async load(p){await super.load(p);this.history=[];this.selected=null;this.render()}
 source(sel){if(sel.type==='col')return this.s.cols[sel.col];if(sel.type==='waste')return this.s.waste;return this.s.foundation[sel.col]}
 updateScore(){this.s.score=this.s.foundation.reduce((n,p)=>n+p.length,0)*100+this.s.revealed*5;if(this.s.foundation.every(p=>p.length===13)){this.s.won=true;this.s.over=true;this.sound.effect('win')}}
 move(dest){if(!this.selected)return false;const sel=this.selected,src=this.source(sel),moving=src.slice(sel.index),first=moving[0];if(!first)return false;let target;
  if(dest.type==='foundation'){if(moving.length!==1||sel.type==='foundation')return false;target=this.s.foundation[dest.col];if(first.suit!==dest.col||first.rank!==target.length+1)return false}
  else {if(sel.type==='col'&&sel.col===dest.col)return false;target=this.s.cols[dest.col];const last=target.at(-1);if(last?(!last.up||last.rank!==first.rank+1||red(last)===red(first)):first.rank!==13)return false}
  this.remember();target.push(...src.splice(sel.index));if(sel.type==='col'&&src.length&&!src.at(-1).up){src.at(-1).up=true;this.s.revealed++}this.s.moves++;this.selected=null;this.updateScore();this.sound.effect('card');return true;
 }
 click(p){if(this.over)return;const col=Math.floor((p.x-18)/116);if(col<0||col>6)return;if(p.y<151){if(col===0){this.remember();if(this.s.stock.length){let card=this.s.stock.pop();card.up=true;this.s.waste.push(card)}else {this.s.stock=this.s.waste.reverse();this.s.waste=[]}this.s.moves++;this.selected=null;this.sound.effect('card')}
  else if(col===1&&this.s.waste.length){const sel={type:'waste',index:this.s.waste.length-1};if(this.selected?.type==='waste')this.auto(sel);else this.selected=sel}
  else if(col>=3){const dest={type:'foundation',col:col-3};if(!this.move(dest)&&this.s.foundation[col-3].length)this.selected={...dest,index:this.s.foundation[col-3].length-1}}}
  else {const pile=this.s.cols[col],step=this.step(pile),i=Math.min(pile.length-1,Math.floor((p.y-183)/step));if(p.y<183)return;const dest={type:'col',col};if(!this.move(dest)){if(i>=0&&pile[i].up){const sel={...dest,index:i};if(this.selected?.type==='col'&&this.selected.col===col&&this.selected.index===i)this.auto(sel);else this.selected=sel}else this.selected=null}}this.render();}
 auto(sel){this.selected=sel;const card=this.source(sel)[sel.index];if(card)this.move({type:'foundation',col:card.suit})}
 action(k){if(k==='u'&&this.history.length){this.s=this.history.pop();this.selected=null;this.sound.effect('card')}if(k==='a'){for(let col=0;col<7;col++){const p=this.s.cols[col];if(p.length)this.auto({type:'col',col,index:p.length-1})}if(this.s.waste.length)this.auto({type:'waste',index:this.s.waste.length-1});this.selected=null}this.render()}
 update(dt){this.s.time+=dt}get info(){return `${this.s.moves} movimientos · ${this.s.foundation.reduce((n,p)=>n+p.length,0)}/52 cartas`}
 step(p){return Math.min(32,340/Math.max(1,p.length-1))}
 render(){const c=this.canvas.getContext('2d'),s=this.s;c.fillStyle='#24694f';c.fillRect(0,0,840,650);c.strokeStyle='#ffffff0b';for(let y=0;y<650;y+=18){c.beginPath();c.moveTo(0,y);c.lineTo(840,y+150);c.stroke()}c.textAlign='center';c.textBaseline='middle';
  const empty=(x,y,label)=>{c.strokeStyle='#b0d9ad77';c.lineWidth=2;c.setLineDash([5,5]);c.strokeRect(x,y,96,126);c.setLineDash([]);c.font='26px Georgia';c.fillStyle='#b6d5ab';c.fillText(label,x+48,y+64)};
  const card=(v,x,y,selected=false)=>{c.fillStyle=selected?'#f4e8aa':v.up?'#fffcf0':'#193f47';c.beginPath();c.roundRect(x,y,96,126,7);c.fill();c.strokeStyle=selected?'#ffd354':v.up?'#e4e3cb':'#8bb8a2';c.lineWidth=selected?4:1;c.stroke();if(v.up){c.fillStyle=red(v)?'#b44838':'#183b3e';c.font='bold 21px Georgia';c.textAlign='left';c.fillText(rank(v)+suits[v.suit],x+7,y+18);c.textAlign='center';c.font='42px Georgia';c.fillText(suits[v.suit],x+48,y+71);c.font='16px Georgia';c.fillText(rank(v),x+78,y+109)}else {c.strokeStyle='#749e8a';c.strokeRect(x+7,y+7,82,112);c.font='35px Georgia';c.fillStyle='#b6c89b';c.fillText('✦',x+48,y+64)}};
  empty(18,22,'↻');if(s.stock.length)card({up:false},18,22);empty(134,22,'');if(s.waste.length)card(s.waste.at(-1),134,22,this.selected?.type==='waste');for(let i=0;i<4;i++){let x=18+(i+3)*116;empty(x,22,suits[i]);if(s.foundation[i].length)card(s.foundation[i].at(-1),x,22,this.selected?.type==='foundation'&&this.selected.col===i)}
  s.cols.forEach((p,col)=>{const x=18+col*116;empty(x,183,'K');p.forEach((v,i)=>card(v,x,183+i*this.step(p),this.selected?.type==='col'&&this.selected.col===col&&i>=this.selected.index))});
 }
}
