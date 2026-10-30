"""Run paired external worlds from identical source Tekton snapshots.
Usage: python3 experiments/acceptance/live.py --retryctl /path --context kind-retry-paper
Creates its own namespace; retains it for audit. No existing objects are changed.
"""
import argparse,copy,json,pathlib,subprocess,time,urllib.request,urllib.error,socket,random
ap=argparse.ArgumentParser();ap.add_argument('--retryctl',required=True);ap.add_argument('--context',default='kind-retry-paper');ap.add_argument('--namespace',default='retry-acceptance-20261030');a=ap.parse_args()
root=pathlib.Path(__file__).resolve().parents[2];out=root/'results/acceptance/live';out.mkdir(parents=True,exist_ok=True)
def save(name,data):
 p=out/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_text(json.dumps(data,indent=2,sort_keys=True)+'\n');return str(p)
def k(*args,body=None):
 p=subprocess.run(['kubectl','--context',a.context,'-n',a.namespace,*args],input=None if body is None else json.dumps(body),text=True,capture_output=True,timeout=180)
 if p.returncode:raise RuntimeError(p.stderr)
 return p.stdout
def apply(obj):return k('apply','-f','-',body=obj)
def capture(run):
 r=json.loads(k('get','pipelinerun',run,'-o','json'));t=json.loads(k('get','taskruns','-l','tekton.dev/pipelineRun='+run,'-o','json'));return r,t
def wait(run):
 end=time.monotonic()+180
 while time.monotonic()<end:
  r,t=capture(run);c=next((c for c in r.get('status',{}).get('conditions',[]) if c['type']=='Succeeded'),{})
  if c.get('status') in ['True','False'] and all(next((c.get('status') for c in tr.get('status',{}).get('conditions',[]) if c['type']=='Succeeded'),'Unknown') in ['True','False'] for tr in t['items']):return r,t,c['status']=='True'
  time.sleep(1)
 raise RuntimeError('timeout '+run)
def api(path,data=None):
 req=urllib.request.Request('http://127.0.0.1:29173'+path,data=None if data is None else json.dumps(data).encode(),headers={'Content-Type':'application/json'})
 return json.load(urllib.request.urlopen(req,timeout=10))
# Explicit isolated namespace creation; refuse accidental reuse.
p=subprocess.run(['kubectl','--context',a.context,'create','namespace',a.namespace],capture_output=True,text=True)
if p.returncode:raise RuntimeError('new namespace required: '+p.stderr)
server=(root/'experiments/acceptance/oracle.py').read_text()
apply({'apiVersion':'v1','kind':'ConfigMap','metadata':{'name':'oracle-code'},'data':{'oracle.py':server}})
apply({'apiVersion':'apps/v1','kind':'Deployment','metadata':{'name':'oracle'},'spec':{'replicas':1,'selector':{'matchLabels':{'app':'retry-acceptance-oracle'}},'template':{'metadata':{'labels':{'app':'retry-acceptance-oracle'}},'spec':{'containers':[{'name':'oracle','image':'python:3.12-alpine','command':['python','/code/oracle.py'],'ports':[{'containerPort':8000}],'volumeMounts':[{'name':'code','mountPath':'/code'}]}],'volumes':[{'name':'code','configMap':{'name':'oracle-code'}}]}}}})
apply({'apiVersion':'v1','kind':'Service','metadata':{'name':'oracle'},'spec':{'selector':{'app':'retry-acceptance-oracle'},'ports':[{'port':8000,'targetPort':8000}]}})
k('rollout','status','deployment/oracle','--timeout=120s')
forward=subprocess.Popen(['kubectl','--context',a.context,'-n',a.namespace,'port-forward','service/oracle','29173:8000'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
try:
 for _ in range(100):
  try:
   with socket.create_connection(('127.0.0.1',29173),timeout=.2):break
  except OSError:time.sleep(.2)
 names=['warmup','acquire','publish','sign','scan','deploy'];tasks=[]
 for i,name in enumerate(names):
  script='#!/bin/sh\nset -eu\necho warmed\n' if name=='warmup' else '#!/bin/sh\nset -eu\nwget -q -O- --post-data=\'{}\' http://oracle:8000/execute/'+name+'\n'
  task={'name':name,'taskSpec':{'steps':[{'name':'execute','image':'alpine:3.20','script':script}]}}
  if i:task['runAfter']=[names[i-1]]
  tasks.append(task)
 pipeline={'apiVersion':'tekton.dev/v1','kind':'Pipeline','metadata':{'name':'external-recovery'},'spec':{'tasks':tasks}};apply(pipeline)
 live_pipeline=json.loads(k('get','pipeline','external-recovery','-o','json'));save('pipeline.json',live_pipeline)
 policy={'metadata':{'name':'external-contract','generation':1},'spec':{'pipelineRef':'external-recovery','tasks':{name:{'retryWith':[names[names.index(name)-1]]} for name in ['publish','sign','scan','deploy']}}}
 save('policy.json',policy);save('empty-policy.json',{'metadata':{'name':'empty'},'spec':{'pipelineRef':'external-recovery','tasks':{}}})
 observations=[];rng=random.Random(20261030)
 for fail in ['publish','sign','scan']:
  api('/reset',{'fail_at':fail});source_name='source-'+fail
  apply({'apiVersion':'tekton.dev/v1','kind':'PipelineRun','metadata':{'name':source_name},'spec':{'pipelineRef':{'name':'external-recovery'}}})
  source,trs,ok=wait(source_name);assert not ok
  prefix=fail+'/';save(prefix+'source-run.json',source);save(prefix+'source-taskruns.json',trs)
  snapshot=api('/state');assert snapshot['injected'];save(prefix+'source-external-state.json',snapshot)
  worlds={'stable':snapshot,'rotated':dict(copy.deepcopy(snapshot),epoch=2)}
  for world,snap in worlds.items():
   arms=['empty','declared','manual','full'];rng.shuffle(arms)
   save(prefix+world+'/initial-state.json',snap)
   for arm in arms:
    api('/restore',copy.deepcopy(snap));base=prefix+world+'/'+arm+'/'
    if arm=='full':manifest={'apiVersion':'tekton.dev/v1','kind':'PipelineRun','metadata':{'name':'full-'+fail+'-'+world},'spec':{'pipelineRef':{'name':'external-recovery'}}}
    else:
     policy_path=out/('policy.json' if arm=='declared' else 'empty-policy.json')
     args=['--pipeline',str(out/'pipeline.json'),'--run',str(out/(prefix+'source-run.json')),'--taskruns',str(out/(prefix+'source-taskruns.json')),'--policy',str(policy_path)]
     if arm=='manual':args+=['--tasks',','.join(names[1:names.index(fail)])]
     p=subprocess.run([a.retryctl,'plan',*args],capture_output=True,text=True);assert p.returncode==0,p.stderr
     plan=json.loads(p.stdout);save(base+'plan.json',plan);assert plan['decision']=='recover';assert not plan.get('warnings'),plan
     p=subprocess.run([a.retryctl,'create',*args],capture_output=True,text=True);assert p.returncode==0,p.stderr
     # Parse YAML using kubectl's client serializer to preserve actual generated manifest.
     (out/(base+'generated.yaml')).write_text(p.stdout)
     p=subprocess.run(['kubectl','--context',a.context,'-n',a.namespace,'create','-f',str(out/(base+'generated.yaml')),'-o','json'],capture_output=True,text=True);assert p.returncode==0,p.stderr
     created=json.loads(p.stdout);name=created['metadata']['name']
    if arm=='full':apply(manifest);name=manifest['metadata']['name']
    run,taskruns,success=wait(name);state=api('/state')
    save(base+'run.json',run);save(base+'taskruns.json',taskruns);save(base+'external-state.json',state)
    semantic=state['deployed']==state['epoch'] and all(state[x]==state['epoch'] for x in ['credential','artifact','signature','scanned'])
    row={'failure':fail,'world':world,'arm':arm,'succeeded':success,'semantic_valid':semantic,'taskruns':len(taskruns['items']),'warnings':0 if arm!='full' else None,'arm_order':arms}
    observations.append(row);save('observations.json',observations);print(json.dumps(row),flush=True)
    expected=(world=='stable' or arm!='empty');assert success==expected and semantic==expected,row
 # Native active-finally boundary trace.
 probe={'apiVersion':'tekton.dev/v1','kind':'Pipeline','metadata':{'name':'active-source-probe'},'spec':{'tasks':[{'name':'fail','taskSpec':{'steps':[{'name':'fail','image':'alpine:3.20','script':'exit 1'}]}}],'finally':[{'name':'cleanup','taskSpec':{'steps':[{'name':'wait','image':'alpine:3.20','script':'sleep 25'}]}}]}}
 apply(probe);apply({'apiVersion':'tekton.dev/v1','kind':'PipelineRun','metadata':{'name':'active-source'},'spec':{'pipelineRef':{'name':'active-source-probe'}}})
 save('active/pipeline.json',json.loads(k('get','pipeline','active-source-probe','-o','json')));save('active/policy.json',{'metadata':{'name':'empty'},'spec':{'pipelineRef':'active-source-probe','tasks':{}}})
 end=time.monotonic()+90
 while time.monotonic()<end:
  r,t=capture('active-source')
  failed=any(any(c.get('status')=='False' for c in tr.get('status',{}).get('conditions',[]) if c['type']=='Succeeded') for tr in t['items'])
  unknown=any(c.get('status')=='Unknown' for c in r.get('status',{}).get('conditions',[]) if c['type']=='Succeeded')
  if failed and unknown:break
  time.sleep(.5)
 else:raise AssertionError('active-source interval not captured')
 for phase in ['active','terminal']:
  save('active/'+phase+'-run.json',r);save('active/'+phase+'-taskruns.json',t)
  args=['--pipeline',str(out/'active/pipeline.json'),'--run',str(out/('active/'+phase+'-run.json')),'--taskruns',str(out/('active/'+phase+'-taskruns.json')),'--policy',str(out/'active/policy.json')]
  p=subprocess.run([a.retryctl,'plan',*args],capture_output=True,text=True);plan=json.loads(p.stdout);save('active/'+phase+'-plan.json',plan);assert plan['decision']==('refuse' if phase=='active' else 'recover')
  if phase=='active':
   p=subprocess.run([a.retryctl,'create',*args],capture_output=True,text=True);assert p.returncode!=0 and not p.stdout;save('active/create-refusal.json',{'exit_code':p.returncode,'stderr':p.stderr,'stdout':p.stdout})
   r,t,ok=wait('active-source');assert not ok
 save('environment.json',{'namespace':a.namespace,'context':a.context,'seed':20261030,'controller':json.loads(k('get','deployment','tekton-pipelines-controller','-n','tekton-pipelines','-o','json'))['spec']['template']['spec']['containers'][0]['image'],'scope':'24 authored native recovery/restart arms; one repetition per failure/world/arm; external service oracle is not a production registry or independent policy-author study'})
 print('LIVE EXPERIMENT COMPLETE',flush=True)
finally:
 forward.terminate();forward.wait(timeout=10)
