"""Resolved public-task OCI recovery case. --setup prepares profiles; --run executes.
Requires PyYAML, kubectl, and an isolated kind context. Retains its namespace.
"""
import argparse,copy,hashlib,io,json,pathlib,random,socket,subprocess,tarfile,time,urllib.request,urllib.error
import yaml
ap=argparse.ArgumentParser();ap.add_argument('--setup',action='store_true');ap.add_argument('--run',action='store_true');ap.add_argument('--mirror-only',action='store_true');ap.add_argument('--retryctl');ap.add_argument('--context',default='kind-retry-paper');ap.add_argument('--namespace',default='retry-public-oci-20261030');a=ap.parse_args()
root=pathlib.Path(__file__).resolve().parents[2];inputs=root/'experiments/public-oci';out=root/'results/public-oci';out.mkdir(parents=True,exist_ok=True)
def save(p,d):
 p=out/p;p.parent.mkdir(parents=True,exist_ok=True);p.write_text(json.dumps(d,indent=2,sort_keys=True)+'\n');return p
def k(*args,body=None,timeout=180):
 p=subprocess.run(['kubectl','--context',a.context,'-n',a.namespace,*args],input=None if body is None else json.dumps(body),text=True,capture_output=True,timeout=timeout)
 if p.returncode:raise RuntimeError(p.stderr)
 return p.stdout
def apply(o):return k('apply','-f','-',body=o)
def delivery(active):
 k('patch','service','registry-delivery','--type=merge','-p',json.dumps({'spec':{'selector':{'app':'paper-oci-registry' if active else 'delivery-disabled'}}}))
 end=time.monotonic()+60
 while time.monotonic()<end:
  e=json.loads(k('get','endpoints','registry-delivery','-o','json'))
  ready=any(x.get('addresses') for x in e.get('subsets',[]))
  if ready==active:return
  time.sleep(.5)
 raise RuntimeError('delivery endpoint transition timed out')
def capture(n):return json.loads(k('get','pipelinerun',n,'-o','json')),json.loads(k('get','taskruns','-l','tekton.dev/pipelineRun='+n,'-o','json'))
def condition(o):return next((c['status'] for c in o.get('status',{}).get('conditions',[]) if c['type']=='Succeeded'),'Unknown')
def wait(n):
 end=time.monotonic()+1200
 while time.monotonic()<end:
  r,t=capture(n)
  if condition(r) in ['True','False'] and all(condition(x) in ['True','False'] for x in t['items']):return r,t
  time.sleep(2)
 raise RuntimeError('timeout '+n)
def skopeo(src,dst,public=False):
 return k('exec','deployment/oci-admin','--','skopeo','copy','--src-tls-verify='+str(public).lower(),'--dest-tls-verify=false',src,dst,timeout=600)
def http(path,method='GET',data=None,ctype=None):
 h={'Accept':'application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.oci.image.index.v1+json'}
 if ctype:h['Content-Type']=ctype
 try:
  with urllib.request.urlopen(urllib.request.Request('http://127.0.0.1:29174/v2/'+path,data=data,headers=h,method=method),timeout=60) as r:return r.status,dict(r.headers),r.read()
 except urllib.error.HTTPError as e:return e.code,dict(e.headers),e.read()
def manifest(repo,ref):
 status,h,b=http(repo+'/manifests/'+ref)
 if status!=200:return {'status':status,'headers':h,'body':b.decode()},None
 digest=hashlib.sha256(b).hexdigest();path=out/'manifests'/('sha256-'+digest+'.json');path.parent.mkdir(exist_ok=True);path.write_bytes(b)
 return {'status':status,'headers':h,'manifest':json.loads(b),'body_sha256':digest},b

def inspect_image(repo,ref):
 m,b=manifest(repo,ref)
 if b is None:return m
 obj=m['manifest'];assert 'layers' in obj,obj
 def blob(d):
  path=out/'blobs'/d.replace(':','-');path.parent.mkdir(exist_ok=True)
  if path.exists():bb=path.read_bytes()
  else:
   st,_,bb=http(repo+'/blobs/'+d);assert st==200,st;path.write_bytes(bb)
  assert 'sha256:'+hashlib.sha256(bb).hexdigest()==d
  return bb
 config=json.loads(blob(obj['config']['digest']));app=None;layer_records=[]
 for layer in obj['layers']:
  bb=blob(layer['digest']);names=[]
  with tarfile.open(fileobj=io.BytesIO(bb),mode='r:*') as tar:
   for entry in tar:
    name=entry.name.removeprefix('./').lstrip('/');names.append(name)
    if name in ['.wh.app','.wh..wh..opq']:app=None
    if name=='app' and entry.isfile():app=tar.extractfile(entry).read()
  layer_records.append({'digest':layer['digest'],'entries':names})
 m.update({'config':config,'layers':layer_records,'app_sha256':None if app is None else hashlib.sha256(app).hexdigest(),'app_bytes':0 if app is None else len(app),'app_is_elf':app is not None and app[:4]==b'\x7fELF'})
 return m

def profile(variant,images):
 def spec(f):return copy.deepcopy(yaml.safe_load((inputs/'upstream'/f).read_text())['spec'])
 clone=spec('git-clone.yaml');build=spec('kaniko.yaml');copy_spec=spec('skopeo-copy.yaml')
 copy_spec['steps'][0]['image']=images['quay.io/skopeo/stable:v1']['repo_digests'][0]
 params=lambda d:[{'name':k,'value':v} for k,v in d.items()]
 src='registry-source:5000/application';release='registry-source:5000/release:latest'
 tasks=[{'name':'clone','taskSpec':clone,'params':params({'url':'https://github.com/GoogleContainerTools/skaffold','revision':'035e3ddc3589c1c64cb69c563b178f34b1cccc5e','submodules':'false','depth':'1','sparseCheckoutDirectories':'examples/getting-started','gitInitImage':images['ghcr.io/tektoncd-catalog/git-clone:v1.1.0']['repo_digests'][0]}),'workspaces':[{'name':'output','workspace':'source'}]},
 {'name':'build','runAfter':['clone'],'taskSpec':build,'params':params({'IMAGE':src+':build','DOCKERFILE':'examples/getting-started/Dockerfile','CONTEXT':'examples/getting-started','EXTRA_ARGS':['--insecure','--insecure-pull','--cache=false','--reproducible','--force','--build-arg=BASE_PREFIX=registry-source:5000/base/']}),'workspaces':[{'name':'source','workspace':'source'}]}]
 for name,ref,dest in [('retag',src+'@$(tasks.build.results.IMAGE_DIGEST)',release),('promote',release if variant=='tagged' else src+'@$(tasks.build.results.IMAGE_DIGEST)','registry-delivery:5000/delivery:latest')]:
  tasks.append({'name':name,'runAfter':['build' if name=='retag' else 'retag'],'taskSpec':copy.deepcopy(copy_spec),'params':params({'srcImageURL':'docker://'+ref,'destImageURL':'docker://'+dest,'srcTLSverify':'false','destTLSverify':'false'}),'workspaces':[{'name':'images-url','workspace':'copy-input'}]})
 return {'apiVersion':'tekton.dev/v1','kind':'Pipeline','metadata':{'name':'public-oci-'+variant},'spec':{'workspaces':[{'name':'source'},{'name':'copy-input'}],'tasks':tasks}}

if a.setup:
 p=subprocess.run(['kubectl','--context',a.context,'create','namespace',a.namespace],text=True,capture_output=True)
 if p.returncode:raise RuntimeError('new namespace required: '+p.stderr)
 k('label','namespace',a.namespace,'retry-paper-owner=public-oci-case')
 images=json.loads((inputs/'upstream/runtime-images.json').read_text())
 apply({'apiVersion':'apps/v1','kind':'Deployment','metadata':{'name':'registry'},'spec':{'replicas':1,'selector':{'matchLabels':{'app':'paper-oci-registry'}},'template':{'metadata':{'labels':{'app':'paper-oci-registry'}},'spec':{'containers':[{'name':'registry','image':images['registry:2.8.3']['repo_digests'][0],'env':[{'name':'REGISTRY_STORAGE_DELETE_ENABLED','value':'true'}],'ports':[{'containerPort':5000}],'volumeMounts':[{'name':'data','mountPath':'/var/lib/registry'}]}],'volumes':[{'name':'data','emptyDir':{}}]}}}})
 for service in ['registry-source','registry-delivery']:
  apply({'apiVersion':'v1','kind':'Service','metadata':{'name':service},'spec':{'selector':{'app':'paper-oci-registry' if service=='registry-source' else 'delivery-disabled'},'ports':[{'port':5000,'targetPort':5000}]}})
 apply({'apiVersion':'apps/v1','kind':'Deployment','metadata':{'name':'oci-admin'},'spec':{'replicas':1,'selector':{'matchLabels':{'app':'paper-oci-admin'}},'template':{'metadata':{'labels':{'app':'paper-oci-admin'}},'spec':{'containers':[{'name':'admin','image':images['quay.io/skopeo/stable:v1']['repo_digests'][0],'command':['/bin/sh','-c','sleep 36000'],'securityContext':{'runAsUser':65532,'runAsNonRoot':True},'env':[{'name':'HOME','value':'/tmp'}]}]}}}})
 apply({'apiVersion':'v1','kind':'PersistentVolumeClaim','metadata':{'name':'source'},'spec':{'accessModes':['ReadWriteOnce'],'resources':{'requests':{'storage':'2Gi'}}}})
 k('rollout','status','deployment/registry','--timeout=300s',timeout=330);k('rollout','status','deployment/oci-admin','--timeout=300s',timeout=330)
 for v in ['tagged','digest']:
  p=profile(v,images);apply(p);actual=json.loads(k('get','pipeline',p['metadata']['name'],'-o','json'));save(v+'/pipeline.json',actual);(inputs/('profile-'+v+'.json')).write_text(json.dumps(actual,indent=2)+'\n')
 save('setup.json',{'namespace':a.namespace,'context':a.context,'registry_pods':json.loads(k('get','pods','-l','app=paper-oci-registry','-o','json')),'created_at':time.time()})
 print('SETUP COMPLETE: no failed sources or recovery arms executed',flush=True)

if a.run or a.mirror_only:
 assert a.retryctl or a.mirror_only
 ns=json.loads(k('get','namespace',a.namespace,'-o','json'));assert ns['metadata']['labels']['retry-paper-owner']=='public-oci-case'
 forward=subprocess.Popen(['kubectl','--context',a.context,'-n',a.namespace,'port-forward','service/registry-source','29174:5000'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
 try:
  for _ in range(100):
   try:
    with socket.create_connection(('127.0.0.1',29174),timeout=.2):break
   except OSError:time.sleep(.2)
  images=json.loads((inputs/'upstream/runtime-images.json').read_text())
  for image,target in [('mirror.gcr.io/library/golang:1.18','golang:1.18'),('mirror.gcr.io/library/alpine:3','alpine:3')]:
   skopeo('docker-archive:/tmp/'+target.split(':')[0]+'.tar','docker://registry-source:5000/base/'+target)
   base_meta,_=manifest('base/'+target.split(':')[0],target.split(':')[1]);save('base-'+target.split(':')[0]+'-manifest.json',base_meta)
   config_digest=base_meta['manifest']['config']['digest'];st,_,config_bytes=http('base/'+target.split(':')[0]+'/blobs/'+config_digest);assert st==200 and 'sha256:'+hashlib.sha256(config_bytes).hexdigest()==config_digest
   (out/('base-'+target.split(':')[0]+'-config.json')).write_bytes(config_bytes)
   config=json.loads(config_bytes);transport=json.loads((out/'offline-transports.json').read_text())['images'][image]
   assert config['rootfs']['diff_ids']==transport['source_rootfs']['Layers'];assert config['architecture']==transport['source_architecture'] and config['os']==transport['source_os']
   for field in ['Env','Cmd','Entrypoint','WorkingDir','User']:
    default=[] if field in ['Env','Cmd','Entrypoint'] else '';assert (config.get('config',{}).get(field) or default)==(transport['source_config'].get(field) or default),(image,field)

  skopeo('docker-archive:/tmp/foreign.tar','docker://registry-source:5000/foreign:stable')
  save('foreign-image.json',inspect_image('foreign','stable'))
  if a.mirror_only:
   print('REGISTRY MIRRORS VERIFIED: source rootfs/config match pinned host images',flush=True);raise SystemExit(0)
  observations=[];rng=random.Random(20261030)
  for variant in ['tagged','digest']:
   prefix=variant+'/'
   delivery(False)
   skopeo('docker://registry-source:5000/foreign:stable','docker://registry-source:5000/delivery:latest')
   name='source-final-'+variant
   spec={'pipelineRef':{'name':'public-oci-'+variant},'workspaces':[{'name':'source','persistentVolumeClaim':{'claimName':'source'}},{'name':'copy-input','emptyDir':{}}],'timeouts':{'pipeline':'20m'},'taskRunTemplate':{'podTemplate':{'securityContext':{'fsGroup':65532}}}}
   apply({'apiVersion':'tekton.dev/v1','kind':'PipelineRun','metadata':{'name':name},'spec':spec})
   source,trs=wait(name);save(prefix+'source-run.json',source);save(prefix+'source-taskruns.json',trs);assert condition(source)=='False'
   by={tr['metadata']['labels']['tekton.dev/pipelineTask']:tr for tr in trs['items']};assert set(by)=={'clone','build','retag','promote'};assert all(condition(by[n])=='True' for n in ['clone','build','retag']) and condition(by['promote'])=='False'
   for tr in trs['items']:
    pod=tr['status']['podName'];(out/(prefix+tr['metadata']['labels']['tekton.dev/pipelineTask']+'-logs.txt')).write_text(k('logs',pod,'--all-containers=true'))
   results={x['name']:x['value'] for x in by['build']['status']['results']};digest=results['IMAGE_DIGEST'];expected=inspect_image('application',digest);save(prefix+'source-image.json',expected);assert expected['app_is_elf'] and expected['app_bytes']>100000
   original,raw=manifest('application',digest);ctype=original['headers'].get('Content-Type','application/vnd.docker.distribution.manifest.v2+json')
   delivery(True)
   policies={'empty':{},'retag':{'promote':{'retryWith':['retag']}},'rebuild':{'promote':{'retryWith':['build','retag'] if variant=='tagged' else ['build']}}}
   for label,rules in policies.items():save(prefix+'policy-'+label+'.json',{'metadata':{'name':'public-oci-'+label,'generation':1},'spec':{'pipelineRef':'public-oci-'+variant,'tasks':rules}})
   for world in ['tag-moved','manifest-missing']:
    arms=['empty','retag','manual','rebuild','full'] if variant=='tagged' else ['empty','rebuild','manual'];rng.shuffle(arms)
    for arm in arms:
     st,_,_=http('application/manifests/build','PUT',raw,ctype);assert st==201,st
     skopeo('docker://registry-source:5000/foreign:stable','docker://registry-source:5000/release:latest');skopeo('docker://registry-source:5000/foreign:stable','docker://registry-source:5000/delivery:latest')
     if world=='manifest-missing':st,_,_=http('application/manifests/'+digest,'DELETE');assert st==202,st
     base=prefix+world+'/'+arm+'/';initial={'source':manifest('application',digest)[0],'release':inspect_image('release','latest'),'delivery':inspect_image('delivery','latest')};save(base+'initial-registry.json',initial);assert initial['source']['status']==(404 if world=='manifest-missing' else 200)
     if arm=='full':
      run_name='full-'+variant+'-'+world;apply({'apiVersion':'tekton.dev/v1','kind':'PipelineRun','metadata':{'name':run_name},'spec':spec})
     else:
      policy_label=arm if arm in ['retag','rebuild'] else 'empty'
      args=['--pipeline',str(out/(prefix+'pipeline.json')),'--run',str(out/(prefix+'source-run.json')),'--taskruns',str(out/(prefix+'source-taskruns.json')),'--policy',str(out/(prefix+'policy-'+policy_label+'.json'))]
      if arm=='manual':args+=['--tasks','retag' if variant=='tagged' else 'build']
      p=subprocess.run([a.retryctl,'plan',*args],capture_output=True,text=True);assert p.returncode==0,p.stderr;plan=json.loads(p.stdout);save(base+'plan.json',plan);assert plan['decision']=='recover',plan
      p=subprocess.run([a.retryctl,'create',*args,'--confirm-warnings'],capture_output=True,text=True);assert p.returncode==0,p.stderr;path=out/(base+'generated.yaml');path.write_text(p.stdout)
      created=json.loads(k('create','-f',str(path),'-o','json'));run_name=created['metadata']['name']
     run,taskruns=wait(run_name);save(base+'run.json',run);save(base+'taskruns.json',taskruns)
     state=inspect_image('delivery','latest');save(base+'delivery-image.json',state)
     valid=condition(run)=='True' and state.get('app_sha256')==expected['app_sha256'] and state.get('app_is_elf')
     row={'profile':variant,'world':world,'arm':arm,'succeeded':condition(run)=='True','valid_delivery':bool(valid),'taskruns':len(taskruns['items']),'arm_order':arms,'app_sha256':state.get('app_sha256')};observations.append(row);save('observations.json',observations);print(json.dumps(row),flush=True)
  save('environment.json',{'namespace':a.namespace,'context':a.context,'seed':20261030,'registry_pods':json.loads(k('get','pods','-l','app=paper-oci-registry','-o','json')),'scope':'16 native execution arms plus 2 failed sources; public-task authored integration, not production incidence or practitioner evidence'})
  print('PUBLIC OCI EXPERIMENT COMPLETE',flush=True)
 finally:forward.terminate();forward.wait(timeout=10)
