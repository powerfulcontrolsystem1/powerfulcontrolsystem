import * as THREE from 'three';
import {mergeGeometries} from '../vendor/three/BufferGeometryUtils.js';
import {LANDMARKS,ROAD_LINES,placeBounds} from './world-data.js';

const mats=new Map();
export function material(color){if(!mats.has(color))mats.set(color,new THREE.MeshStandardMaterial({color,roughness:.82}));return mats.get(color)}
export function box(parent,x,y,z,w,h,d,color){const m=new THREE.Mesh(new THREE.BoxGeometry(w,h,d),material(color));m.position.set(x,y,z);m.castShadow=true;m.receiveShadow=true;parent.add(m);return m}
export function label(parent,text,x,y,z,width=12,color='#185763',background='#fff0bc'){
 const c=document.createElement('canvas');c.width=768;c.height=128;const ctx=c.getContext('2d');ctx.fillStyle=background;ctx.fillRect(0,0,768,128);ctx.fillStyle=color;ctx.textAlign='center';ctx.textBaseline='middle';ctx.font=`bold ${Math.min(66,1100/text.length)}px sans-serif`;ctx.fillText(text,384,67,740);const texture=new THREE.CanvasTexture(c);texture.colorSpace=THREE.SRGBColorSpace;const m=new THREE.Mesh(new THREE.PlaneGeometry(width,width/6),new THREE.MeshBasicMaterial({map:texture,side:THREE.DoubleSide}));m.position.set(x,y,z);parent.add(m);return m;
}
function cylinder(parent,x,y,z,r1,r2,h,color,segments=16){const m=new THREE.Mesh(new THREE.CylinderGeometry(r1,r2,h,segments),material(color));m.position.set(x,y,z);m.castShadow=true;m.receiveShadow=true;parent.add(m);return m}
function batchStatic(group,cars){
 const dynamic=new Set(cars.map(c=>c.root)),batches=new Map();group.updateMatrixWorld(true);
 group.traverse(mesh=>{if(!mesh.isMesh||mesh.isInstancedMesh||Array.isArray(mesh.material)||mesh.material.transparent)return;for(let p=mesh;p;p=p.parent)if(dynamic.has(p))return;
  const key=mesh.material.uuid+':'+mesh.castShadow+':'+mesh.receiveShadow+':'+Object.keys(mesh.geometry.attributes).sort().join(',')+':'+Boolean(mesh.geometry.index);
  if(!batches.has(key))batches.set(key,[]);batches.get(key).push(mesh);
 });
 for(const meshes of batches.values()){if(meshes.length<2)continue;const geometries=meshes.map(m=>m.geometry.clone().applyMatrix4(m.matrixWorld)),geometry=mergeGeometries(geometries);geometries.forEach(g=>g.dispose());if(!geometry)continue;const merged=new THREE.Mesh(geometry,meshes[0].material);merged.castShadow=meshes[0].castShadow;merged.receiveShadow=meshes[0].receiveShadow;meshes.forEach(m=>m.removeFromParent());group.add(merged)}
}
function roof(parent,w,d,y,color){const shape=new THREE.Shape();shape.moveTo(-w/2,0);shape.lineTo(w/2,0);shape.lineTo(0,w*.27);shape.closePath();const m=new THREE.Mesh(new THREE.ExtrudeGeometry(shape,{depth:d,bevelEnabled:false}),material(color));m.position.set(0,y,-d/2);m.castShadow=true;parent.add(m)}
function windows(parent,w,d,h,color='#b9edf2'){for(let x=-w/2+3;x<w/2-1;x+=5){box(parent,x,h*.58,d/2+.09,2.2,2.3,.15,'#f2e2b2');box(parent,x,h*.58,d/2+.19,1.85,1.95,.12,color);box(parent,x,h*.58,d/2+.28,.1,1.9,.12,'#e6f2dd')}}
function landmark(parent,p){
 const g=new THREE.Group();g.position.set(p.x,0,p.z);parent.add(g);const {w,d,h,kind}=p;
 const colors={home:'#dcaa79',flanders:'#db91b3',moe:'#75688d',toot:'#d6bd72',kwik:'#e7dbbf',school:'#cb7058',police:'#93aebb',hall:'#e8dab2',church:'#d4c5aa',krusty:'#efd369',lard:'#e3d6b4',plant:'#bed4b1'};
 box(g,0,h/2,0,w,h,d,colors[kind]);windows(g,w,d,h);
 if(['home','flanders'].includes(kind)){
  roof(g,w+1,d+1,h,kind==='home'?'#a65e48':'#714d70');box(g,w*.3,h+1.8,-2,1.4,4,1.4,'#ac6d59');box(g,-w*.26,2,d/2+.2,6,3.8,.3,'#9a6951');for(let y=.5;y<3.7;y+=.7)box(g,-w*.26,y,d/2+.4,5.8,.06,.12,'#dbb18a');box(g,1,1.7,d/2+.25,1.8,3.4,.4,'#744b40');box(g,0,.08,d/2+3,w*.92,.15,5,'#ddcba9');
 }else if(kind==='moe'){
  box(g,0,1.8,d/2+.25,w,.4,.4,'#b08959');for(const x of [-5,5]){box(g,x,2.6,d/2+.25,4.5,2.6,.3,'#66845c');for(let k=-2;k<=2;k++){const s=box(g,x+k*.8,2.6,d/2+.45,.12,3,.1,'#b4c58d');s.rotation.z=.55}}
  label(g,"MOE’S TAVERN",0,5.7,d/2+.5,w-1,'#a43c3b','#e6cf8f');
 }else if(kind==='plant'){
  box(g,0,5,0,w,10,d,'#acc3a0');for(const x of [-16,16]){const points=[];for(let i=0;i<=20;i++){const y=i*1.6;points.push(new THREE.Vector2(6.4+Math.pow((y-20)/15,2)*2.7,y))}const tower=new THREE.Mesh(new THREE.LatheGeometry(points,32),material('#e5ddc7'));tower.position.set(x,0,0);tower.castShadow=true;g.add(tower)}label(g,'SPRINGFIELD NUCLEAR',0,8,d/2+.2,36);
 }else if(kind==='church'){
  roof(g,w+1,d+1,h,'#748391');box(g,0,15,d/2-4,7,16,7,'#e5d9bc');cylinder(g,0,27,d/2-4,0,5,12,'#788e97',4);box(g,0,34,d/2-4,.45,5,.45,'#fff3cc');box(g,0,34.7,d/2-4,2.5,.4,.45,'#fff3cc');
 }else if(kind==='hall'){
  for(let x=-10;x<=10;x+=5)cylinder(g,x,6,d/2+2,.7,.85,12,'#f4eacb');box(g,0,12.8,d/2+2,w+3,1.5,5,'#f5e7c5');roof(g,w+3,6,13.5,'#d3b786');label(g,'SPRINGFIELD',0,11,d/2+4.6,19);
 }else if(kind==='school'){
  box(g,0,12,d/2-4,10,8,9,'#d98a64');cylinder(g,0,17,d/2-4,0,7,5,'#817c72',4);label(g,'SPRINGFIELD ELEMENTARY',0,8,d/2+.3,w-3);
 }else if(kind==='lard'){
  const donut=new THREE.Mesh(new THREE.TorusGeometry(5,1.7,12,40),material('#e49ab3'));donut.position.set(0,h+6,0);g.add(donut);label(g,'LARD LAD DONUTS',0,h-1,d/2+.3,w-1);
 }else{const title={toot:"KING TOOT’S",kwik:'KWIK-E-MART',police:'SPRINGFIELD POLICE',krusty:'KRUSTY BURGER'}[kind];box(g,0,h-.5,d/2+.5,w+1,1.2,1.4,kind==='kwik'?'#dd6652':'#566e89');if(title)label(g,title,0,h+1,d/2+.4,w-1)}
 if(p.door)label(g,'ENTRAR · E',1,2,p.d/2+.6,2,'#ffffff','#3e726d');return g;
}
export function buildCity(assets){
 const group=new THREE.Group(),colliders=LANDMARKS.map(placeBounds),cars=[],pedestrianSpawns=[];
 const ground=new THREE.Mesh(new THREE.CircleGeometry(225,96),material('#97bd79'));ground.rotation.x=-Math.PI/2;ground.receiveShadow=true;group.add(ground);
 const blenderWorld=assets.raw('springfield-world');blenderWorld.name='Springfield detallado en Blender';blenderWorld.traverse(o=>{if(o.isMesh){o.castShadow=true;o.receiveShadow=true}});group.add(blenderWorld);
 for(const axis of [0,1])for(const line of ROAD_LINES){const road=box(group,axis?0:line,.015,axis?line:0,axis?372:11,.03,axis?11:372,'#68747b');road.castShadow=false;for(const side of [-1,1]){const sidewalk=box(group,axis?0:line+side*7,.04,axis?line+side*7:0,axis?372:2.5,.08,axis?2.5:372,'#d7d4c0');sidewalk.castShadow=false}for(let i=-177;i<180;i+=9){if(ROAD_LINES.some(l=>Math.abs(i-l)<9))continue;box(group,axis?i:line,.04,axis?line:i,axis?4:.16,.02,axis?.16:4,'#f4e4a4')}}
 for(const x of ROAD_LINES)for(const z of ROAD_LINES)for(let i=-4;i<=4;i+=2){box(group,x+i,.05,z+8,1,.03,2.5,'#f8f3db');box(group,x+8,.05,z+i,2.5,.03,1,'#f8f3db')}
 LANDMARKS.forEach(p=>landmark(group,p));
 const houseNames=['building-type-a','building-type-b','building-type-c','building-type-e','building-type-g','building-type-j','building-type-m','building-type-q'];let count=0;
 for(const x of [-150,-130,-88,-64,-16,8,57,80,136,157])for(const z of [-139,-89,-65,-17,7,57,80,140]){
  if(Math.hypot(x,z)>176||colliders.some(b=>x>b.x-13&&x<b.x+b.w+13&&z>b.z-13&&z<b.z+b.d+13)||ROAD_LINES.some(l=>Math.abs(x-l)<16||Math.abs(z-l)<16))continue;
  const model=assets.model(houseNames[count++%houseNames.length],7+count%4);const b=new THREE.Box3().setFromObject(model),s=b.getSize(new THREE.Vector3());model.position.set(x,0,z);group.add(model);colliders.push({x:x-s.x/2,z:z-s.z/2,w:s.x,d:s.z,minY:0,maxY:12});pedestrianSpawns.push({x:x,z:z+s.z/2+2});
 }
 // Shared instanced geometry keeps foliage inexpensive on mobile.
 const spots=[];for(let i=0;i<240;i++){const angle=i*2.399963,r=35+Math.sqrt((i*71)%1000)/31.6*150,x=Math.cos(angle)*r,z=Math.sin(angle)*r;if(ROAD_LINES.some(l=>Math.abs(l-x)<10||Math.abs(l-z)<10)||colliders.some(b=>x>b.x-3&&x<b.x+b.w+3&&z>b.z-3&&z<b.z+b.d+3))continue;spots.push({x,z})}
 const trunks=new THREE.InstancedMesh(new THREE.CylinderGeometry(.3,.5,3,6),material('#926a45'),spots.length),leaves=new THREE.InstancedMesh(new THREE.IcosahedronGeometry(2.8,1),material('#518b56'),spots.length),dummy=new THREE.Object3D();spots.forEach((p,i)=>{dummy.position.set(p.x,1.5,p.z);dummy.scale.set(1,1,1);dummy.updateMatrix();trunks.setMatrixAt(i,dummy.matrix);dummy.position.y=4.7;dummy.scale.set(1,1.3,1);dummy.updateMatrix();leaves.setMatrixAt(i,dummy.matrix)});trunks.castShadow=true;leaves.castShadow=true;group.add(trunks,leaves);
 for(let i=0;i<8;i++){const model=assets.model(['sedan','police','taxi','van'][i%4],1.7),x=i===0?-65:ROAD_LINES[i%4]+2.6,z=i===0?101:65-i*19;model.position.set(x,0,z);group.add(model);cars.push({root:model,x,z,y:0,yaw:i===0?Math.PI/2:0,speed:0,parked:i===0||i===1,turn:0});if(i===1)label(model,'POLICE',0,1.5,.4,1.5,'#24497b','#ffffff')}
 // A visible hemisphere and a physical circular limit share the same centre.
 const dome=new THREE.Mesh(new THREE.SphereGeometry(200,64,32,0,Math.PI*2,0,Math.PI/2),new THREE.MeshPhysicalMaterial({color:'#b4e8f1',transparent:true,opacity:.075,roughness:.15,metalness:.05,side:THREE.DoubleSide,depthWrite:false}));dome.scale.y=.8;group.add(dome);
 for(let i=0;i<8;i++){const points=[];for(let j=0;j<=64;j++){const a=j*Math.PI/64;points.push(new THREE.Vector3(Math.cos(a)*200,Math.sin(a)*160,0))}const arc=new THREE.Line(new THREE.BufferGeometry().setFromPoints(points),new THREE.LineBasicMaterial({color:'#a4d5e1',transparent:true,opacity:.23}));arc.rotation.y=i*Math.PI/8;group.add(arc)}
 const ring=new THREE.Mesh(new THREE.TorusGeometry(198,.3,6,128),material('#a8d0d7'));ring.rotation.x=Math.PI/2;group.add(ring);
 for(let i=0;i<16;i++){const angle=i*Math.PI/8;const hill=new THREE.Mesh(new THREE.SphereGeometry(45,12,8),material(i%2?'#95b6a1':'#adc8ab'));hill.scale.y=.6;hill.position.set(Math.cos(angle)*250,-8,Math.sin(angle)*250);group.add(hill)}
 // The Blender vegetation is decorative; omit trees whose canopy intersects a building.
 blenderWorld.updateMatrixWorld(true);const blockedTrees=[];blenderWorld.traverse(o=>{if(o.name.startsWith('Arbol')){const p=new THREE.Vector3();o.getWorldPosition(p);if(colliders.some(b=>p.x>b.x-3&&p.x<b.x+b.w+3&&p.z>b.z-3&&p.z<b.z+b.d+3))blockedTrees.push(o)}});blockedTrees.forEach(o=>o.removeFromParent());
 batchStatic(group,cars);return {group,colliders,cars,pedestrianSpawns};
}
