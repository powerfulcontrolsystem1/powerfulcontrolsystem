import {BaseGame} from './core.js';
import {Renderer,V,triangle,sphere,cylinder,box,transform,hash} from './render3d.js';
const clamp=(v,a,b)=>Math.max(a,Math.min(b,v)),dist=(a,b)=>Math.hypot(a.x-b.x,a.z-b.z);
export class WorldGame extends BaseGame {
 constructor(c,a,id){super(c,a,id);if(id!=='sorpresa')throw Error('Juego retirado');this.ocean=true;this.renderer=new Renderer(c);this.obstacles=[];this.meshes={ball:this.renderer.mesh(sphere()),small:this.renderer.mesh(sphere(7,10)),cylinder:this.renderer.mesh(cylinder()),cone:this.renderer.mesh(cylinder(14,0)),box:this.renderer.mesh(box())};this.buildWorld();this.reset();}
 height(x,z){return -3+Math.sin(x*.17)*.6+Math.cos(z*.21)*.5}
 buildWorld(){const r=this.renderer,data=[],decor=[],foliage=[],ball=sphere(10,14),trunk=cylinder(10,.7);const append=(out,values)=>{for(const v of values)out.push(v)};
  for(let x=-48;x<48;x++)for(let z=-48;z<48;z++){const y=this.height(x,z),shade=.9+hash(x,z)*.1,color=[.79*shade,.78*shade,.55*shade];for(const tri of [[[x,y,z],[x,this.height(x,z+1),z+1],[x+1,this.height(x+1,z),z]],[[x+1,this.height(x+1,z),z],[x,this.height(x,z+1),z+1],[x+1,this.height(x+1,z+1),z+1]]])append(data,triangle(...tri,color))}

   // Coral colonies, sponges, swaying kelp and the remains of an old vessel.
   for(let i=0;i<115;i++){const x=(hash(i,2)-.5)*68,z=(hash(i,3)-.5)*68,y=this.height(x,z),tone=i%4===0?[.92,.49,.33]:i%4===1?[.78,.62,.88]:i%4===2?[.89,.79,.45]:[.3,.68,.55];
    append(decor,transform(ball,[x,y+.3,z],[.8,.5,.7],[0,i,0],[.41,.53,.41]));for(let j=0;j<5;j++){let dx=Math.cos(j*1.2)*.5,dz=Math.sin(j*1.2)*.5,h=.8+hash(i,j)*1.7;append(decor,transform(trunk,[x+dx,y,z+dz],[.10,h,.1],[.25*Math.sin(j),j,.2*Math.cos(j)],tone));append(decor,transform(ball,[x+dx,y+h,z+dz],[.20,.25,.20],[0,0,0],tone));}
    if(i%3===0)for(let j=0;j<3;j++){const h=2+hash(i,j)*3;for(let k=0;k<10;k++){const yy=y+k*h/10,xx=x+Math.sin(k*.5+j)*.2;append(foliage,triangle([xx-.12,yy,z+j*.18],[xx+.12,yy,z+j*.18],[x+Math.sin((k+1)*.5+j)*.2,yy+h/10,z+j*.18],[.18,.48,.24]))}}
   }
   for(let i=0;i<11;i++){const z=-12+i*.65,w=Math.sin((i+.5)/12*Math.PI)*3;append(decor,transform(box(),[8,this.height(8,z)+.7,z],[w,.16,.22],[0,0,.15],[.31,.27,.18]));for(const x of [-w,w])append(decor,transform(trunk,[8+x,-1.8,z],[.12,1.7,.12],[0,0,-x*.16],[.37,.31,.2]));}append(decor,transform(trunk,[8,-2,-8],[.2,8,.2],[0,0,.2],[.40,.34,.23]));

  this.terrain=r.mesh(data);this.decor=r.mesh(decor);this.foliage=r.mesh(foliage);
  r.shadowBegin();r.draw(this.terrain);r.draw(this.decor);r.draw(this.foliage);r.shadowEnd();this.water=r.mesh([...triangle([-70,0,-70],[-70,0,70],[70,0,-70]),...triangle([70,0,-70],[-70,0,70],[70,0,70])]);
 }
 reset(){this.s={score:0,over:false,won:false,time:0,x:0,z:10,y:3,yaw:0,pitch:0,power:100,cooldown:0,treasures:[],oxygen:240,collected:0,sonar:0,hurt:0};this.keys={};for(let i=0;i<10;i++){const a=i*2.399,radius=7+i*1.8;this.s.treasures.push({x:Math.sin(a)*radius,z:Math.cos(a)*radius,found:false})}this.render()}
 async load(p){await super.load(p);this.flow=null;this.flowTime=0}

 look(dx,dy){this.s.yaw-=dx;this.s.pitch=clamp(this.s.pitch-dy,-1.1,1.1)}
 action(k){if(k==='e'&&this.s.power>=50){this.s.power-=50;this.s.sonar=2;this.sound.effect('power')}}
 get info(){const s=this.s;return `${s.collected}/10 tesoros · O₂ ${Math.ceil(s.oxygen)} s`}

 move(entity,dx,dz){entity.x=clamp(entity.x+dx,-36,36);entity.z=clamp(entity.z+dz,-36,36)}





 update(dt){const s=this.s,k=this.keys;if(s.over)return;s.time+=dt;s.cooldown=Math.max(0,s.cooldown-dt);s.hurt=Math.max(0,s.hurt-dt);s.flash=Math.max(0,(s.flash||0)-dt);s.sonar=Math.max(0,s.sonar-dt);s.power=Math.min(100,s.power+dt*3);
  if(k.ArrowLeft)s.yaw+=dt*1.5;if(k.ArrowRight)s.yaw-=dt*1.5;if(k.ArrowUp)s.pitch=clamp(s.pitch+dt,-1.1,1.1);if(k.ArrowDown)s.pitch=clamp(s.pitch-dt,-1.1,1.1);
  let f=(k.w?1:0)-(k.s?1:0),side=(k.d?1:0)-(k.a?1:0),norm=Math.hypot(f,side)||1,speed=4.4*dt/norm;this.move(s,(-Math.sin(s.yaw)*f+Math.cos(s.yaw)*side)*speed,(-Math.cos(s.yaw)*f-Math.sin(s.yaw)*side)*speed);
s.y=clamp(s.y+((k[' ']?1:0)-(k.c?1:0))*dt*2.5,this.height(s.x,s.z)+1,8.5);s.oxygen-=dt;if(s.y>7.7)s.oxygen=Math.min(240,s.oxygen+dt*5);for(const t of s.treasures)if(!t.found&&dist(s,t)<2.1&&s.y<this.height(t.x,t.z)+3.7){t.found=true;s.collected++;s.score+=1000+Math.max(0,Math.floor(s.oxygen));this.sound.effect('score')}if(s.collected===10){s.over=true;s.won=true;s.score+=5000}if(s.oxygen<=0){s.oxygen=0;s.over=true}if(Math.floor(s.time*3)!==Math.floor((s.time-dt)*3)&&hash(Math.floor(s.time*3),8)>.9)this.sound.tone(180+hash(s.time)*220,.4,'sine',.025);this.render();
 }
 part(type,p,scale,color,rot=[0,0,0],kind=0,alpha=1){this.renderer.draw(this.meshes[type],p,scale,color,rot,kind,alpha)}

 fish(x,y,z,size,phase,color=[.72,.83,.82],whale=false){const yaw=Math.sin(phase*.23)*.3,tail=Math.sin(phase*3)*.25;this.part('ball',[x,y,z],[size*(whale?.42:.30),size*.26,size],color,[0,yaw,0]);this.part('ball',[x,y-.05*size,z-size*.55],[size*.22,size*.20,size*.53],whale?[.40,.59,.62]:color,[0,yaw,0]);for(const sign of [-1,1]){this.part('small',[x+sign*size*.25,y+size*.06,z-size*.58],[size*.035,size*.035,size*.035],[.02,.10,.13]);this.part('ball',[x+sign*size*.38,y-size*.13,z],[size*.38,size*.035,size*.17],[color[0]*.75,color[1]*.83,color[2]*.9],[0,sign*.5,sign*.2]);this.part('ball',[x+sign*size*.25+tail*size*.3,y,z+size*.96],[size*.38,size*.045,size*.2],color,[0,sign*.3+tail,0]);}if(whale)for(let i=0;i<7;i++)this.part('ball',[x+(i-3)*size*.053,y-size*.245,z-size*.3],[size*.016,size*.009,size*.5],[.65,.78,.76]);}
 render(){if(!this.renderer||!this.s)return;const s=this.s,eye=[s.x,s.y,s.z],dir=[-Math.sin(s.yaw)*Math.cos(s.pitch),Math.sin(s.pitch),-Math.cos(s.yaw)*Math.cos(s.pitch)],r=this.renderer;r.begin(eye,V.add(eye,dir),s.time,this.ocean);r.draw(this.terrain,[0,0,0],[1,1,1],[1,1,1],[0,0,0],2);r.draw(this.decor,[0,0,0],[1,1,1],[1,1,1],[0,0,0],2);r.draw(this.foliage,[0,0,0],[1,1,1],[1,1,1],[0,0,0],1);
for(let i=0;i<32;i++){const a=i*2.4+s.time*.10,x=Math.sin(a)*(6+i%9)+Math.cos(i)*5,z=Math.cos(a)*(8+i%7)-8,y=1.1+(i%7)*.75+Math.sin(s.time+i)*.2;this.fish(x,y,z,.26+(i%4)*.08,s.time+i,i%3===0?[.94,.75,.25]:i%3===1?[.28,.64,.81]:[.8,.52,.39]);}
   this.fish(Math.sin(s.time*.035)*18-4,5.8,Math.cos(s.time*.035)*13-23,4.5,s.time*.3,[.30,.48,.52],true);
   for(const t of s.treasures)if(!t.found){const y=this.height(t.x,t.z);this.part('ball',[t.x,y+.32,t.z],[.47,.25,.42],[.65,.45,.34]);this.part('ball',[t.x,y+.67+Math.sin(s.time*2)*.07,t.z],[.19,.19,.19],[.9,.94,.8],[0,s.time,0],3);if(s.sonar>0)this.part('ball',[t.x,y+1,t.z],[.5,.5,.5],[.82,1,.68],[0,0,0],3,.18)}
   for(let i=0;i<26;i++){const x=(hash(i,4)-.5)*50,z=(hash(i,5)-.5)*50,y=((s.time*.6+i*.53)%12)-3;this.part('small',[x,y,z],[.035,.045,.035],[.70,.94,.94],[0,0,0],3,.45)}
   r.draw(this.water,[0,10,0],[1,1,1],[.42,.84,.85],[0,0,0],4,.82);

  // Small reticle is DOM so it stays crisp at every rendering resolution.
  if(!this.worldMap){this.worldMap=document.createElement('canvas');this.worldMap.width=120;this.worldMap.height=120;this.worldMap.setAttribute('aria-label','Radar: jugador blanco, tesoros dorados');this.worldMap.style.cssText='position:absolute;right:10px;top:45px;width:62px;height:62px;max-width:62px;pointer-events:none;border:1px solid #e3f6d099;border-radius:50%;background:#10382dcc';this.canvas.parentElement.append(this.worldMap)}const m=this.worldMap.getContext('2d');m.clearRect(0,0,120,120);m.strokeStyle='#c4e3b755';m.beginPath();m.arc(60,60,44,0,Math.PI*2);m.stroke();const dots=s.treasures.filter(t=>!t.found);for(const o of dots){let x=60+(o.x-s.x)*1.7,y=60+(o.z-s.z)*1.7;if(Math.hypot(x-60,y-60)>56)continue;m.fillStyle='#ffea8e';m.beginPath();m.arc(x,y,3,0,Math.PI*2);m.fill()}m.save();m.translate(60,60);m.rotate(-s.yaw);m.fillStyle='#fff9d7';m.beginPath();m.moveTo(0,-7);m.lineTo(-4,5);m.lineTo(4,5);m.closePath();m.fill();m.restore();
  if(!this.reticle){this.reticle=document.createElement('span');this.reticle.textContent='◦';this.reticle.style.cssText='position:absolute;left:50%;top:50%;transform:translate(-50%,-50%);font:26px system-ui;color:#fffbe8;text-shadow:0 1px 3px #17372a;pointer-events:none';this.canvas.parentElement.append(this.reticle)}
 }
 resize(){this.renderer.resize();this.render()}
}
