import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root=path.resolve(import.meta.dirname,'../web/juegos'),manifest=path.join(root,'simsong/assets/manifest.json');
function files(dir){return fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?files(path.join(dir,e.name)):[path.join(dir,e.name)])}
const entries=[...files(path.join(root,'simsong/assets')),...files(path.join(root,'vendor/three'))].filter(p=>p!==manifest).sort().map(p=>{const b=fs.readFileSync(p);return {path:path.relative(root,p).replaceAll('\\','/'),bytes:b.length,sha256:crypto.createHash('sha256').update(b).digest('hex')}});
if(process.argv.includes('--write'))fs.writeFileSync(manifest,JSON.stringify({three:'0.186.0',licenses:['MIT (Three.js)','CC0 (Kenney)','Original PCS (GTAS Blender assets)'],files:entries},null,2)+'\n');
else assert.deepEqual(JSON.parse(fs.readFileSync(manifest)).files,entries,'Asset hashes differ; review provenance before regenerating');
for(const e of entries.filter(e=>e.path.includes('/characters/')&&e.path.endsWith('.glb'))){
 const b=fs.readFileSync(path.join(root,e.path)),j=JSON.parse(b.toString('utf8',20,20+b.readUInt32LE(12)));
 for(const name of ['idle','run','jump']){const a=j.animations.find(a=>a.name===name);assert.ok(a,`${e.path}: ${name}`);assert.ok(a.samplers.some(s=>j.accessors[s.input].count>2),`${e.path}: ${name} must contain movement, not the targeting pose`)}
 assert.ok(j.images.every(i=>i.uri&&!i.uri.startsWith('http')),`${e.path}: local external textures required by CSP`);
}
console.log(`Springfield: ${entries.length} local assets verified (${(entries.reduce((n,e)=>n+e.bytes,0)/1048576).toFixed(2)} MiB)`);
