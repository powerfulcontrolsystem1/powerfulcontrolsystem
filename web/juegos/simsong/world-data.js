// World units are metres. This layout is an original synthesis of Springfield references.
export const WORLD_RADIUS=198;
export const ROAD_LINES=[-112,-40,32,104];
export const LANDMARKS=[
 {id:'home',name:'Casa de Homero · 742 Evergreen Terrace',x:-73,z:80,w:21,d:19,h:8,kind:'home',door:{x:-72,z:91}},
 {id:'flanders',name:'Casa de Flanders',x:-101,z:80,w:19,d:18,h:8,kind:'flanders'},
 {id:'moe',name:'Taberna de Moe',x:-65,z:13,w:18,d:17,h:6,kind:'moe',door:{x:-65,z:23}},
 {id:'toot',name:"King Toot’s Music Store",x:-86,z:13,w:18,d:17,h:6,kind:'toot'},
 {id:'kwik',name:'Kwik-E-Mart',x:58,z:10,w:29,d:22,h:7,kind:'kwik'},
 {id:'school',name:'Escuela Primaria de Springfield',x:63,z:-67,w:41,d:30,h:11,kind:'school'},
 {id:'police',name:'Policía de Springfield',x:-12,z:-66,w:27,d:27,h:11,kind:'police'},
 {id:'hall',name:'Ayuntamiento',x:0,z:3,w:29,d:27,h:13,kind:'hall'},
 {id:'church',name:'Primera Iglesia de Springfield',x:65,z:70,w:26,d:32,h:14,kind:'church'},
 {id:'krusty',name:'Krusty Burger',x:131,z:75,w:27,d:23,h:7,kind:'krusty'},
 {id:'lard',name:'Lard Lad Donuts',x:135,z:5,w:24,d:25,h:7,kind:'lard'},
 {id:'plant',name:'Central Nuclear de Springfield',x:-61,z:-149,w:63,d:39,h:16,kind:'plant',door:{x:-61,z:-127}}
];
export const MISSIONS=[
 {id:'neighbors',title:'Un día por el barrio',description:'Visita la casa de Homero, Moe y el Kwik-E-Mart, en el orden que prefieras.',targets:['home','moe','kwik'],reward:500},
 {id:'city',title:'Conoce Springfield',description:'Descubre la escuela, la comisaría y la central nuclear.',targets:['school','police','plant'],reward:750},
 {id:'interiors',title:'Puertas abiertas',description:'Entra en la casa y en la taberna; recorre sus interiores.',targets:['inside-home','inside-moe'],reward:600}
];
import {combatState} from './combat.js';
export function createState(){return {...combatState(),worldVersion:1,score:0,over:false,won:false,time:0,x:-72,y:0,z:118,vy:0,yaw:Math.PI,cameraYaw:Math.PI,cameraPitch:.18,health:100,interior:null,driving:-1,carSpeed:0,firstPerson:false,visited:[],completed:[],carStates:[]}}
export function placeBounds(place){return {x:place.x-place.w/2,z:place.z-place.d/2,w:place.w,d:place.d,minY:0,maxY:place.h}}
export function nearestPlace(x,z){return LANDMARKS.reduce((best,p)=>Math.hypot(p.x-x,p.z-z)<Math.hypot(best.x-x,best.z-z)?p:best,LANDMARKS[0])}
