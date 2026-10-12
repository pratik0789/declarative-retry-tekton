#!/usr/bin/env python3
"""Verify retained objects/bytes and compare correct manual selection plans."""
import argparse
import base64
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
    checks = []
    results = {}
    for mechanism, prefix in [('partial-write','pw'),('forward-invalidation','fi')]:
        counts = {a: {'n': 0, 'succeeded': 0, 'consistentReceipts': 0, 'taskRuns': []} for a in ['empty','closure','full']}
        for rep in range(1, protocol['repetitionsPerMechanism']+1):
            d = root/f'{prefix}-{rep}'
            snapshot = json.loads((d/'source-snapshot.json').read_text())
            before_hashes = {f: hashlib.sha256(base64.b64decode(b)).hexdigest() for f,b in snapshot.items()}
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
                submitted = json.loads((d/f'{arm}-submitted.json').read_text())
                assert submitted['spec']['workspaces'][0]['persistentVolumeClaim']['claimName']==f'{prefix}-{rep}-{arm}'
                if arm=='empty':
                    marker = 'DIRTY-INPUT' if prefix=='pw' else 'STALE-ARTIFACT'
                    assert any(marker in x for x in log.values())
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
    print(json.dumps(out,indent=2))


if __name__=='__main__':
    main()
