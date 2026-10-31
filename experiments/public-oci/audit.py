"""Audit native records and content-addressed OCI bytes without trusting the ledger.
Usage: python3 experiments/public-oci/audit.py
"""
import hashlib,io,json,pathlib,tarfile,datetime
root=pathlib.Path(__file__).resolve().parents[2];r=root/'results/public-oci'
def read(p):return json.loads((r/p).read_text())
def status(o):return next(c['status'] for c in o['status']['conditions'] if c['type']=='Succeeded')
def task_map(p):return {x['metadata']['labels']['tekton.dev/pipelineTask']:x for x in read(p)['items']}
def result(t,n):return next(x['value'] for x in t['status']['results'] if x['name']==n)
def bytes_at(kind,d):
 path=r/kind/(d.replace(':','-')+('.json' if kind=='manifests' else ''));b=path.read_bytes();assert hashlib.sha256(b).hexdigest()==d.removeprefix('sha256:');return b

def content(o):
 assert o['status']==200
 d='sha256:'+o['body_sha256'];raw=bytes_at('manifests',d);m=json.loads(raw);assert m==o['manifest']
 hd={k.lower():v for k,v in o['headers'].items()};assert hd['docker-content-digest']==d
 config=json.loads(bytes_at('blobs',m['config']['digest']));assert config==o['config']
 # Newest relevant layer decides the application file; reverse traversal differs
 # from the harness's forward last-write reconstruction.
 app=None;decided=False
 for layer in reversed(m['layers']):
  b=bytes_at('blobs',layer['digest']);assert len(b)==layer['size']
  with tarfile.open(fileobj=io.BytesIO(b),mode='r:*') as t:
   entries=list(t)
   if not decided:
    for e in reversed(entries):
     n=e.name.removeprefix('./').lstrip('/')
     if n=='app' and e.isfile():app=t.extractfile(e).read();decided=True;break
     if n in ['.wh.app','.wh..wh..opq']:decided=True;break
 sha=None if app is None else hashlib.sha256(app).hexdigest()
 assert sha==o['app_sha256'] and (0 if app is None else len(app))==o['app_bytes']
 return {'app_sha256':sha,'app_bytes':0 if app is None else len(app),'app_is_elf':app is not None and app[:4]==b'\x7fELF','cmd':config.get('config',{}).get('Cmd'),'arch':config.get('architecture'),'os':config.get('os'),'env':config.get('config',{}).get('Env') or [],'entrypoint':config.get('config',{}).get('Entrypoint') or [],'working_dir':config.get('config',{}).get('WorkingDir') or '', 'user':config.get('config',{}).get('User') or ''}
builder=read('builder-image-inspect.json');builder_digest='gcr.io/kaniko-project/executor@sha256:9e69fd4330ec887829c780f5126dd80edc663df6def362cd22e79bcdf00ac53f'
assert builder_digest in builder['RepoDigests'] and builder['Config']['Entrypoint']==['/kaniko/executor']
def check_builder(t):
 assert next(x for x in t['status']['steps'] if x['name']=='build-and-push')['imageID']==builder_digest
 params={x['name']:x['value'] for x in t['spec']['params']};assert params['BUILDER_IMAGE']==builder_digest
 assert '--ignore-path=/product_uuid' in params['EXTRA_ARGS']
rows=[];manual_matches=0;initials={};source_hashes={}
for profile in ['tagged','digest']:
 base=profile+'/';source=read(base+'source-run.json');tasks=task_map(base+'source-taskruns.json');assert status(source)=='False';assert set(tasks)=={'clone','build','retag','promote'}
 assert all(status(tasks[n])=='True' for n in ['clone','build','retag']) and status(tasks['promote'])=='False'
 check_builder(tasks['build'])
 assert result(tasks['clone'],'commit')=='035e3ddc3589c1c64cb69c563b178f34b1cccc5e'
 expected=content(read(base+'source-image.json'));assert expected['app_is_elf'] and expected['app_bytes']>100000 and expected['cmd']==['./app']
 source_hashes[profile]=expected
 for world in ['tag-moved','manifest-missing']:
  arms=['empty','retag','manual','rebuild','full'] if profile=='tagged' else ['empty','rebuild','manual']
  plans={}
  for arm in arms:
   p=base+world+'/'+arm+'/';run=read(p+'run.json');tr=task_map(p+'taskruns.json');initial=read(p+'initial-registry.json');delivered=content(read(p+'delivery-image.json'))
   assert initial['source']['status']==(200 if world=='tag-moved' else 404)
   if initial['source']['status']==200:assert initial['source']['body_sha256']==result(tasks['build'],'IMAGE_DIGEST').removeprefix('sha256:')
   key=(profile,world);state={'source':initial['source']['status'],'release':content(initial['release']),'delivery':content(initial['delivery'])}
   if key in initials:assert initials[key]==state
   else:initials[key]=state
   assert state['release']['app_sha256'] is None and state['delivery']['app_sha256'] is None
   valid=status(run)=='True' and delivered==expected
   native=status(run)=='True'
   if arm!='full':
    plan=read(p+'plan.json');assert plan['decision']=='recover';plans[arm]={x['name']:x['action'] for x in plan['tasks']}
    ann=run['metadata']['annotations'];body=ann['retry.tekton.dev/policy-spec'];digest='sha256:'+hashlib.sha256(body.encode()).hexdigest();assert digest==ann['retry.tekton.dev/policy-spec-sha256']==plan['policySpecSHA256'];assert json.loads(body)==plan['policySpec']
    assert ann['retry.tekton.dev/source-pipeline-run-uid']==source['metadata']['uid']
    assert all(t in plans[arm] and plans[arm][t]!='inherit' for t in tr)
   if 'clone' in tr:assert result(tr['clone'],'commit')=='035e3ddc3589c1c64cb69c563b178f34b1cccc5e'
   if 'build' in tr and status(tr['build'])=='True':
    check_builder(tr['build'])
    if 'retag' in tr and status(tr['retag'])=='True':
     params={x['name']:x['value'] for x in tr['retag']['spec']['params']};assert params['srcImageURL'].endswith('@'+result(tr['build'],'IMAGE_DIGEST'))
    if profile=='digest' and 'promote' in tr:
     params={x['name']:x['value'] for x in tr['promote']['spec']['params']};assert params['srcImageURL'].endswith('@'+result(tr['build'],'IMAGE_DIGEST'))
     dt=lambda v:datetime.datetime.fromisoformat(v.replace('Z','+00:00'));assert dt(tr['build']['status']['completionTime'])<=dt(tr['promote']['status']['startTime'])
   rows.append({'profile':profile,'world':world,'arm':arm,'native_success':native,'valid_delivery':valid,'taskruns':len(tr),'warnings':0 if arm=='full' else len(plan.get('warnings',[])),'app_sha256':delivered['app_sha256']})
  assert plans['manual']==plans['retag' if profile=='tagged' else 'rebuild'];manual_matches+=1
ledger=read('observations.json');assert len(ledger)==len(rows)==16
for x in rows:
 y=next(z for z in ledger if all(z[k]==x[k] for k in ['profile','world','arm']));assert y['succeeded']==x['native_success'] and y['valid_delivery']==x['valid_delivery'] and y['taskruns']==x['taskruns']
expected={('tagged','tag-moved','empty'):(True,False,1),('tagged','tag-moved','retag'):(True,True,2),('tagged','tag-moved','manual'):(True,True,2),('tagged','tag-moved','rebuild'):(True,True,3),('tagged','tag-moved','full'):(True,True,4),('tagged','manifest-missing','empty'):(True,False,1),('tagged','manifest-missing','retag'):(False,False,1),('tagged','manifest-missing','manual'):(False,False,1),('tagged','manifest-missing','rebuild'):(True,True,3),('tagged','manifest-missing','full'):(True,True,4),('digest','tag-moved','empty'):(True,True,1),('digest','manifest-missing','empty'):(False,False,1)}
for world in ['tag-moved','manifest-missing']:
 for arm in ['rebuild','manual']:expected[('digest',world,arm)]=(True,True,2)
for x in rows:assert (x['native_success'],x['valid_delivery'],x['taskruns'])==expected[(x['profile'],x['world'],x['arm'])],x
setup=read('setup.json')['registry_pods']['items'];end=read('environment.json')['registry_pods']['items'];assert len(setup)==len(end)==1 and setup[0]['metadata']['uid']==end[0]['metadata']['uid'];assert all(c['restartCount']==0 for c in end[0]['status']['containerStatuses'])
# Independently check effective base-image identities across the run interval.
transports=read('offline-transports.json');bindings=read('end-input-bindings.json')
for label in ['golang','alpine']:
 im='mirror.gcr.io/library/'+label+(':'+('1.18' if label=='golang' else '3'))
 t=transports['images'][im];m=read('base-'+label+'-manifest.json');cfg=(r/('base-'+label+'-config.json')).read_bytes();c=json.loads(cfg)
 assert 'sha256:'+hashlib.sha256(cfg).hexdigest()==m['manifest']['config']['digest']
 assert c['rootfs']['diff_ids']==t['source_rootfs']['Layers']
 assert c['architecture']==t['source_architecture'] and c['os']==t['source_os']
 for field in ['Env','Cmd','Entrypoint','WorkingDir','User']:
  default=[] if field in ['Env','Cmd','Entrypoint'] else ''
  assert (c.get('config',{}).get(field) or default)==(t['source_config'].get(field) or default)
 assert bindings['base_images'][label]['body_sha256']==m['body_sha256']
 assert bindings['base_images'][label]['config_sha256']==hashlib.sha256(cfg).hexdigest()
assert bindings['git_head']=='035e3ddc3589c1c64cb69c563b178f34b1cccc5e'
# The namespace inventory accounts for both excluded native calibrations too.
expected_uids=set()
for profile in ['tagged','digest']:
 expected_uids.add(read(profile+'/source-run.json')['metadata']['uid'])
 for world in ['tag-moved','manifest-missing']:
  arms=['empty','retag','manual','rebuild','full'] if profile=='tagged' else ['empty','rebuild','manual']
  for arm in arms:expected_uids.add(read(profile+'/'+world+'/'+arm+'/run.json')['metadata']['uid'])
for c in ['controller-entrypoint','multistage-filesystem']:
 run=read('calibration/'+c+'/source-run.json');assert status(run)=='False';expected_uids.add(run['metadata']['uid'])
actual_uids={o['metadata']['uid'] for o in read('all-native-pipelineruns.json')['items']}
expected_task_uids=set()
for p in r.rglob('*taskruns.json'):
 if p.name.startswith('all-native-'):continue
 for t in json.loads(p.read_text())['items']:expected_task_uids.add(t['metadata']['uid'])
assert expected_task_uids=={t['metadata']['uid'] for t in read('all-native-taskruns.json')['items']}
assert expected_uids==actual_uids and len(actual_uids)==20
assert sum(len(read(profile+'/'+world+'/'+arm+'/taskruns.json')['items']) for profile in ['tagged','digest'] for world in ['tag-moved','manifest-missing'] for arm in (['empty','retag','manual','rebuild','full'] if profile=='tagged' else ['empty','rebuild','manual']))==sum(x['taskruns'] for x in rows)
audit={'arms':len(rows),'failed_sources':2,'excluded_calibrations':2,'base_image_bindings_verified':2,'manual_policy_disposition_matches':manual_matches,'source_content':source_hashes,'rows':rows,'scope':'one execution per chosen cell; authored integration of public tasks and registry, no population/authoring superiority inference'}
(r/'audit.json').write_text(json.dumps(audit,indent=2)+'\n');print(json.dumps(audit,indent=2))
