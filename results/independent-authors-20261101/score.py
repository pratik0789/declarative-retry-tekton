import json, subprocess, sys, pathlib, yaml
B=sys.argv[1]; R=pathlib.Path('results'); A=pathlib.Path('results/independent-authors-20261101/authors')
def plan(pipe,run,trs,pol,sel=None):
    a=[B,'plan','--pipeline',str(pipe),'--run',str(run),'--taskruns',str(trs),'--policy',str(pol)]
    if sel: a+=['--tasks',','.join(sel)]
    return json.loads(subprocess.run(a,capture_output=True,text=True).stdout)
disp=lambda p:(p['decision'],tuple(sorted((t['name'],t['action']) for t in p.get('tasks') or [])))
out={}
# IV-G (final-version closure captures, 5 each)
C=R/'final-version-20261030/closure'
for mech,pfx,sel,ver in [('partial-write','pw',None,'closure-verification'),('stale-build','fi',['prepare'],None)]:
    for a in (1,2,3):
        pol=A/f'{mech}-a{a}/policy.yaml'; res=[]
        for i in range(1,6):
            d=C/f'{pfx}-{i}'
            # rewrite pipelineRef per case
            y=yaml.safe_load(open(pol)); y['spec']['pipelineRef']=f'{pfx}-{i}'; tmp=pathlib.Path(f'/tmp/ia-{mech}-{a}-{i}.yaml'); yaml.safe_dump(y,open(tmp,'w'))
            p=plan(d/'pipeline-live.json',d/'source-run.json',d/'source-taskruns.json',tmp,sel)
            arms={arm:disp(json.load(open(d/f'{arm}-plan.json'))) for arm in ('empty','closure')}
            m=[arm for arm,v in arms.items() if v==disp(p)]
            res.append((m[0] if m else 'unexecuted', [w['code'] for w in p.get('warnings') or []]))
        out[f'{mech}-a{a}']=res
# IV-H external
L=R/'acceptance/live'
for a in (1,2,3):
    res=[]
    for f in ('publish','sign','scan'):
        p=plan(L/'pipeline.json',L/f/'source-run.json',L/f/'source-taskruns.json',A/f'external-a{a}/policy.yaml')
        for w in ('stable','rotated'):
            arms={arm:disp(json.load(open(L/f/w/arm/'plan.json'))) for arm in ('empty','declared')}
            m=[arm for arm,v in arms.items() if v==disp(p)]
            res.append((f,w,m[0] if m else 'unexecuted'))
    out[f'external-a{a}']=res
# OCI
O=R/'public-oci'
for prof,arms in (('tagged',('empty','retag','rebuild')),('digest',('empty','rebuild'))):
    for a in (1,2,3):
        p=plan(O/prof/'pipeline.json',O/prof/'source-run.json',O/prof/'source-taskruns.json',A/f'oci-{prof}-a{a}/policy.yaml')
        res=[]
        for w in ('tag-moved','manifest-missing'):
            ex={arm:disp(json.load(open(O/prof/w/arm/'plan.json'))) for arm in arms}
            m=[arm for arm,v in ex.items() if v==disp(p)]
            res.append((w,m[0] if m else 'unexecuted', sorted(t['name'] for t in p['tasks'] if t['action']=='rerun')))
        out[f'oci-{prof}-a{a}']=res
for k,v in out.items(): print(k,v)
