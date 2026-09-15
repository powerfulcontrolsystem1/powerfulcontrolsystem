import test from 'node:test';
import assert from 'node:assert/strict';
import {GTASAudio} from '../web/juegos/simsong/audio.js';

function fixture(){
 const sources=[],gains=[];
 const param=()=>({value:0,setValueAtTime(v){this.value=v},exponentialRampToValueAtTime(){}});
 const node=()=>({connect(){},disconnect(){this.disconnected=true},start(){this.started=true},stop(){this.stopped=true}});
 const sound={enabled:true,ctx:{state:'running',currentTime:0,sampleRate:100,destination:{},
  createOscillator(){const o={...node(),frequency:param()};sources.push(o);return o},
  createGain(){const g={...node(),gain:param()};gains.push(g);return g},
  createBufferSource(){const s=node();sources.push(s);return s},
  createBuffer(c,n){return {getChannelData(){return new Float32Array(n)}}},
  createBiquadFilter(){return {...node(),frequency:param(),Q:param()}}
 }};
 return {sound,audio:new GTASAudio(sound),sources,gains};
}

test('Muted or locked sound creates no audio sources',()=>{
 const f=fixture();for(const state of ['muted','locked']){f.sound.enabled=state!=='muted';f.sound.ctx.state=state==='locked'?'suspended':'running';f.audio.laser();f.audio.alienHit(true);f.audio.playerHit();f.audio.update(3,1)}assert.equal(f.sources.length,0);
});
test('Approaching enemies become louder and distant enemies stay silent',()=>{
 const volumes=[];for(const distance of [2,30,60]){const f=fixture();f.audio.musicTime=100;f.audio.update(3,distance);volumes.push(f.gains.reduce((sum,g)=>sum+g.gain.value,0))}assert.ok(volumes[0]>volumes[1]);assert.ok(volumes[1]>0);assert.equal(volumes[2],0);
});
test('Pause stops active sources and ended sources release their connections',()=>{
 const f=fixture();f.audio.laser();f.audio.alienHit(true);assert.ok(f.audio.nodes.size>0);f.audio.stop();assert.equal(f.audio.nodes.size,0);for(const source of f.sources){assert.ok(source.stopped);source.onended();assert.ok(source.disconnected)}assert.ok(f.gains.every(g=>g.disconnected));
});
test('Music, laser, enemy cries and player damage have distinct sound signatures',()=>{
 const signatures=[];for(const play of [a=>a.update(.1,99),a=>a.laser(),a=>a.alienHit(),a=>a.alienHit(true),a=>a.enemyLaser(),a=>a.playerHit()]){const f=fixture();play(f.audio);assert.ok(f.sources.every(s=>s.started&&s.stopped));signatures.push(f.sources.map(s=>s.frequency?.value||'noise').join(','))}assert.equal(new Set(signatures).size,6);
});
