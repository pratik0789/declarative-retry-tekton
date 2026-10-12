#!/usr/bin/env python3
"""Recompute central manuscript measurements from retained raw artifacts."""
import argparse
import csv
import hashlib
import json
import math
import pathlib
import statistics

import scipy.stats
import yaml


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--output', required=True)
    args = ap.parse_args()
    root = pathlib.Path(__file__).resolve().parents[1]
    checks = []

    def check(name, actual, expected, tolerance=0):
        ok = abs(actual - expected) <= tolerance if isinstance(actual, (float, int)) and isinstance(expected, (float, int)) else actual == expected
        checks.append({'check': name, 'actual': actual, 'expected': expected, 'reproduced': bool(ok)})

    dataset = root / 'results/v0.3.0-ci-60-trials-kind-v1.31.4-tekton-v1.6.0/trials.csv'
    rows = list(csv.DictReader(dataset.open()))
    check('timing paired trials', len(rows), 60)
    expected = {
        (5, 'early'): (25.187, 20.111, 1, 19.94, 14.61, 25.27),
        (5, 'middle'): (25.500, 14.957, 2, 41.29, 39.97, 42.61),
        (5, 'late'): (25.109, 4.821, 4, 80.65, 78.08, 83.23),
        (10, 'early'): (49.934, 44.755, 1, 10.06, 4.96, 15.16),
        (10, 'middle'): (49.402, 30.518, 4, 38.14, 35.08, 41.20),
        (10, 'late'): (48.292, 5.024, 9, 89.57, 89.18, 89.97),
    }
    timings = []
    for (size, position), values in expected.items():
        group = [r for r in rows if int(r['size']) == size and r['position'] == position]
        check(f'{size}/{position} trial count', len(group), 10)
        baseline = [int(r['baselineMs']) / 1000 for r in group]
        recovery = [int(r['recoveryMs']) / 1000 for r in group]
        reductions = [100 * (b - r) / b for b, r in zip(baseline, recovery)]
        mean = statistics.mean(reductions)
        delta = scipy.stats.t.ppf(.975, len(group) - 1) * statistics.stdev(reductions) / math.sqrt(len(group))
        results = [statistics.mean(baseline), statistics.mean(recovery), int(group[0]['tasksAvoided']), mean, mean - delta, mean + delta]
        for field, actual, value in zip(['restart mean s', 'recovery mean s', 'avoided tasks', 'paired reduction %', 'CI lower %', 'CI upper %'], results, values):
            check(f'{size}/{position} {field}', actual, value, .00051 if field.endswith('s') else .006)
        check(f'{size}/{position} TaskRun counts consistent', all(int(r['baselineTasks']) == size and int(r['recoveryTasks']) == size-values[2] for r in group), True)
        timings.append({'size': size, 'position': position, 'n': len(group), 'restartMeanSeconds': results[0], 'recoveryMeanSeconds': results[1],
                        'tasksAvoided': values[2], 'reductionMeanPercent': mean, 'ci95': [mean-delta, mean+delta],
                        'negativeReductionTrials': sum(r < 0 for r in reductions)})
    pipeline = yaml.safe_load((root/'evaluation/konflux/template-build.yaml').read_text())
    tasks = pipeline['spec']['tasks']
    check('Konflux task count', len(tasks), 16)
    check('Konflux Workspace users', sum(bool(t.get('workspaces')) for t in tasks), 8)
    replays = {}
    for model in ['upstream', 'independent']:
        prefix = 'replay-' if model == 'upstream' else 'replay-independent-'
        for policy in ['policy', 'failed-task-only']:
            for binding in ['persistentVolumeClaim', 'volumeClaimTemplate']:
                name = prefix+policy+'-'+binding
                data = json.loads((root/'evaluation/konflux'/ (name+'.json')).read_text())
                replays[name] = data
                check(name+' failures', len(data['rows']), 16)
                check(name+' dispositions partition', all(len(r['rerun'])+r['inherited']+r['continuing']==16 for r in data['rows']), True)
        declared = replays[prefix+'policy-persistentVolumeClaim']['rows']
        empty = replays[prefix+'failed-task-only-persistentVolumeClaim']['rows']
        check(model+' changed closures', sum(a['rerun']!=b['rerun'] for a,b in zip(declared,empty)), 3)
        check(model+' declared closure sizes', [len(r['rerun']) for r in declared[:5]], [1,1,2,3,4])
        check(model+' later inherited', sorted(set(r['inherited'] for r in declared[5:])), [5 if model=='upstream' else 15])
        check(model+' declared PVC warning failures', sum(bool(r.get('warnings')) for r in declared[5:]), 5)
        check(model+' empty PVC warning failures', sum(bool(r.get('warnings')) for r in empty), 7)
        check(model+' declared VCT warning failures', sum(bool(r.get('warnings')) for r in replays[prefix+'policy-volumeClaimTemplate']['rows'][5:]), 11 if model=='upstream' else 5)
        check(model+' empty VCT warning failures', sum(bool(r.get('warnings')) for r in replays[prefix+'failed-task-only-volumeClaimTemplate']['rows']), 14 if model=='upstream' else 8)
    summaries = list((root/'results/v0.3.0-kind-v1.31.4-tekton-v1.6.0/stateful').glob('run-*/summary.json'))
    summaries += [root/'results/v0.3.1-kind-v1.31.4-tekton-v1.6.0/stateful/summary.json']
    check('retained stateful trials', len(summaries), 7)
    for p in summaries:
        data = json.loads(p.read_text())
        runlist = yaml.safe_load((p.parent/'pipelineruns.yaml').read_text())['items']
        trlist = yaml.safe_load((p.parent/'taskruns.yaml').read_text())['items']
        for key, success in [('persistentClaim', True), ('volumeClaimTemplate', False)]:
            case = data[key]
            run = next(r for r in runlist if r['metadata']['name']==case['recoveryRun'])
            actual = next(c['status'] for c in run['status']['conditions'] if c['type']=='Succeeded')=='True'
            check(str(p.relative_to(root))+' '+key+' status', actual, success)
        pvc = data['persistentClaim']
        source_tr = next(t for t in trlist if t['metadata']['labels'].get('tekton.dev/pipelineRun')=='retry-stateful-pvc-source' and t['metadata']['labels'].get('tekton.dev/pipelineTask')=='clone')
        build_tr = next(t for t in trlist if t['metadata']['labels'].get('tekton.dev/pipelineRun')==pvc['recoveryRun'] and t['metadata']['labels'].get('tekton.dev/pipelineTask')=='build')
        source_commit = next(r['value'] for r in source_tr['status']['results'] if r['name']=='commit')
        reused_commit = next(r['value'] for r in build_tr['spec']['params'] if r['name']=='commit')
        check(str(p.relative_to(root))+' literal commit', reused_commit, source_commit)
        check(str(p.relative_to(root))+' verified digest log', 'STATE-VERIFIED:' in pvc['buildLog'], True)
        check(str(p.relative_to(root))+' missing state log', 'STATE-MISSING:' in data['volumeClaimTemplate']['buildLog'], True)
    report = {'checks': checks, 'reproduced': sum(c['reproduced'] for c in checks), 'failed': sum(not c['reproduced'] for c in checks),
              'timings': timings, 'rawTimingSHA256': hashlib.sha256(dataset.read_bytes()).hexdigest(),
              'scope': 'Checks retained measurement artifacts, not original cluster execution or inaccessible external citations.'}
    pathlib.Path(args.output).write_text(json.dumps(report, indent=2)+'\n')
    print(json.dumps({k:v for k,v in report.items() if k!='checks'}, indent=2))
    for c in checks:
        if not c['reproduced']:
            print('FAILED', c)
    if report['failed']:
        raise SystemExit(1)


if __name__ == '__main__':
    main()
