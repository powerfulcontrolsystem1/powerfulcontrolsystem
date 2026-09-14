"""Build original GTAS Blender assets. Run only with Blender's bundled Python."""
import bpy, math, os
from mathutils import Vector

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..'))
OUT = os.path.join(ROOT, 'web', 'juegos', 'simsong', 'assets', 'gtas')
os.makedirs(OUT, exist_ok=True)
bpy.ops.wm.read_factory_settings(use_empty=True)
bpy.context.preferences.filepaths.save_version=0

def mat(name, color, metallic=0.0, rough=.55, emission=None):
    m=bpy.data.materials.new(name); m.diffuse_color=(*color,1); m.use_nodes=True
    bs=m.node_tree.nodes.get('Principled BSDF'); bs.inputs['Base Color'].default_value=(*color,1); bs.inputs['Metallic'].default_value=metallic; bs.inputs['Roughness'].default_value=rough
    if emission:
        bs.inputs['Emission Color'].default_value=(*emission,1); bs.inputs['Emission Strength'].default_value=4
    return m

M={
 'olive':mat('Uniforme verde oliva',(0.14,.21,.095),0,.75),'olive2':mat('Camuflado claro',(.28,.34,.16),0,.78),
 'dark':mat('Proteccion grafito',(.055,.075,.07),.15,.4),'metal':mat('Metal cepillado',(.16,.22,.22),.78,.27),
 'skin':mat('Piel',(.63,.40,.28),0,.55),'glass':mat('Visor',(0.05,.24,.22),.5,.15, (0.02,.45,.38)),
 'laser':mat('Nucleo laser',(.12,1,.8),.1,.14,(.12,1,.8)),'alien':mat('Piel alienigena',(.22,.55,.11),0,.42),
 'alien2':mat('Pliegues alienigenas',(.1,.31,.055),0,.62),'mouth':mat('Boca alienigena',(.12,.025,.035),0,.6),
 'tooth':mat('Dientes',(.94,.86,.64),0,.48),'eye':mat('Esclerotica',(.92,.92,.72),0,.3),'pupil':mat('Pupila',(.04,.04,.025),0,.25),
 'ufo':mat('Aleacion OVNI',(.3,.4,.43),.9,.2),'ufo2':mat('Panel OVNI',(.08,.13,.15),.65,.25),'ufoGlow':mat('Luces OVNI',(.3,.9,1),.1,.1,(.2,.9,1)),
 'asphalt':mat('Asfalto',(.075,.09,.095),0,.93),'concrete':mat('Concreto',(.52,.54,.5),0,.82),'paint':mat('Pintura vial',(.9,.82,.48),0,.7),
 'steel':mat('Acero urbano',(.18,.23,.24),.72,.3),'red':mat('Rojo urbano',(.55,.045,.035),.15,.38),'wood':mat('Madera',(.23,.115,.05),0,.8),
 'leaf':mat('Hojas',(.12,.34,.12),0,.76),'flower':mat('Flores',(.72,.16,.4),0,.55),'water':mat('Agua',(.08,.4,.52),.25,.12),
}
M['cockpit']=mat('Cabina transparente',(.38,.8,.78),.1,.12)
M['cockpit'].node_tree.nodes.get('Principled BSDF').inputs['Alpha'].default_value=.22
M['cockpit'].surface_render_method='DITHERED'

def smooth(obj, bevel=0.0):
    if obj.type=='MESH':
        for p in obj.data.polygons:p.use_smooth=True
        if bevel:
            mod=obj.modifiers.new('Bordes suaves','BEVEL');mod.width=bevel;mod.segments=3
    return obj
def uv(name,loc,scale,ma,seg=24):
    bpy.ops.mesh.primitive_uv_sphere_add(segments=seg, ring_count=max(12,seg//2), location=loc);o=bpy.context.object;o.name=name;o.scale=scale;o.data.materials.append(ma);return smooth(o,.01)
def cube(name,loc,scale,ma,bev=.04):
    bpy.ops.mesh.primitive_cube_add(location=loc);o=bpy.context.object;o.name=name;o.scale=scale;o.data.materials.append(ma);return smooth(o,bev)
def cyl(name,loc,r,depth,ma,rot=(0,0,0),verts=20):
    bpy.ops.mesh.primitive_cylinder_add(vertices=verts, radius=r, depth=depth, location=loc, rotation=rot);o=bpy.context.object;o.name=name;o.data.materials.append(ma);return smooth(o,.025)
def torus(name,loc,major,minor,ma,rot=(0,0,0)):
    bpy.ops.mesh.primitive_torus_add(major_radius=major,minor_radius=minor,major_segments=32,minor_segments=10,location=loc,rotation=rot);o=bpy.context.object;o.name=name;o.data.materials.append(ma);return smooth(o)
def collection(name):
    c=bpy.data.collections.new(name);bpy.context.scene.collection.children.link(c);return c
def move_to(obj,col):
    for c in list(obj.users_collection):c.objects.unlink(obj)
    col.objects.link(obj);return obj
def objects_recursive(col):
    out=list(col.objects)
    for c in col.children:out+=objects_recursive(c)
    return out
def export(col,filename):
    bpy.ops.object.select_all(action='DESELECT')
    for o in objects_recursive(col):o.select_set(True)
    bpy.context.view_layer.objects.active=objects_recursive(col)[0]
    bpy.ops.export_scene.gltf(filepath=os.path.join(OUT,filename),export_format='GLB',use_selection=True,export_apply=True,export_yup=True,export_materials='EXPORT',export_cameras=False,export_lights=False)

# Detailed military equipment, aligned to the existing 1.8 m animated rig.
soldier=collection('GTAS_Soldier_Gear')
def S(o):return move_to(o,soldier)
def limb(name,start,end,r,ma):
    a,b=Vector(start),Vector(end);o=cyl(name,(a+b)/2,r,(b-a).length,ma,verts=20);o.rotation_euler=(b-a).to_track_quat('Z','Y').to_euler();return S(o)
S(uv('Soldier_Head',(0,-.02,1.55),(.19,.18,.23),M['skin'],32))
S(uv('Soldier_Nose',(0,-.205,1.52),(.045,.04,.065),M['skin'],20))
S(cyl('Soldier_Neck',(0,0,1.36),.105,.2,M['skin']))
S(uv('Soldier_Uniform',(0,0,1.12),(.27,.17,.33),M['olive']))
S(uv('Soldier_Hips',(0,0,.82),(.26,.16,.15),M['olive2']))
for side,suffix in [(-1,'L'),(1,'R')]:
    pivot=bpy.data.objects.new('SoldierLeg'+suffix,None);soldier.objects.link(pivot);pivot.location=(side*.14,0,.82)
    thigh=S(uv('Soldier_Thigh'+suffix,(side*.14,0,.62),(.125,.14,.25),M['olive2']))
    shin=S(uv('Soldier_Shin'+suffix,(side*.14,0,.27),(.10,.105,.23),M['olive']))
    knee=S(uv('Soldier_KneePad'+suffix,(side*.14,-.115,.46),(.105,.055,.105),M['dark']))
    boot=S(cube('Soldier_Boot'+suffix,(side*.14,-.075,.085),(.115,.21,.085),M['dark'],.05))
    for o in [thigh,shin,knee,boot]:
        loc=o.location.copy();o.parent=pivot;o.location=loc-pivot.location
    shoulder=(side*.3,0,1.36);elbow=(side*.39,-.18,1.12);hand=(.29 if side>0 else .24,-.43 if side>0 else -.71,1.14)
    S(uv('Soldier_Shoulder'+suffix,shoulder,(.13,.14,.15),M['olive2']))
    limb('Soldier_UpperArm'+suffix,shoulder,elbow,.10,M['olive'])
    limb('Soldier_Forearm'+suffix,elbow,hand,.08,M['olive2'])
    S(uv('Soldier_Glove'+suffix,hand,(.075,.085,.07),M['dark']))
S(uv('Casco balistico',(0,0,1.67),(.28,.24,.19),M['olive2'],32));S(cube('Borde casco',(0,-.17,1.61),(.29,.12,.045),M['dark'],.025));S(cube('Visor tactico',(0,-.245,1.6),(.19,.035,.07),M['glass'],.025))
S(cube('Chaleco placas',(0,-.02,1.17),(.30,.20,.31),M['olive'],.055));S(cube('Placa frontal',(0,-.225,1.18),(.245,.045,.23),M['dark'],.025));S(cube('Mochila',(0,.22,1.2),(.24,.13,.28),M['olive2'],.045))
for x in (-.2,0,.2):S(cube('Bolsa cargador',(x,-.285,1.05),(.075,.055,.11),M['olive2'],.018))
S(cube('Cinturon',(0,-.02,.87),(.32,.19,.055),M['dark'],.02));S(cyl('Cantimplora',(-.33,.02,.82),.075,.20,M['olive2'],(math.pi/2,0,0),18));S(cube('Radio',(.31,.12,1.28),(.07,.045,.13),M['dark'],.015));S(cyl('Antena',(.33,.13,1.48),.01,.22,M['metal']))
# Long laser rifle; animated hands remain visible around it.
S(cube('Rifle laser',(.32,-.42,1.16),(.09,.36,.075),M['metal'],.025));S(cyl('Canon laser',(.32,-.83,1.16),.055,.54,M['dark'],(math.pi/2,0,0),24));S(cyl('Emisor laser',(.32,-1.12,1.16),.07,.10,M['laser'],(math.pi/2,0,0),24));S(cube('Culata',(.32,-.02,1.14),(.13,.18,.11),M['olive2'],.04));S(cube('Mira holografica',(.32,-.49,1.29),(.06,.09,.06),M['glass'],.018))
export(soldier,'soldier-gear.glb')

# Alien with named tentacles so Three.js can animate each appendage.
alien=collection('GTAS_Alien')
def A(o):return move_to(o,alien)
A(uv('Alien_Torso',(0,0,1.04),(.46,.38,.61),M['alien'],32));A(uv('Alien_Cranium',(0,-.01,1.55),(.55,.42,.43),M['alien'],36));A(uv('Alien_Brow',(0,-.37,1.69),(.38,.085,.11),M['alien2'],28));A(uv('Alien_Eye',(0,-.42,1.57),(.27,.09,.23),M['eye'],32));A(uv('Alien_Pupil',(0,-.505,1.56),(.075,.025,.12),M['pupil'],24));A(uv('Alien_Mouth',(0,-.38,1.28),(.34,.075,.17),M['mouth'],28))
for i in range(7):
    x=(i-3)*.085;A(cyl('Alien_Tooth', (x,-.455,1.31),.03,.15,M['tooth'],(0,0,math.pi),10))
for i in range(6):
    a=i*math.tau/6;root=A(uv(f'Tentacle{i+1:02d}',(math.sin(a)*.3,math.cos(a)*.18,.78),(.16,.16,.28),M['alien'],22));root.rotation_euler=(math.sin(a)*.5,math.cos(a)*.5,a)
    A(uv(f'TentacleTip{i+1:02d}',(math.sin(a)*.68,math.cos(a)*.5,.29),(.14,.14,.38),M['alien2'],22)).rotation_euler=(math.sin(a)*.65,math.cos(a)*.65,a)
for side in (-1,1):A(uv('Alien_Collar', (side*.33,0,1.02),(.17,.28,.14),M['alien2'],22))
export(alien,'alien.glb')

ufo=collection('GTAS_UFO')
def U(o):return move_to(o,ufo)
U(uv('UFO_Hull',(0,0,.05),(3.1,3.1,.48),M['ufo'],40));U(torus('UFO_Ring',(0,0,.08),2.65,.24,M['ufo2']));U(uv('UFO_Cockpit',(0,0,.7),(1.48,1.48,.95),M['cockpit'],40));U(uv('UFO_Pilot',(0,0,.68),(.48,.44,.62),M['alien'],28));U(uv('UFO_Eye',(0,-.43,.85),(.23,.08,.18),M['eye'],24));U(uv('UFO_Pupil',(0,-.50,.85),(.07,.025,.09),M['pupil'],20))
for i in range(16):
    a=i*math.tau/16;U(uv(f'UFO_Light_{i:02d}',(math.sin(a)*2.7,math.cos(a)*2.7,-.22),(.13,.13,.09),M['ufoGlow'],18))
for i in range(5):
    a=i*math.tau/5;U(uv(f'UFO_Tentacle_{i:02d}',(math.sin(a)*.55,math.cos(a)*.55,.48),(.12,.12,.35),M['alien2'],18))
export(ufo,'ufo.glb')

# A single optimized Blender environment layer: roads, curbs and street furniture.
world=collection('GTAS_Springfield_World_Detail')
def W(o):return move_to(o,world)
for line in (-112,-40,32,104):
    W(cube('Asfalto vertical',(line,0,.025),(5.6,186,.025),M['asphalt'],.02));W(cube('Asfalto horizontal',(0,-line,.024),(186,5.6,.024),M['asphalt'],.02))
    for side in (-1,1):
        W(cube('Anden vertical',(line+side*7,0,.04),(1.25,186,.04),M['concrete'],.03));W(cube('Anden horizontal',(0,-line-side*7,.04),(186,1.25,.04),M['concrete'],.03))
    for i in range(-174,175,14):
        W(cube('Linea vial',(line,i,.075),(.08,3.2,.018),M['paint'],.01));W(cube('Linea vial',(i,-line,.074),(3.2,.08,.018),M['paint'],.01))
for x in (-119,-47,25,97,111):
  for y in (-119,-47,25,97,111):
    if math.hypot(x,y)>175:continue
    W(cyl('Farola poste',(x,-y,2.8),.065,5.6,M['steel'],verts=18));W(cube('Farola brazo',(x-.35,-y,5.5),(.4,.055,.055),M['steel'],.02));W(uv('Farola luz',(x-.72,-y,5.42),(.13,.13,.09),M['laser'],18))
for x,y in [(-88,106),(50,26),(117,28),(-28,-54),(82,-48),(-126,66),(18,118)]:
    W(cyl('Hidrante',(x,-y,.42),.16,.66,M['red'],verts=20));W(torus('Hidrante aro',(x,-y,.65),.19,.045,M['metal'],(math.pi/2,0,0)))
for x,y,rot in [(-83,62,0),(50,54,math.pi/2),(11,-20,0),(121,52,math.pi/2)]:
    W(cube('Banco asiento',(x,-y,.58),(1.15,.32,.10),M['wood'],.06));W(cube('Banco respaldo',(x,-y+.26,1.0),(1.15,.08,.38),M['wood'],.06));W(cube('Banco pata',(x-.8,-y,.28),(.08,.22,.3),M['steel'],.02));W(cube('Banco pata',(x+.8,-y,.28),(.08,.22,.3),M['steel'],.02))
for i in range(54):
    a=i*2.399963;radius=55+(i*47)%112;x=math.cos(a)*radius;y=math.sin(a)*radius
    if any(abs(x-r)<11 or abs(y-r)<11 for r in (-112,-40,32,104)):continue
    W(cyl('Arbol tronco',(x,-y,1.5),.28,3,M['wood'],verts=14));W(uv('Arbol copa',(x,-y,4.3),(2.1,2.1,2.8),M['leaf'],24))
for x,y in [(-55,55),(82,89),(10,70),(118,-70),(-140,-46)]:
    W(cube('Jardin',(x,-y,.20),(3.1,1.2,.2),M['concrete'],.12))
    for j in range(9):W(uv('Flor',(x-2.4+j*.6,-y,.55),(.18,.18,.25),M['flower'],14))
W(uv('Estanque Springfield',(145,120,.04),(19,12,.12),M['water'],40))
export(world,'springfield-world.glb')

bpy.ops.wm.save_as_mainfile(filepath=os.path.join(os.path.dirname(__file__),'gtas_models.blend'),compress=True)
print('GTAS_ASSETS_BUILT', OUT)
