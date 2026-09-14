import * as THREE from 'three';
import {mergeGeometries} from '../vendor/three/BufferGeometryUtils.js';
import {box,label,material} from './city.js';

function ellipsoid(parent,x,y,z,sx,sy,sz,color){const m=new THREE.Mesh(new THREE.SphereGeometry(1,16,12),material(color));m.position.set(x,y,z);m.scale.set(sx,sy,sz);m.castShadow=true;parent.add(m);return m}
function tube(parent,points,r,color){const curve=new THREE.CatmullRomCurve3(points.map(p=>new THREE.Vector3(...p))),mesh=new THREE.Mesh(new THREE.TubeGeometry(curve,12,r,7,false),material(color));parent.add(mesh);return mesh}
function eyes(parent,y=1.57,z=.28){for(const x of [-.105,.105]){ellipsoid(parent,x,y,z,.115,.13,.065,'#fffdf4');ellipsoid(parent,x,y,z+.06,.038,.046,.02,'#222b29')}}
// Original stylized fan models; the military body is authored in Blender.
export function citizen(name){
 const root=new THREE.Group(),yellow='#f4cb35',blue='#3b83b1';root.name=name;
 const child=['Bart','Lisa','Maggie','Milhouse'].includes(name),fat=['Homero','Barney','Gorgory'].includes(name),female=['Marge','Lisa','Maggie'].includes(name);
 const shirts={Homero:'#f7f5e7',Marge:'#91b940',Bart:'#e66232',Lisa:'#e96036',Maggie:'#82c5dd',Moe:'#929fc3',Burns:'#397e6b',Smithers:'#749444',Gorgory:'#315780',Apu:'#4f915c',Flanders:'#639960',Skinner:'#5381a1',Krusty:'#dfcee8',Milhouse:'#b7a0ce',Barney:'#df8975'};
 const skin=name==='Apu'?'#bc8752':yellow,torso=ellipsoid(root,0,1.02,0,fat?.38:.25,.42,.18,shirts[name]||'#6d9fbb');
 if(female){const dress=new THREE.Mesh(new THREE.ConeGeometry(.34,.7,16),material(shirts[name]));dress.position.y=.74;root.add(dress)}
 const legs=[];for(const side of [-1,1]){const pivot=new THREE.Group();pivot.position.set(side*.13,.67,0);root.add(pivot);box(pivot,0,-.22,0,.17,.46,.18,female?skin:blue);box(pivot,0,-.48,.045,.2,.11,.32,name==='Marge'?'#dd5443':'#40484b');legs.push(pivot)}
 const arms=[];for(const side of [-1,1]){const pivot=new THREE.Group();pivot.position.set(side*(fat?.34:.23),1.25,0);root.add(pivot);ellipsoid(pivot,side*.035,-.2,0,.09,.25,.095,skin);ellipsoid(pivot,side*.035,-.43,.01,.095,.1,.085,skin);arms.push(pivot)}
 ellipsoid(root,0,1.53,0,fat?.29:.24,.3,.23,skin);eyes(root);ellipsoid(root,0,1.48,.29,name==='Burns'?.075:.09,.075,name==='Burns'?.24:.13,skin);
 tube(root,[[-.12,1.34,.19],[0,1.31,.235],[.12,1.34,.19]],.012,'#705334');
 if(name==='Homero'||name==='Barney'){ellipsoid(root,0,1.36,.17,.2,.115,.115,'#c2a778');tube(root,[[-.15,1.76,0],[-.1,1.85,0],[0,1.78,0]],.015,'#504a40');tube(root,[[.02,1.79,0],[.09,1.85,0],[.15,1.76,0]],.015,'#504a40')}
 if(name==='Marge'){for(let i=0;i<8;i++)ellipsoid(root,0,1.8+i*.095,-.015,.235+(i%2)*.015,.14,.23,'#3568ac')}
 if(['Bart','Lisa','Maggie'].includes(name)){for(let i=0;i<(name==='Bart'?7:9);i++){const a=i/(name==='Bart'?7:9)*Math.PI*2,spike=new THREE.Mesh(new THREE.ConeGeometry(.12,.28,4),material(yellow));if(name==='Bart'){spike.position.set((i-3)*.06,1.86,0)}else{spike.position.set(Math.sin(a)*.24,1.6+Math.cos(a)*.24,0);spike.rotation.z=-a}root.add(spike)}}
 if(['Marge','Lisa'].includes(name))for(let i=0;i<9;i++){const a=i/8*Math.PI;ellipsoid(root,Math.cos(a)*.18,1.29-Math.sin(a)*.06,.13+Math.sin(a)*.075,.036,.036,.036,name==='Marge'?'#d9543f':'#fff5da')}
 if(name==='Maggie'){for(const x of [-.08,.08])box(root,x,1.83,.15,.15,.09,.06,'#78b9de');ellipsoid(root,0,1.34,.31,.06,.06,.035,'#dd5e47')}
 if(name==='Moe'){box(root,0,1,.18,.32,.56,.03,'#e5dcc5');for(const x of [-.2,.2])ellipsoid(root,x,1.74,-.06,.12,.12,.18,'#647481')}
 if(name==='Burns'){for(const x of [-.22,.22])ellipsoid(root,x,1.53,-.06,.055,.17,.14,'#a4bbb0');box(root,0,1.2,.18,.06,.24,.02,'#304e4c');torso.scale.x=.2}
 if(['Smithers','Flanders','Milhouse'].includes(name)){for(const x of [-.105,.105]){const ring=new THREE.Mesh(new THREE.TorusGeometry(.123,.013,6,18),material('#514d43'));ring.position.set(x,1.57,.35);root.add(ring)}box(root,0,1.57,.35,.1,.02,.02,'#514d43')}
 if(name==='Flanders')ellipsoid(root,0,1.4,.3,.14,.05,.035,'#72513b');
 if(['Smithers','Flanders','Skinner','Milhouse','Apu'].includes(name)){ellipsoid(root,0,1.8,-.07,.245,.12,.2,name==='Milhouse'?'#4969b4':name==='Apu'?'#353239':'#816344')}
 if(name==='Krusty'){for(const x of [-.28,.28])ellipsoid(root,x,1.58,-.05,.18,.24,.16,'#469487');ellipsoid(root,0,1.48,.37,.085,.085,.07,'#d65b48');box(root,0,1.25,.21,.22,.09,.03,'#80b7d2')}
 if(name==='Gorgory'){ellipsoid(root,0,1.81,0,.3,.08,.26,'#244866');box(root,0,1.79,.24,.36,.035,.25,'#213449');box(root,.12,1.19,.19,.065,.085,.015,'#e7c65d');box(root,0,.79,.02,.69,.09,.37,'#353940')}
 if(child)root.scale.setScalar(name==='Maggie'?.43:.68);
 root.userData={animate(time){arms.forEach((p,i)=>p.rotation.z=Math.sin(time*1.3+i)*.035);root.rotation.y=Math.sin(time*.25)*.12},legs};
 compact(root);return root;
}
export function laserRifle(){const root=new THREE.Group();box(root,0,0,0,.13,.17,.62,'#354c54');box(root,0,.07,.25,.09,.08,.42,'#d2e6dd');box(root,0,-.14,-.12,.09,.2,.11,'#283837');const core=box(root,0,.025,.45,.07,.06,.16,'#57f4dc');core.material=new THREE.MeshStandardMaterial({color:'#63ffe7',emissive:'#28efd1',emissiveIntensity:3});return root}
export function equipSoldier(actor,assets){const root=actor.root,gear=assets.raw('soldier-gear');root.traverse(o=>{if(o.isSkinnedMesh)o.visible=false});root.userData.military=true;gear.name='Militar creado en Blender';root.add(gear);const legs=['L','R'].map(s=>{const node=gear.getObjectByName('SoldierLeg'+s);return {node,x:node?.rotation.x||0}});root.userData.animateMilitary=(name,time)=>{legs.forEach((leg,i)=>{if(leg.node)leg.node.rotation.x=leg.x+(name==='jump'?(i?-.35:.35):name==='run'?Math.sin(time*8+i*Math.PI)*.6:0)})};return gear;
}
export function populateCast(city,interiors){const cast=[];function add(name,parent,x,z,scale=1){const model=citizen(name);model.position.set(x,0,z);model.scale.multiplyScalar(scale);parent.add(model);label(parent,name,x,2.15*scale,z,1.5*scale,'#3b5149','#fff2c9');cast.push(model);return model}
 for(const [name,x,z]of [['Homero',-5,4],['Marge',5,-4],['Bart',-4,1],['Lisa',4,4],['Maggie',-4,4]])add(name,interiors.home.group,x,z);
 add('Moe',interiors.moe.group,-8,-2);add('Barney',interiors.moe.group,-3,3);
 add('Burns',interiors.plant.group,0,-5);add('Smithers',interiors.plant.group,3,-4);
 for(const [name,x,z]of [['Apu',58,24],['Flanders',-99,92],['Skinner',63,-49],['Krusty',128,90],['Milhouse',72,-48]])add(name,city.group,x,z);
 const officer=citizen('Gorgory');officer.scale.setScalar(.72);officer.position.set(.32,.4,.15);officer.userData.legs.forEach(l=>l.rotation.x=-Math.PI/2);city.cars[1].root.add(officer);cast.push(officer);city.cars[1].parked=false;
 return cast;
}
export function alienModel(assets,ufo=false){const root=assets.raw(ufo?'ufo':'alien'),tentacles=[];root.traverse(o=>{if(o.name.includes('Tentacle'))tentacles.push({node:o,x:o.rotation.x,z:o.rotation.z});if(o.isMesh){o.castShadow=true;o.receiveShadow=true}});root.userData.animate=time=>tentacles.forEach((p,i)=>{p.node.rotation.x=p.x+Math.sin(time*3+i)*.16;p.node.rotation.z=p.z+Math.cos(time*2.3+i)*.12});return root;
}

function compact(group){for(const child of [...group.children])if(child.isGroup)compact(child);const batches=new Map();for(const mesh of group.children){if(!mesh.isMesh||mesh.material.transparent||Array.isArray(mesh.material))continue;const key=mesh.material.uuid+Object.keys(mesh.geometry.attributes).join(',');if(!batches.has(key))batches.set(key,[]);batches.get(key).push(mesh)}for(const meshes of batches.values()){if(meshes.length<2)continue;const pieces=meshes.map(m=>{m.updateMatrix();return m.geometry.clone().applyMatrix4(m.matrix)}),geometry=mergeGeometries(pieces);pieces.forEach(g=>g.dispose());if(!geometry)continue;const merged=new THREE.Mesh(geometry,meshes[0].material);merged.castShadow=true;meshes.forEach(m=>m.removeFromParent());group.add(merged)}}
