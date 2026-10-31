#!/usr/bin/env python3
"""Verify retained objects/bytes and compare correct manual selection plans."""
import argparse
import base64
import copy
import hashlib
import json
from pathlib import Path
import subprocess


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--results', required=True)
    ap.add_argument('--retryctl', default='/tmp/retryctl')
    args = ap.parse_args()
    root = Path(args.results)
    protocol = json.loads((root/'protocol.json').read_text())
    metadata = json.loads((root/'build-metadata.json').read_text())
    assert metadata['binarySHA256'] == hashlib.sha256(Path(args.retryctl).read_bytes()).hexdigest(), 'verify with the evaluated binary'
    checks = []
    guard_checks = []
    results = {}
    stale_mechanism = 'stale-build' if 'stale-build' in protocol['mechanisms'] else 'forward-invalidation'
    for mechanism, prefix in [('partial-write','pw'),(stale_mechanism,'fi')]:
        counts = {a: {'n': 0, 'succeeded': 0, 'consistentReceipts': 0, 'taskRuns': []} for a in ['empty','closure','full']}
        for rep in range(1, protocol['repetitionsPerMechanism']+1):
            d = root/f'{prefix}-{rep}'
            snapshot = json.loads((d/'source-snapshot.json').read_text())
            before_hashes = {f: hashlib.sha256(base64.b64decode(b)).hexdigest() for f,b in snapshot.items()}
            if protocol.get('definitionGuard'):
                pipeline = json.loads((d/'pipeline-live.json').read_text())
                source = json.loads((d/'source-run.json').read_text())
                # Ignore null serialization fields, as the planner does. Do not
                # substitute expressions or rewrite either native capture.
                def without_nulls(value):
                    if isinstance(value, dict):
                        return {k: without_nulls(v) for k, v in value.items() if v is not None}
                    if isinstance(value, list):
                        return [without_nulls(v) for v in value]
                    return value
                current = without_nulls(pipeline['spec'])
                recorded = without_nulls(source['status']['pipelineSpec'])
                assert current == recorded, (d.name, 'native Pipeline specification mismatch')
                saved_hash = hashlib.sha256(json.dumps(recorded, sort_keys=True).encode()).hexdigest()
                for edit in ['script', 'context-expression']:
                    changed = copy.deepcopy(pipeline)
                    task_index = -1 if edit == 'script' else (2 if prefix == 'pw' else 1)
                    step = changed['spec']['tasks'][task_index]['taskSpec']['steps'][0]
                    if edit == 'script':
                        step['script'] += '\necho edited-pipeline\n'
                    else:
                        assert '$(context.taskRun.name)' in step['script']
                        step['script'] = step['script'].replace('$(context.taskRun.name)', '$(context.taskRun.namespace)')
                    changed_path = d/f'guard-changed-{edit}-pipeline.json'
                    changed_path.write_text(json.dumps(changed, indent=2)+'\n')
                    inputs = ['--pipeline', str(changed_path), '--run', str(d/'source-run.json'),
                              '--taskruns', str(d/'source-taskruns.json'), '--policy', str(d/'closure-policy.json')]
                    refused = json.loads(subprocess.check_output([args.retryctl, 'plan']+inputs, text=True))
                    assert refused['decision'] == 'refuse' and 'Pipeline definition changed' in refused['refusalReason']
                    generated = subprocess.run([args.retryctl, 'create']+inputs+['--confirm-warnings'], capture_output=True, text=True, timeout=60)
                    assert generated.returncode != 0 and not generated.stdout and 'Pipeline definition changed' in generated.stderr
                    guard_checks.append({'case': d.name, 'edit': edit, 'nativeSpecificationSHA256': saved_hash,
                                         'plan': refused, 'createExitCode': generated.returncode, 'createStderr': generated.stderr,
                                         'manifestProduced': False})
            if protocol.get('retryWithDirection'):
                policy = json.loads((d/'closure-policy.json').read_text())['spec']['tasks']
                expected = {'mutate': {'retryWith': ['prepare']}} if prefix=='pw' else {'gate': {'retryWith': ['prepare','build']}}
                assert policy == expected, (d.name, policy, expected)
            for arm in counts:
                run = json.loads((d/f'{arm}-run.json').read_text())
                trs = json.loads((d/f'{arm}-taskruns.json').read_text())['items']
                log = json.loads((d/f'{arm}-logs.json').read_text())
                values = {f:base64.b64decode(b) for f,b in json.loads((d/f'{arm}-workspace.json').read_text()).items()}
                succeeded = next(c['status'] for c in run['status']['conditions'] if c['type']=='Succeeded')=='True'
                consistent = ('receipt.txt' in values and values.get('artifact-input.sha') == hashlib.sha256(values['source.txt']).hexdigest().encode()
                              and values.get('artifact.txt')==b'built:'+values['artifact-input.sha']
                              and values['receipt.txt'].decode()==hashlib.sha256(values['source.txt']).hexdigest()+' '+hashlib.sha256(values['artifact.txt']).hexdigest())
                assert succeeded == (arm!='empty') and consistent == (arm!='empty')
                expected_names = {'observe', 'prepare', 'publish', 'mutate' if prefix=='pw' else 'build', 'verify' if prefix=='pw' else 'gate'}
                if arm == 'closure':
                    expected_names.remove('observe')
                elif arm == 'empty':
                    expected_names = {'mutate'} if prefix=='pw' else {'prepare', 'gate', 'publish'}
                by_name = {t['metadata']['labels']['tekton.dev/pipelineTask']: t for t in trs}
                assert set(by_name) == expected_names and len(trs) == len(expected_names), (d.name, arm, set(by_name))
                failed_name = 'mutate' if prefix=='pw' else 'publish'
                assert all(next(c['status'] for c in t['status']['conditions'] if c['type']=='Succeeded') ==
                           ('False' if arm=='empty' and name==failed_name else 'True') for name,t in by_name.items())
                if prefix == 'fi':
                    assert values['source.txt'] != base64.b64decode(snapshot['source.txt']), (d.name, arm, 'input did not refresh')
                submitted = json.loads((d/f'{arm}-submitted.json').read_text())
                assert submitted['spec']['workspaces'][0]['persistentVolumeClaim']['claimName']==f'{prefix}-{rep}-{arm}'
                if arm=='empty':
                    marker = 'DIRTY-INPUT' if prefix=='pw' else 'STALE-ARTIFACT'
                    assert any(marker in x for x in log.values())
                if arm != 'full':
                    plan = json.loads((d/f'{arm}-plan.json').read_text())
                    assert plan['decision'] == 'recover'
                    assert [w['code'] for w in plan.get('warnings', [])] == (['shared-workspace-across-closure'] if arm=='empty' else [])
                    if arm == 'empty':
                        assert submitted['metadata']['annotations']['retry.tekton.dev/shared-state-warnings-accepted'] == 'true'
                counts[arm]['n'] += 1
                counts[arm]['succeeded'] += int(succeeded)
                counts[arm]['consistentReceipts'] += int(consistent)
                counts[arm]['taskRuns'].append(len(trs))
            # Compare the automatic closure to a knowledgeable operator naming its members.
            selected = ['prepare'] if prefix=='pw' else ['prepare','build']
            command = [args.retryctl, 'plan', '--pipeline',str(d/'pipeline-live.json'), '--run',str(d/'source-run.json'),
                       '--taskruns',str(d/'source-taskruns.json'),'--policy',str(d/'empty-policy.json'),'--tasks',','.join(selected)]
            manual = json.loads(subprocess.check_output(command,text=True))
            declared = json.loads((d/'closure-plan.json').read_text())
            actions = lambda plan:{t['name']:t['action'] for t in plan['tasks']}
            assert manual['decision']==declared['decision']=='recover' and actions(manual)==actions(declared)
            (d/'correct-manual-selection-plan.json').write_text(json.dumps(manual,indent=2)+'\n')
            checks.append({'case':d.name,'manualSelection':selected,'sameDispositions':True,'sourceSnapshotHashes':before_hashes})
        results[mechanism]=counts
    trials=json.loads((root/'trials.json').read_text())
    assert len(trials)==protocol['repetitionsPerMechanism']*2*3
    for trial in trials:
        snapshot=json.loads((root/f'{"pw" if trial["mechanism"]=="partial-write" else "fi"}-{trial["repetition"]}'/'source-snapshot.json').read_text())
        assert trial['beforeHashes']=={f:hashlib.sha256(base64.b64decode(b)).hexdigest() for f,b in snapshot.items()}
    out={'nativeObjectsAndBytesVerified':True,'triplets':len(checks),'arms':len(trials),'manualSelectionComparisons':checks,'recomputedResults':results,
         'oracleScope':'Local consistency only; task checks and policy semantics share authorship. No independent human policy labels.'}
    (root/'verification.json').write_text(json.dumps(out,indent=2)+'\n')
    if protocol.get('definitionGuard'):
        (root/'definition-guard-verification.json').write_text(json.dumps({'nativeSpecificationsMatch': True, 'changedDefinitionsRefused': guard_checks}, indent=2)+'\n')
    print(json.dumps(out,indent=2))


if __name__=='__main__':
    main()
