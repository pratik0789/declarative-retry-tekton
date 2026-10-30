"""Recompute published outcomes from native conditions and external snapshots.
This checker does not use retryctl's predicted metrics or the observations ledger
as its outcome oracle. It is still project-authored, not independent human coding.
"""
import hashlib,json,pathlib,collections
root=pathlib.Path(__file__).resolve().parents[2];live=root/'results/acceptance/live';rows=[]
def read(p):return json.loads(p.read_text())
for failure in ['publish','sign','scan']:
 source=read(live/failure/'source-run.json');tasks=read(live/failure/'source-taskruns.json')
 obs_hash=hashlib.sha256((json.dumps(source,sort_keys=True)+json.dumps(tasks,sort_keys=True)).encode()).hexdigest()
 for world in ['stable','rotated']:
  for arm in ['empty','declared','manual','full']:
   base=live/failure/world/arm
   r=read(base/'run.json');trs=read(base/'taskruns.json')['items'];state=read(base/'external-state.json')
   status=next(c['status'] for c in r['status']['conditions'] if c['type']=='Succeeded')
   initial=read(live/failure/world/'initial-state.json')
   assert state['epoch']==initial['epoch'] and state['injected'] and state['fail_at']==failure
   assert state['events'][:len(initial['events'])]==initial['events']
   suffix=state['events'][len(initial['events']):]
   order=['acquire','publish','sign','scan','deploy']
   expected_tasks=([failure] if world=='rotated' else order[order.index(failure):]) if arm=='empty' else order
   assert [e['task'] for e in suffix]==expected_tasks
   expected_http=409 if world=='rotated' and arm=='empty' else 200
   assert all(e['http']==expected_http and e['epoch']==initial['epoch'] for e in suffix)
   assert len(trs)==len(suffix)+(1 if arm=='full' else 0)
   valid=all(state[x]==initial['epoch'] for x in ['credential','artifact','signature','scanned','deployed'])
   expected=world=='stable' or arm!='empty';assert(status=='True')==expected and valid==expected
   row={'failure':failure,'world':world,'arm':arm,'succeeded':status=='True','external_valid':valid,'taskruns':len(trs),'source_observation_sha256':obs_hash}
   if arm!='full':
    plan=read(base/'plan.json');assert not plan.get('warnings');assert plan['sourceRunUID']==source['metadata']['uid'];row['warnings']=len(plan.get('warnings',[]))
   rows.append(row)
  declared=read(live/failure/world/'declared/plan.json');manual=read(live/failure/world/'manual/plan.json')
  actions=lambda p:{t['name']:t['action'] for t in p['tasks']}
  assert actions(declared)==actions(manual)
summary={'arms':len(rows),'failed_sources':3,'repetitions_per_failure_world_arm':1,'identical_source_pairs':3,'zero_warning_plans':18,'manual_declared_disposition_matches':6,'groups':[],'rows':rows}
for world in ['stable','rotated']:
 for arm in ['empty','declared','manual','full']:
  subset=[r for r in rows if r['world']==world and r['arm']==arm]
  summary['groups'].append({'world':world,'arm':arm,'successes':sum(r['succeeded'] for r in subset),'valid_external_states':sum(r['external_valid'] for r in subset),'cases':len(subset),'total_taskruns':sum(r['taskruns'] for r in subset)})
# State-aware minimum comparator reuses existing recorded arms, not additional trials.
minimum=[r for r in rows if (r['world']=='stable' and r['arm']=='empty') or (r['world']=='rotated' and r['arm']=='manual')]
summary['state_aware_minimum_from_existing_arms']={'cases':len(minimum),'successes':sum(r['succeeded'] for r in minimum),'total_taskruns':sum(r['taskruns'] for r in minimum),'additional_trials':0}
active=read(live/'active/active-plan.json');terminal=read(live/'active/terminal-plan.json');assert active['decision']=='refuse' and terminal['decision']=='recover'
refusal=read(live/'active/create-refusal.json');assert refusal['exit_code']!=0 and not refusal['stdout'];summary['active_source_boundary']={'active':active['decision'],'terminal':terminal['decision'],'active_create_manifest_bytes':len(refusal['stdout'])}
path=root/'results/acceptance/audit.json';path.write_text(json.dumps(summary,indent=2,sort_keys=True)+'\n');print(json.dumps({k:v for k,v in summary.items() if k!='rows'},indent=2))
