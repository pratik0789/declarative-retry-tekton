#!/usr/bin/env python3
"""Derive current timing results and verify native paired-run records."""
import argparse
import csv
from datetime import datetime
import hashlib
import json
import math
from pathlib import Path
import statistics


def stamp(value):
    return datetime.fromisoformat(value.replace('Z', '+00:00'))


def terminal(run):
    conditions = [c for c in run['status']['conditions'] if c['type'] == 'Succeeded']
    assert len(conditions) == 1
    return conditions[0]['status']


def audit(directory):
    import yaml
    rows = list(csv.DictReader((directory / 'trials.csv').open()))
    orders = json.loads((directory / 'arm-order.json').read_text())['orders']
    assert len(rows) == 60 and len(orders) == 60
    derived = []
    for row in rows:
        size, repetition = int(row['size']), int(row['repetition'])
        position = row['position']
        case_id = f's{size}-{position[0]}-r{repetition}'
        case = directory / case_id
        source = yaml.safe_load((case / 'source-run.yaml').read_text())
        assert terminal(source) == 'False', case_id
        baseline = json.loads((case / 'baseline-run.json').read_text())
        recovery = json.loads((case / 'recovery-run.json').read_text())
        idx = {'early': 2, 'middle': (size + 1)//2, 'late': size}[position]
        assert int(row['baselineTasks']) == size
        assert int(row['recoveryTasks']) == size - idx + 1
        assert int(row['tasksAvoided']) == idx - 1
        native = {}
        for arm, run, expected in [('baseline', baseline, size), ('recovery', recovery, size-idx+1)]:
            assert terminal(run) == 'True', (case_id, arm)
            children = json.loads((case / f'{arm}-taskruns.json').read_text())['items']
            assert len(children) == expected and all(terminal(t) == 'True' for t in children)
            assert all(t['metadata']['labels']['tekton.dev/pipelineRun'] == run['metadata']['name'] for t in children)
            native[arm] = (stamp(run['status']['completionTime'])-stamp(run['status']['startTime'])).total_seconds()
            assert native[arm] >= 0
        first, second = (baseline, recovery) if orders[case_id] == 'baseline-first' else (recovery, baseline)
        assert stamp(second['status']['startTime']) >= stamp(first['status']['completionTime']), case_id
        manifest = yaml.safe_load((case / 'recovery.yaml').read_text())
        annotation = manifest['metadata']['annotations']
        body = annotation['retry.tekton.dev/policy-spec']
        digest = 'sha256:' + hashlib.sha256(body.encode()).hexdigest()
        assert digest == annotation['retry.tekton.dev/policy-spec-sha256']
        plan = json.loads((case / 'plan.json').read_text())
        assert plan['decision'] == 'recover' and plan['policySpecSHA256'] == digest
        wall_base, wall_recovery = int(row['baselineMs'])/1000, int(row['recoveryMs'])/1000
        assert wall_base > 0 and wall_recovery > 0
        derived.append({'case':case_id,'size':size,'position':position,'repetition':repetition,
                        'order':orders[case_id],'restartWallSeconds':wall_base,'recoveryWallSeconds':wall_recovery,
                        'restartNativeSeconds':native['baseline'],'recoveryNativeSeconds':native['recovery'],
                        'tasksAvoided':idx-1,'pairedWallReductionPercent':100*(wall_base-wall_recovery)/wall_base})
    # t(.975, 9), because the predeclared replication has ten pairs per cell.
    critical = 2.2621571628540993
    summary = []
    for size in [5, 10]:
        for position in ['early','middle','late']:
            group = [r for r in derived if r['size']==size and r['position']==position]
            assert len(group) == 10 and {r['repetition'] for r in group} == set(range(1,11))
            assert sum(r['order']=='baseline-first' for r in group) == 5
            changes = [r['pairedWallReductionPercent'] for r in group]
            mean = statistics.mean(changes)
            half = critical*statistics.stdev(changes)/math.sqrt(len(group))
            summary.append({'size':size,'position':position,'n':10,
                            'restartMeanSeconds':statistics.mean(r['restartWallSeconds'] for r in group),
                            'recoveryMeanSeconds':statistics.mean(r['recoveryWallSeconds'] for r in group),
                            'restartNativeMeanSeconds':statistics.mean(r['restartNativeSeconds'] for r in group),
                            'recoveryNativeMeanSeconds':statistics.mean(r['recoveryNativeSeconds'] for r in group),
                            'tasksAvoided':group[0]['tasksAvoided'],'pairedReductionMeanPercent':mean,
                            'ci95Percent':[mean-half,mean+half],
                            'negativeReductionPairs':sum(v<0 for v in changes)})
    return {'scope':'Descriptive paired local replication of synthetic serial 0.2-second tasks. Confidence intervals describe these repeated cells, not a population of real pipelines or operator benefit.',
            'pairs':derived,'summary':summary,
            'csvSHA256':hashlib.sha256((directory/'trials.csv').read_bytes()).hexdigest(),
            'orderSHA256':hashlib.sha256((directory/'arm-order.json').read_bytes()).hexdigest()}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--directory', required=True, type=Path)
    args = parser.parse_args()
    data = audit(args.directory)
    (args.directory/'audit.json').write_text(json.dumps(data,indent=2)+'\n')
    print(json.dumps({'verifiedPairs':len(data['pairs']),'summary':data['summary']},indent=2))


if __name__ == '__main__':
    main()
