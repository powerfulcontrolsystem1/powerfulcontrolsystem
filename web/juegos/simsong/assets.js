import * as THREE from 'three';
import {GLTFLoader} from '../vendor/three/GLTFLoader.js';
import {clone} from '../vendor/three/SkeletonUtils.js';

export async function loadAssets(){
 const loader=new GLTFLoader(),assets=new Map();
 const groups={characters:['skaterMaleA','skaterFemaleA','criminalMaleA'],cars:['sedan','police','taxi','van'],suburban:['building-type-a','building-type-b','building-type-c','building-type-e','building-type-g','building-type-j','building-type-m','building-type-q'],furniture:['loungeSofa','televisionVintage','lampRoundFloor','tableCoffee','kitchenFridge','kitchenStove','kitchenSink','kitchenCabinet','table','chair','stoolBar','kitchenBar','bedDouble','bedSingle','desk','bookcaseOpen','bathtub','toilet','bathroomSink','rugRectangle','pottedPlant'],gtas:['soldier-gear','alien','ufo','springfield-world']};
 await Promise.all(Object.entries(groups).flatMap(([group,names])=>names.map(async name=>{
  const gltf=await loader.loadAsync(`/juegos/simsong/assets/${group}/${name}.glb`);
  gltf.scene.traverse(o=>{if(o.isMesh){o.castShadow=true;o.receiveShadow=true}});assets.set(name,gltf);
 })));
 return {
  raw(name){const source=assets.get(name);if(!source)throw Error(`Modelo no disponible: ${name}`);return clone(source.scene)},
  model(name,height){const source=assets.get(name);if(!source)throw Error(`Modelo no disponible: ${name}`);const root=clone(source.scene),box=new THREE.Box3().setFromObject(root),size=box.getSize(new THREE.Vector3()),scale=height/size.y;root.scale.multiplyScalar(scale);root.position.y-=box.min.y*scale;const group=new THREE.Group();group.add(root);return group},
  actor(name){
   const root=this.model(name,1.8),mixer=new THREE.AnimationMixer(root),actions={};
   for(const clip of assets.get(name).animations){const key=['idle','run','jump'].find(k=>clip.name.toLowerCase().includes(k));if(key)actions[key]=mixer.clipAction(clip)}
   const arms=[['LeftArm','LeftForeArm',-1],['RightArm','RightForeArm',1]].map(([a,b,side])=>({bone:root.getObjectByName(a),child:root.getObjectByName(b),side}));
   const position=new THREE.Vector3(),direction=new THREE.Vector3(),desired=new THREE.Vector3(),rotation=new THREE.Quaternion(),parent=new THREE.Quaternion(),world=new THREE.Quaternion(),facing=new THREE.Quaternion();let current;
   return {root,mixer,animate(name,dt){
    const next=actions[name]||actions.idle;if(next&&next!==current){current?.fadeOut(.18);next.reset().fadeIn(.18).play();current=next}mixer.update(dt);root.userData.animateMilitary?.(name,mixer.time);
    // Correct the source FBX shoulder spread after retargeting to the medium mesh.
    // Lower-body motion and forearm articulation remain the imported clips.
    root.updateMatrixWorld(true);root.getWorldQuaternion(facing);
    for(const {bone,child,side}of arms){if(!bone||!child)continue;bone.getWorldPosition(position);child.getWorldPosition(direction).sub(position).normalize();const swing=Math.sin(mixer.time*8)*side*(name==='run'?.65:.025);desired.set(root.userData.military?side*.1:side*.16,root.userData.military?-.45:name==='jump'?-.45:-1,root.userData.military?.85:swing+(name==='jump'?.7:0)).normalize().applyQuaternion(facing);rotation.setFromUnitVectors(direction,desired);bone.getWorldQuaternion(world);bone.parent.getWorldQuaternion(parent).invert();bone.quaternion.copy(parent.multiply(rotation.multiply(world)));bone.updateMatrixWorld(true)}
   }};
  }
 };
}
