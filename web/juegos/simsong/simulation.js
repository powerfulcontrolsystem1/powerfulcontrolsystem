// Pure simulation: reusable by physics tests without WebGL or a browser.
export const clamp=(n,a,b)=>Math.max(a,Math.min(b,n));
export const distance=(a,b)=>Math.hypot(a.x-b.x,a.z-b.z);
export function intersects(x,z,y,box,r=.34,height=1.7){return y+height>box.minY+.08&&y<box.maxY-.05&&x+r>box.x&&x-r<box.x+box.w&&z+r>box.z&&z-r<box.z+box.d}
export function moveBody(body,dx,dz,boxes,r=.34,boundary=198){
 const steps=Math.max(1,Math.ceil(Math.hypot(dx,dz)/.25));let blocked=false;
 for(let i=0;i<steps;i++){const x=body.x+dx/steps,z=body.z+dz/steps;
  if(Math.hypot(x,body.z)<boundary-r&&!boxes.some(b=>intersects(x,body.z,body.y||0,b,r)))body.x=x;else blocked=true;
  if(Math.hypot(body.x,z)<boundary-r&&!boxes.some(b=>intersects(body.x,z,body.y||0,b,r)))body.z=z;else blocked=true;
 }return blocked;
}
export function gravity(body,dt,floor,jump=false){
 const grounded=body.y<=floor+.055&&body.vy<=0;if(jump&&grounded)body.vy=6.4;
 body.vy-=18*dt;body.y+=body.vy*dt;
 if(body.y<=floor){body.y=floor;body.vy=0}return grounded;
}
export function angleTowards(a,b,t){let d=(b-a+Math.PI)%(Math.PI*2);if(d<0)d+=Math.PI*2;return a+(d-Math.PI)*Math.min(1,t)}
// A* uses a world grid only for NPC planning. No route is displayed to the player.
export function createNavigation(boxes,radius=178,cell=5){
 const size=Math.floor(radius*2/cell)+1,grid=new Uint8Array(size*size);
 const world=(x,z)=>({x:x*cell-radius,z:z*cell-radius});
 for(let z=0;z<size;z++)for(let x=0;x<size;x++){const p=world(x,z);grid[z*size+x]=Math.hypot(p.x,p.z)<radius&&!boxes.some(b=>intersects(p.x,p.z,0,b,.65))?1:0}
 return {grid,size,radius,cell,world};
}
export function findPath(nav,start,target){
 const {size,cell,radius,grid,world}=nav,toIndex=p=>clamp(Math.round((p.z+radius)/cell),0,size-1)*size+clamp(Math.round((p.x+radius)/cell),0,size-1);
 const source=toIndex(start),dest=toIndex(target);if(!grid[source]||!grid[dest])return [];
 const open=[source],parents=new Int32Array(grid.length).fill(-1),g=new Float32Array(grid.length).fill(Infinity),closed=new Uint8Array(grid.length);g[source]=0;
 const heuristic=i=>Math.abs(i%size-dest%size)+Math.abs(Math.floor(i/size)-Math.floor(dest/size));
 for(let rounds=0;open.length&&rounds<6000;rounds++){
  let best=0;for(let i=1;i<open.length;i++)if(g[open[i]]+heuristic(open[i])<g[open[best]]+heuristic(open[best]))best=i;
  const current=open.splice(best,1)[0];if(current===dest){const path=[];for(let i=dest;i!==source;i=parents[i]){if(i<0)return [];path.push(world(i%size,Math.floor(i/size)))}return path.reverse()}
  closed[current]=1;const x=current%size,z=Math.floor(current/size);
  for(const [dx,dz]of [[1,0],[-1,0],[0,1],[0,-1]]){const xx=x+dx,zz=z+dz,n=zz*size+xx;if(xx<0||zz<0||xx>=size||zz>=size||!grid[n]||closed[n])continue;const cost=g[current]+1;if(cost<g[n]){parents[n]=current;g[n]=cost;if(!open.includes(n))open.push(n)}}
 }return [];
}
// Reynolds-style arrival + separation + predicted traffic avoidance.
export function pedestrianSteering(agent,target,neighbors,cars){
 let dx=target.x-agent.x,dz=target.z-agent.z,d=Math.hypot(dx,dz),speed=1.25;dx/=d||1;dz/=d||1;
 for(const car of cars){const ahead={x:car.x+Math.sin(car.yaw)*(car.speed||0)*.8,z:car.z+Math.cos(car.yaw)*(car.speed||0)*.8};if(distance(agent,ahead)<5.5&&distance(agent,car)<9){speed=.05;agent.behavior='cediendo el paso'}}
 for(const other of neighbors){if(other===agent)continue;const sep=distance(agent,other);if(sep>0&&sep<2.1){dx+=(agent.x-other.x)/(sep*sep)*.8;dz+=(agent.z-other.z)/(sep*sep)*.8;agent.behavior='evitando peatón'}}
 const length=Math.hypot(dx,dz)||1;return {x:dx/length*speed*Math.min(1,d),z:dz/length*speed*Math.min(1,d)};
}
