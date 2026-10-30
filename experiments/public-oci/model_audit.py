import pathlib,json,copy,hashlib,yaml,re
root=pathlib.Path(__file__).resolve().parents[2];p=root/'experiments/public-oci';r=root/'results/public-oci';profiles={v:json.loads((p/('profile-'+v+'.json')).read_text()) for v in ['tagged','digest']}
def normalized(o):
 o=copy.deepcopy(o);o.pop('metadata',None)
 if o.get('spec') is None:o.pop('spec',None)
 for a in o.get('params',[])+o.get('results',[]):a.setdefault('type','string')
 for a in o.get('steps',[]):a.setdefault('computeResources',{})
 return o
checks={}
for t in profiles['tagged']['spec']['tasks']:
 f={'clone':'git-clone.yaml','build':'kaniko.yaml','retag':'skopeo-copy.yaml','promote':'skopeo-copy.yaml'}[t['name']];source=normalized(yaml.safe_load((p/'upstream'/f).read_text())['spec']);actual=normalized(t['taskSpec'])
 if t['name']=='build':assert actual['steps'][0].pop('command')==['/kaniko/executor']
 if t['name'] in ['retag','promote']:source['steps'][0]['image']=actual['steps'][0]['image']
 assert source==actual,t['name'];checks[t['name']]=True
x=copy.deepcopy(profiles['tagged']['spec']);y=copy.deepcopy(profiles['digest']['spec']);next(v for v in y['tasks'][-1]['params'] if v['name']=='srcImageURL')['value']=next(v for v in x['tasks'][-1]['params'] if v['name']=='srcImageURL')['value'];assert x==y
names=['clone','build','retag','promote']
def reach(profile):
 edges=set()
 for task in profile['spec']['tasks']:
  edges.update((dep,task['name']) for dep in task.get('runAfter',[]))
  edges.update((dep,task['name']) for dep in re.findall(r'\$\(tasks\.([^.)]+)\.results\.',json.dumps(task)))
 for k in names:
  for i in names:
   for j in names:
    if (i,k) in edges and (k,j) in edges:edges.add((i,j))
 return sorted(edges)
actual=reach(profiles['tagged']);assert actual==reach(profiles['digest'])==sorted((a,b) for i,a in enumerate(names) for b in names[i+1:])
d={'task_bodies_match_after_only_api_defaults_skopeo_pin_and_pinned_kaniko_entrypoint':checks,'supported_builder_param':'Kaniko 1.23.2 sha256:9e69fd4330ec887829c780f5126dd80edc663df6def362cd22e79bcdf00ac53f','only_binding_difference':'promote srcImageURL','identical_execution_reachability':actual,'profile_sha256':{v:hashlib.sha256((p/('profile-'+v+'.json')).read_bytes()).hexdigest() for v in profiles},'scope':'structure/input check after two excluded calibrations; no outcomes used'};(r/'model-audit.json').write_text(json.dumps(d,indent=2)+'\n');print(d['profile_sha256'])
