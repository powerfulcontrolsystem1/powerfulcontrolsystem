import test from 'node:test';
import assert from 'node:assert/strict';
import {moveBody,gravity,createNavigation,findPath,pedestrianSteering,intersects} from '../web/juegos/simsong/simulation.js';
import {LANDMARKS,MISSIONS,placeBounds} from '../web/juegos/simsong/world-data.js';
test('Fast movement cannot cross buildings or the dome; blocked motion can slide',()=>{
 const wall={x:2,z:-4,w:1,d:8,minY:0,maxY:10},body={x:0,z:0,y:0};assert.equal(moveBody(body,20,2,[wall]),true);assert.ok(body.x<1.67);assert.ok(body.z>1.9);
 const edge={x:197,z:0,y:0};moveBody(edge,50,0,[]);assert.ok(Math.hypot(edge.x,edge.z)<198);assert.ok(edge.x>=197);
});
test('Jump rises and lands without a second jump in the air',()=>{
 const body={y:0,vy:0};gravity(body,.016,0,true);const first=body.vy;gravity(body,.016,0,true);assert.ok(body.vy<first);let highest=0;for(let i=0;i<120;i++){gravity(body,.016,0);highest=Math.max(highest,body.y)}assert.ok(highest>1);assert.equal(body.y,0);assert.equal(body.vy,0);
});
test('NPC A* routes around walls and preserves impassable cells',()=>{
 const walls=[{x:-2,z:-15,w:4,d:24,minY:0,maxY:4}],nav=createNavigation(walls,25,2),path=findPath(nav,{x:-10,z:0},{x:10,z:0});assert.ok(path.length>12);assert.ok(path.every(p=>!walls.some(b=>intersects(p.x,p.z,0,b,.65))));assert.deepEqual(findPath(nav,{x:0,z:0},{x:10,z:0}),[]);
});
test('Pedestrians yield to approaching traffic and separate from neighbors',()=>{
 const p={x:0,z:0},target={x:0,z:10};const normal=pedestrianSteering(p,target,[],[]);const yieldSpeed=pedestrianSteering(p,target,[],[{x:0,z:-4,yaw:0,speed:5}]);assert.ok(Math.hypot(yieldSpeed.x,yieldSpeed.z)<Math.hypot(normal.x,normal.z)/10);assert.equal(p.behavior,'cediendo el paso');const separate=pedestrianSteering(p,target,[{x:.3,z:0}],[]);assert.ok(separate.x<0);
});
test('Both interior entrances are outside walls; missions have no required ordering',()=>{
 for(const p of LANDMARKS.filter(p=>p.door))assert.equal(intersects(p.door.x,p.door.z,0,placeBounds(p)),false,p.id);const targets=MISSIONS.flatMap(m=>m.targets);assert.ok(targets.includes('inside-home'));assert.ok(targets.includes('inside-moe'));assert.ok(LANDMARKS.some(p=>p.id==='plant'));
});
