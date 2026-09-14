export const copy = value => JSON.parse(JSON.stringify(value));
export class Sound {
 constructor(){this.enabled=true;this.ctx=null;this.last=0;}
 async unlock(){if(!this.ctx)this.ctx=new(window.AudioContext||window.webkitAudioContext)();await this.ctx.resume();}
 tone(freq=440,duration=.09,type='sine',volume=.08){
  if(!this.enabled||!this.ctx||this.ctx.state!=='running')return;
  const t=this.ctx.currentTime;if(t-this.last<.025)return;this.last=t;
  const osc=this.ctx.createOscillator(),g=this.ctx.createGain();osc.type=type;osc.frequency.setValueAtTime(freq,t);osc.frequency.exponentialRampToValueAtTime(Math.max(30,freq*.6),t+duration);
  g.gain.setValueAtTime(volume,t);g.gain.exponentialRampToValueAtTime(.0001,t+duration);osc.connect(g);g.connect(this.ctx.destination);osc.start();osc.stop(t+duration);
 }
 effect(kind){const notes={move:330,score:880,win:1320,lose:90,shoot:130,power:660,card:520,flag:420};this.tone(notes[kind]||440,kind==='shoot'?.14:.12,kind==='shoot'?'sawtooth':'sine');}
}
export class BaseGame {
 constructor(canvas,sound,id){this.canvas=canvas;this.sound=sound;this.id=id;this.s={score:0,over:false};this.keys={};this.active=false;}
 get score(){return this.s.score||0}get over(){return this.s.over}get info(){return ''}
 save(){return {schema:1,game:this.id,score:this.score,data:copy(this.s)}}
 async load(p){if(p?.schema!==1||p.game!==this.id||!p.data)throw Error('La versión de esta partida no es compatible.');this.s=copy(p.data);this.keys={};this.render();}
 key(key,down){this.keys[key]=down;if(down)this.action(key)}
 action(){}update(){}render(){}resize(){}pause(){this.active=false;this.keys={}}
 point(e){const r=this.canvas.getBoundingClientRect();return {x:(e.clientX-r.left)*this.canvas.width/r.width,y:(e.clientY-r.top)*this.canvas.height/r.height}}
}
