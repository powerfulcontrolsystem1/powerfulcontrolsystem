import * as THREE from 'three';
import {alienModel,laserRifle,equipSoldier,populateCast} from './characters.js';
import {hitEnemy,hurtPlayer,advanceDeath,finishInvasion,raySphere,wallDistance} from './combat.js';
import {findPath,moveBody,distance,angleTowards} from './simulation.js';

export class Invasion {
 constructor(game){this.game=game;equipSoldier(game.player);this.cast=populateCast(game.city,game.interiors);this.group=new THREE.Group();game.city.group.add(this.group);this.models=[];this.paths=new Map();this.effects=[];this.bolts=[];
  this.viewWeapon=laserRifle();this.viewWeapon.scale.setScalar(.65);this.viewWeapon.position.set(.22,-.24,-.6);this.viewWeapon.rotation.y=Math.PI;game.camera.add(this.viewWeapon);game.scene.add(game.camera);
  this.ghost=new THREE.Group();const m=new THREE.MeshBasicMaterial({color:'#e3ffff',transparent:true,opacity:.65,depthWrite:false});const head=new THREE.Mesh(new THREE.SphereGeometry(.22,14,10),m);head.position.y=1.55;this.ghost.add(head);const body=new THREE.Mesh(new THREE.ConeGeometry(.35,1.1,16),m);body.position.y=.8;this.ghost.add(body);this.ghost.visible=false;game.scene.add(this.ghost);
  this.reticle=document.createElement('div');this.reticle.className='world-reticle';this.reticle.textContent='+';this.reticle.setAttribute('aria-hidden','true');game.canvas.parentElement.append(this.reticle);
 }
 reset(){this.paths.clear();for(const e of this.effects){e.mesh.removeFromParent();e.mesh.geometry.dispose();e.mesh.material.dispose()}this.effects=[];for(const b of this.bolts){b.removeFromParent();b.geometry.dispose();b.material.dispose()}this.bolts=[];this.ghost.visible=false;}
 beam(from,to,color,duration=.12){const d=new THREE.Vector3().subVectors(to,from),mesh=new THREE.Mesh(new THREE.CylinderGeometry(.025,.045,d.length(),6),new THREE.MeshBasicMaterial({color,transparent:true,opacity:.9}));mesh.position.copy(from).addScaledVector(d,.5);mesh.quaternion.setFromUnitVectors(new THREE.Vector3(0,1,0),d.normalize());this.game.scene.add(mesh);this.effects.push({mesh,ttl:duration,max:duration});}
 burst(enemy){const from=new THREE.Vector3(enemy.x,enemy.y+1,enemy.z);for(let i=0;i<10;i++){const to=from.clone().add(new THREE.Vector3(Math.sin(i*2.4)*3,Math.cos(i*1.7)*2,Math.cos(i*2.4)*3));this.beam(from,to,'#ffd47b',.55)}this.game.sound.tone(enemy.kind==='ufo'?65:190,.3,'sawtooth',.07);}
 fire(){const g=this.game,s=g.s;if(s.fireCooldown>0||s.deathTime||s.driving>=0||s.over)return;s.fireCooldown=.2;
  const shoulder=s.cameraMode===0?0:.65,origin=new THREE.Vector3(s.x+Math.cos(s.cameraYaw)*shoulder,s.y+1.6,s.z-Math.sin(s.cameraYaw)*shoulder),dir=new THREE.Vector3(Math.sin(s.cameraYaw)*Math.cos(s.cameraPitch),-Math.sin(s.cameraPitch),Math.cos(s.cameraYaw)*Math.cos(s.cameraPitch));let range=wallDistance(origin,dir,s.interior?g.interiors[s.interior].colliders:g.city.colliders),target;
  if(!s.interior)for(const e of s.enemies){if(e.hp<=0)continue;const hit=raySphere(origin,dir,{x:e.x,y:e.y+1.1,z:e.z},e.kind==='ufo'?3.15:.85,range);if(hit!==null&&hit<range){range=hit;target=e}}
  const end=origin.clone().addScaledVector(dir,range),muzzle=origin.clone().add(new THREE.Vector3(Math.cos(s.cameraYaw)*.25,-.25,-Math.sin(s.cameraYaw)*.25));this.beam(muzzle,end,'#6cffe0');g.sound.tone(740,.12,'sawtooth',.045);s.yaw=s.cameraYaw;
  if(target){if(hitEnemy(s,target)){this.burst(target);g.notify(`${target.kind==='ufo'?'OVNI destruido':'Extraterrestre derrotado'} · +${target.kind==='ufo'?1000:250}`)}else{this.models[target.id]?.scale.setScalar(1.12)}}
  if(finishInvasion(s)){g.notify('¡Springfield liberada! +2000 · Puedes seguir explorando.');g.sound.effect('win');g.updateJournal()}
 }
 update(dt){const g=this.game,s=g.s;s.fireCooldown=Math.max(0,s.fireCooldown-dt);s.invulnerable=Math.max(0,s.invulnerable-dt);
  for(const e of this.effects){e.ttl-=dt;e.mesh.material.opacity=Math.max(0,e.ttl/e.max);if(e.ttl<=0){e.mesh.removeFromParent();e.mesh.geometry.dispose();e.mesh.material.dispose()}}this.effects=this.effects.filter(e=>e.ttl>0);
  if(s.deathTime){const wasInside=s.interior,dying=advanceDeath(s,dt);if(wasInside!==s.interior)g.syncScene();this.render();return dying}
  if(g.keys.f)this.fire();if(s.interior){this.render();return false}
  for(const e of s.enemies){if(e.hp<=0)continue;e.cooldown-=dt;const d=distance(e,s),origin={x:e.x,y:e.y+1.2,z:e.z},delta=new THREE.Vector3(s.x-origin.x,s.y+1.1-origin.y,s.z-origin.z),length=delta.length(),dir=delta.normalize(),visible=wallDistance(origin,dir,g.city.colliders,length)>=length-.1;
   if(e.kind==='ufo'){const a=s.time*.08+e.id*2.1,target={x:Math.sin(a)*135,z:Math.cos(a)*135};e.x+=(target.x-e.x)*dt*.13;e.z+=(target.z-e.z)*dt*.13;e.y=19+Math.sin(s.time*.65+e.id)*2;if(d<100&&visible&&e.cooldown<=0){this.enemyShot(e,dir,origin,30);e.cooldown=2.4}}
   else if(d<55){let path=this.paths.get(e.id);if(!path||path.next<s.time){path={points:findPath(g.nav,e,s),next:s.time+1.2+e.id*.035};this.paths.set(e.id,path)}const target=visible?s:path.points[0];if(target&&d>7){let dx=target.x-e.x,dz=target.z-e.z;const len=Math.hypot(dx,dz)||1;dx=dx/len*2.4;dz=dz/len*2.4;for(const other of s.enemies)if(other!==e&&other.hp>0&&other.kind==='alien'){const od=distance(e,other);if(od<2&&od>.01){dx+=(e.x-other.x)/od;dz+=(e.z-other.z)/od}}moveBody(e,dx*dt,dz*dt,g.city.colliders,.6,185);e.yaw=angleTowards(e.yaw||0,Math.atan2(dx,dz),dt*5);if(distance(e,target)<1)path.points.shift()}if(d<24&&visible&&e.cooldown<=0){this.enemyShot(e,dir,origin,22);e.cooldown=2.1+e.id*.05}}
  }
  for(const p of s.projectiles){const travel=Math.hypot(p.vx,p.vy,p.vz)*dt,origin={x:p.x,y:p.y,z:p.z},dir=new THREE.Vector3(p.vx,p.vy,p.vz).normalize(),wall=wallDistance(origin,dir,g.city.colliders,travel),hit=raySphere(origin,dir,{x:s.x,y:s.y+1,z:s.z},s.driving>=0?1.4:.55,travel);p.ttl-=dt;
   if(hit!==null&&hit<=wall){const died=hurtPlayer(s,20);p.ttl=0;if(died){g.sound.effect('lose');g.notify(s.lives?`Te quedan ${s.lives} vidas`:'Última vida · Springfield te recordará')}else if(s.invulnerable<=.65)g.sound.tone(90,.08,'triangle',.04)}else if(wall<travel-.001)p.ttl=0;p.x+=p.vx*dt;p.y+=p.vy*dt;p.z+=p.vz*dt;
  }s.projectiles=s.projectiles.filter(p=>p.ttl>0);this.render();return !!s.deathTime;
 }
 enemyShot(e,dir,origin,speed){const s=this.game.s;if(s.projectiles.length>=60)return;s.projectiles.push({...origin,vx:dir.x*speed,vy:dir.y*speed,vz:dir.z*speed,ttl:4});if(distance(e,s)<35)this.game.sound.tone(260,.13,'sawtooth',.02)}
 render(){const g=this.game,s=g.s;for(const e of s.enemies){let model=this.models[e.id];if(!model){model=alienModel(e.kind==='ufo');this.models[e.id]=model;this.group.add(model)}model.visible=e.hp>0;model.position.set(e.x,e.y,e.z);model.rotation.y=e.yaw||s.time*.2;model.scale.lerp(new THREE.Vector3(1,1,1),.18);model.userData.animate(s.time+e.id)}
  this.cast.forEach((c,i)=>c.userData.animate?.(s.time+i));for(let i=0;i<Math.max(s.projectiles.length,this.bolts.length);i++){let bolt=this.bolts[i],p=s.projectiles[i];if(!bolt&&p){bolt=new THREE.Mesh(new THREE.SphereGeometry(.13,8,6),new THREE.MeshBasicMaterial({color:'#ff7560'}));this.bolts.push(bolt);this.group.add(bolt)}if(bolt){bolt.visible=!!p;if(p)bolt.position.set(p.x,p.y,p.z)}}
  this.viewWeapon.visible=s.cameraMode===0&&s.driving<0&&!s.deathTime;this.viewWeapon.position.y=-.24+Math.sin(s.time*8)*((g.keys.w||g.keys.s)?.012:.002);this.reticle.hidden=s.driving>=0||!!s.deathTime;this.reticle.classList.toggle('hit',s.invulnerable>0&&s.health<100);
  this.ghost.visible=s.deathTime>.9;this.ghost.position.set(s.x,s.y+Math.max(0,s.deathTime-.9)*1.9,s.z);this.ghost.children.forEach(c=>c.material.opacity=Math.max(0,.7-(s.deathTime-.9)*.18));g.player.root.rotation.z=s.deathTime?-Math.min(Math.PI/2,s.deathTime*2):0;
 }
}
