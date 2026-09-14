// Serializable combat rules. Rendering and input cannot award the same kill twice.
export function encounters(){return [[-72,137],[-37,85],[28,78],[99,55],[-109,18],[-40,-20],[32,-25],[105,-43],[-110,-85],[30,-112],[-40,-117],[104,115]].map(([x,z],id)=>({id,kind:'alien',x,y:0,z,hp:90,cooldown:2+id*.2})).concat([[-20,120],[90,-90],[-100,-100]].map(([x,z],i)=>({id:12+i,kind:'ufo',x,y:17+i*3,z,hp:300,cooldown:4+i})));}
export function combatState(){return {lives:3,health:100,cameraMode:1,deathTime:0,invulnerable:3,fireCooldown:0,enemies:encounters(),projectiles:[],liberated:false};}
export function normalizeCombat(saved){
 const result=combatState();
 for(const k of ['lives','cameraMode'])if(Number.isInteger(saved[k]))result[k]=Math.max(0,Math.min(k==='lives'?3:2,saved[k]));
 for(const k of ['health','deathTime','invulnerable','fireCooldown'])if(Number.isFinite(saved[k]))result[k]=Math.max(0,Math.min(k==='health'?100:10,saved[k]));
 result.liberated=saved.liberated===true;
 if(Array.isArray(saved.enemies))for(const e of result.enemies){const v=saved.enemies.find(v=>v.id===e.id);if(v&&['x','y','z','hp'].every(k=>Number.isFinite(v[k]))&&Math.hypot(v.x,v.z)<198){e.x=v.x;e.z=v.z;e.y=e.kind==='ufo'?Math.max(10,Math.min(35,v.y)):0;e.hp=Math.max(0,Math.min(e.hp,v.hp));e.cooldown=Number.isFinite(v.cooldown)?Math.max(0,Math.min(10,v.cooldown)):2;}}
 if(Array.isArray(saved.projectiles))result.projectiles=saved.projectiles.slice(0,60).filter(p=>['x','y','z','vx','vy','vz','ttl'].every(k=>Number.isFinite(p[k]))&&p.ttl>0&&p.ttl<=5).map(p=>({...p}));
 return result;
}
export function hitEnemy(state,enemy,damage=30){if(enemy.hp<=0)return false;enemy.hp=Math.max(0,enemy.hp-damage);if(enemy.hp===0){state.score+=enemy.kind==='ufo'?1000:250;return true}return false;}
export function hurtPlayer(state,damage){if(state.deathTime>0||state.invulnerable>0||state.over)return false;state.health=Math.max(0,state.health-damage);state.invulnerable=.65;if(state.health===0){state.lives=Math.max(0,state.lives-1);state.deathTime=.001;state.driving=-1;state.carSpeed=0;return true}return false;}
export function advanceDeath(state,dt){if(!state.deathTime)return false;state.deathTime+=dt;if(state.deathTime<4)return true;if(state.lives===0){state.over=true;return true}Object.assign(state,{x:-72,y:0,z:118,vy:0,health:100,deathTime:0,invulnerable:4,interior:null});return false;}
export function finishInvasion(state){if(!state.liberated&&state.enemies.every(e=>e.hp===0)){state.liberated=true;state.score+=2000;return true}return false;}
// Slab ray test, shared by lasers and line-of-sight. Walls stop both sides' shots.
export function rayBox(origin,direction,box,max=150){let near=0,far=max;for(const [axis,lo,hi]of [['x',box.x,box.x+box.w],['y',box.minY,box.maxY],['z',box.z,box.z+box.d]]){const d=direction[axis],o=origin[axis];if(Math.abs(d)<1e-8){if(o<lo||o>hi)return null}else{let a=(lo-o)/d,b=(hi-o)/d;if(a>b)[a,b]=[b,a];near=Math.max(near,a);far=Math.min(far,b);if(near>far)return null}}return near;}
export function raySphere(origin,direction,center,radius,max=150){const x=center.x-origin.x,y=center.y-origin.y,z=center.z-origin.z,t=x*direction.x+y*direction.y+z*direction.z;if(t<0||t>max)return null;const d=x*x+y*y+z*z-t*t;if(d>radius*radius)return null;return Math.max(0,t-Math.sqrt(radius*radius-d));}
export function wallDistance(origin,direction,boxes,max=150){for(const b of boxes){const d=rayBox(origin,direction,b,max);if(d!==null)max=Math.min(max,d)}return max;}
