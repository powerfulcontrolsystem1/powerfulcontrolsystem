import * as THREE from 'three';
import {box,label,material} from './city.js';
export function buildInterior(id,assets){
 const group=new THREE.Group(),colliders=[],home=id==='home';
 function wall(x,z,w,d,minY=0,h=3.3){box(group,x,minY+h/2,z,w,h,d,home?'#edcbaa':'#ac9a86');colliders.push({x:x-w/2,z:z-d/2,w,d,minY,maxY:minY+h})}
 function furniture(name,x,z,height,y=0,angle=0){const model=assets.model(name,height);model.position.set(x,y,z);model.rotation.y=angle;group.add(model);return model}
 box(group,0,home?6.8:3.4,0,20,.2,18,'#e9ddc6');box(group,0,-.12,0,20,.24,18,home?'#d8a0a1':'#907155');wall(-10,0,.3,18);wall(10,0,.3,18);wall(0,-9,20,.3);wall(-5.75,9,8.5,.3);wall(5.75,9,8.5,.3);
 label(group,'SALIR · E',0,2.1,8.95,2.8,'#ffffff','#3d746d');
 // Windows and warm fill light keep interiors readable on compact displays.
 for(const x of [-6,6]){box(group,x,2,-8.8,4,1.7,.06,'#b8ecf2');box(group,x,2,8.8,3,1.7,.06,'#c4eef0')}
 const light=new THREE.PointLight('#fff4de',25,32,1.4);light.position.set(0,5,1);group.add(light);
 if(home){
  // Stairwell stays open on both floors. Room partitions leave wide doorways.
  wall(-1.7,-5,.2,8);wall(1.7,-5,.2,8);wall(-8.5,0,3,.2);wall(-2.5,0,1,.2);wall(8.5,0,3,.2);wall(2.5,0,1,.2);
  furniture('loungeSofa',-6,4,1.35,0,Math.PI);furniture('tableCoffee',-6,2,.6);furniture('televisionVintage',-8,1,1.35,0,Math.PI/2);furniture('lampRoundFloor',-8,6,2.2);furniture('rugRectangle',-6,3,.03);
  label(group,'⛵',-6,2.2,8.7,2.2,'#346681','#efcf91');
  furniture('table',6,4,1);for(const p of [[4.5,4],[7.5,4],[6,2.5],[6,5.5]])furniture('chair',p[0],p[1],1.2);
  for(let x=2;x<10;x++)for(let z=-8;z<0;z++)box(group,x+.5,.015,z+.5,1,.025,1,(x+z)%2?'#eee8ce':'#76a4ab');
  furniture('kitchenFridge',8,-7,2.5);furniture('kitchenStove',5,-7,1.1);furniture('kitchenSink',3,-7,1.1);furniture('kitchenCabinet',6,-8,1,1.7);
  furniture('bookcaseOpen',-8,-7,2.6);furniture('desk',-6,-6,1.1);
  // Walkable stairs: continuous support ramp matches the step tops.
  for(let i=0;i<17;i++)box(group,0,(i+1)*.1,6-i*.65,3.1,(i+1)*.2,.66,'#b48662');
  for(const x of [-5.8,5.8])box(group,x,3.3,0,8.4,.2,18,'#d9baa2');box(group,0,3.3,-7.9,3.2,.2,2.1,'#d9baa2');
  for(const x of [-10,10])wall(x,0,.3,18,3.4);wall(0,-9,20,.3,3.4);wall(0,9,20,.3,3.4);
  for(const x of [-5.8,5.8])wall(x,0,8.4,.2,3.4); // Enter rooms around the ends beside the stairwell.
  furniture('bedDouble',-6,5,1.3,3.4);furniture('bedSingle',6,5,1.2,3.4);furniture('bedSingle',6,-5,1.2,3.4);furniture('desk',8,-7,1.1,3.4);furniture('bathtub',-7,-6,1.1,3.4);furniture('toilet',-4,-7,1.1,3.4);furniture('bathroomSink',-8,-3,1.2,3.4);
  label(group,'742 EVERGREEN TERRACE',0,2.9,-8.7,3);
 }else if(id==='plant'){
  furniture('desk',0,-4,1.3);furniture('chair',0,-6,1.5);furniture('bookcaseOpen',-7,-7,2.6);furniture('loungeSofa',6,3,1.35);furniture('pottedPlant',-7,5,2);
  label(group,'CENTRAL NUCLEAR · DIRECCIÓN',0,2.7,-8.7,9,'#25584f','#f0e2b8');
  for(const x of [-5,5]){box(group,x,1.2,-7,2,2.4,.4,'#6f9e9b');for(let i=0;i<3;i++)box(group,x+(i-1)*.45,1.5,-6.75,.15,.15,.05,'#e7ce54')}
 }else{
  box(group,-6,1.1,-2,2.2,2.2,10,'#755642');colliders.push({x:-7.1,z:-7,w:2.2,d:10,minY:0,maxY:2.2});
  for(let z=-6;z<=3;z+=2.1)furniture('stoolBar',-3.6,z,1.2);
  for(let z=-7;z<=6;z+=2){furniture('table',5,z,1);furniture('chair',3.5,z,1.15);furniture('chair',6.5,z,1.15)}
  for(let i=0;i<18;i++){const bottle=box(group,-8.8,1.8+(i%3)*.65,-7+Math.floor(i/3)*1.3,.2,.5,.2,['#50735a','#bca451','#8296ab'][i%3]);bottle.rotation.z=.03}
  box(group,0,.9,-5,3.2,1.8,5,'#69523f');box(group,0,1.83,-5,2.9,.06,4.7,'#347d68');
  label(group,"MOE’S TAVERN",0,2.8,-8.75,8,'#ddbb6c','#594c6a');label(group,'DUFF',-8.75,2.8,0,2,'#a84337','#ead6b2').rotation.y=Math.PI/2;
 }
 return {group,colliders,spawn:{x:0,y:0,z:7},floorAt(s){if(!home)return 0;if(Math.abs(s.x)<1.65&&s.z<6.5&&s.z>-5.2)return Math.max(0,Math.min(3.4,(6.4-s.z)/.65*.2));if(s.y>3.15&&(Math.abs(s.x)>1.65||s.z<-5.2))return 3.4;return 0}};
}
