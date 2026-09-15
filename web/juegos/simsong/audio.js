// Original procedural soundtrack and effects. Nothing is streamed or downloaded.
export class GTASAudio{
 constructor(sound){this.sound=sound;this.reset()}
 reset(){this.stop();this.beat=0;this.musicTime=.05;this.nearTime=1.5}
 track(source,...nodes){this.nodes??=new Set();this.nodes.add(source);source.onended=()=>{source.disconnect();nodes.forEach(n=>n.disconnect());this.nodes.delete(source)}}
 stop(){for(const node of this.nodes||[])try{node.stop()}catch{}this.nodes?.clear()}
 synth(freq,duration=.2,type='sine',volume=.02,end=freq){const ctx=this.sound.ctx;if(!this.sound.enabled||!ctx||ctx.state!=='running')return;const now=ctx.currentTime,o=ctx.createOscillator(),gain=ctx.createGain();o.type=type;o.frequency.setValueAtTime(Math.max(25,freq),now);o.frequency.exponentialRampToValueAtTime(Math.max(25,end),now+duration);gain.gain.setValueAtTime(volume,now);gain.gain.exponentialRampToValueAtTime(.0001,now+duration);o.connect(gain);gain.connect(ctx.destination);this.track(o,gain);o.start(now);o.stop(now+duration)}
 noise(duration=.18,volume=.025,frequency=900){const ctx=this.sound.ctx;if(!this.sound.enabled||!ctx||ctx.state!=='running')return;const buffer=ctx.createBuffer(1,Math.ceil(ctx.sampleRate*duration),ctx.sampleRate),data=buffer.getChannelData(0);for(let i=0;i<data.length;i++)data[i]=(Math.random()*2-1)*(1-i/data.length);const source=ctx.createBufferSource(),filter=ctx.createBiquadFilter(),gain=ctx.createGain();filter.type='bandpass';filter.frequency.value=frequency;filter.Q.value=2.5;gain.gain.setValueAtTime(volume,ctx.currentTime);gain.gain.exponentialRampToValueAtTime(.0001,ctx.currentTime+duration);source.buffer=buffer;source.connect(filter);filter.connect(gain);gain.connect(ctx.destination);this.track(source,filter,gain);source.start();source.stop(ctx.currentTime+duration)}
 update(dt,near){this.musicTime-=dt;this.nearTime-=dt;if(this.musicTime<=0){const melody=[220,277,330,415,330,277,247,330,220,277,370,330,247,220,185,247],note=melody[this.beat%melody.length];this.synth(note,.25,this.beat%4?'triangle':'sine',.012,note*.99);if(this.beat%4===0)this.synth(note/2,.52,'sine',.015,note/2);this.beat++;this.musicTime=.28}if(near<38&&this.nearTime<=0){const strength=1-near/38;this.synth(135-near,1.0,'sawtooth',.012+strength*.025,55);this.noise(.6,.008+strength*.018,260);this.nearTime=2.2+near*.07}}
 laser(){this.synth(1180,.19,'sawtooth',.04,170);this.synth(850,.22,'square',.012,120)}
 alienHit(killed=false){this.synth(killed?260:410,killed?.7:.38,'sawtooth',.045,killed?45:105);this.noise(killed?.55:.25,.04,killed?280:720)}
 enemyLaser(){this.synth(340,.16,'square',.018,90)}
 playerHit(){this.synth(95,.35,'sawtooth',.055,35);this.noise(.3,.055,190)}
}
