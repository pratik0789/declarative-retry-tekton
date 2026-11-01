# Plan-level agreement: three blind authors' Konflux policies vs three per-failure operators.
import json, subprocess, sys
B = sys.argv[1]
A = 'results/independent-authors-20261101/authors'
plans = {}
for a in (1, 2, 3):
    r = json.loads(subprocess.run([B, 'replay', '--pipeline', 'evaluation/konflux/template-build.yaml', '--policy', f'{A}/konflux-a{a}/policy.yaml', '--completed', 'upstream'], capture_output=True, text=True).stdout)
    plans[a] = {row['failedTask']: tuple(sorted(row.get('rerun') or [])) for row in r['rows']}
ops = json.load(open('results/blind-operators-20261101/operator-scores.json'))['konflux']
print(json.dumps({
    'konfluxPositions': len(plans[1]),
    'authorPlanDisagreements': [p for p in plans[1] if len({plans[a][p] for a in (1, 2, 3)}) > 1],
    'operatorSelectionDisagreements': [p for p, s in ops.items() if len({tuple(x) for x in s}) > 1],
}, indent=1))
