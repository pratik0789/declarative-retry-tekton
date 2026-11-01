import json, re, subprocess, sys, pathlib, yaml, collections
B = sys.argv[1]
S = pathlib.Path('/tmp/docreview/study'); R = pathlib.Path('results')
sits = json.load(open(S / 'operator-situations.json'))
disp = lambda p: (p['decision'], tuple(sorted((t['name'], t['action']) for t in p.get('tasks') or [])))
empty_tpl = {'apiVersion': 'retry.tekton.dev/v1alpha1', 'kind': 'PipelineRetryPolicy', 'metadata': {'name': 'empty', 'generation': 1}, 'spec': {'tasks': {}}}


def plan(pipe, run, trs, pref, sel):
    pol = dict(empty_tpl); pol['spec'] = {'pipelineRef': pref, 'tasks': {}}
    pf = pathlib.Path('/tmp/docreview/study/tmp-empty.yaml'); yaml.safe_dump(pol, open(pf, 'w'))
    a = [B, 'plan', '--pipeline', str(pipe), '--run', str(run), '--taskruns', str(trs), '--policy', str(pf)]
    if sel:
        a += ['--tasks', ','.join(sorted(sel))]
    return json.loads(subprocess.run(a, capture_output=True, text=True).stdout)


def selection(d):
    t = open(d / 'answer.md').read()
    m = re.findall(r'RERUN:\s*\[([^\]]*)\]', t)
    if not m:
        return None
    return sorted({x.strip().strip('`"\'') for x in m[-1].split(',') if x.strip()})


out = collections.OrderedDict(); konflux = collections.defaultdict(list)
for i, s in enumerate(sits):
    for o in (1, 2, 3):
        d = S / 'operators' / f"{i:02d}-{s['pipe']}-{s['failed']}-o{o}"
        sel = selection(d)
        key = f"{s['pipe']}:{s['failed']}"
        if sel is None:
            out.setdefault(key, []).append(('unparsed',)); continue
        if s['pipe'] == 'konflux':
            konflux[s['failed']].append(tuple(sel)); continue
        res = []
        if s['pipe'] in ('partial-write', 'stale-build'):
            pfx = 'pw' if s['pipe'] == 'partial-write' else 'fi'
            extra = ['prepare'] if pfx == 'fi' else []
            for k in range(1, 6):
                c = R / 'final-version-20261030/closure' / f'{pfx}-{k}'
                p = plan(c / 'pipeline-live.json', c / 'source-run.json', c / 'source-taskruns.json', f'{pfx}-{k}', sel + extra)
                ex = {arm: disp(json.load(open(c / f'{arm}-plan.json'))) for arm in ('empty', 'closure')}
                m2 = [a for a, v in ex.items() if v == disp(p)]
                res.append(m2[0] if m2 else 'unexecuted')
        elif s['pipe'] == 'external':
            L = R / 'acceptance/live'
            p = plan(L / 'pipeline.json', L / s['failed'] / 'source-run.json', L / s['failed'] / 'source-taskruns.json', 'external-recovery', sel)
            for w in ('stable', 'rotated'):
                ex = {arm: disp(json.load(open(L / s['failed'] / w / arm / 'plan.json'))) for arm in ('empty', 'declared')}
                full = set(t['name'] for t in json.load(open(L / s['failed'] / w / 'declared' / 'plan.json'))['tasks'])
                m2 = [a for a, v in ex.items() if v == disp(p)]
                rer = sorted(t['name'] for t in p['tasks'] if t['action'] == 'rerun')
                res.append((w, m2[0] if m2 else ('full' if set(rer) >= full - {'deploy'} else 'unexecuted'), rer))
        else:
            prof = s['pipe'].split('-')[1]; O = R / 'public-oci' / prof
            arms = ('empty', 'retag', 'rebuild') if prof == 'tagged' else ('empty', 'rebuild')
            p = plan(O / 'pipeline.json', O / 'source-run.json', O / 'source-taskruns.json', f'public-oci-{prof}', sel)
            for w in ('tag-moved', 'manifest-missing'):
                ex = {arm: disp(json.load(open(O / w / arm / 'plan.json'))) for arm in arms}
                m2 = [a for a, v in ex.items() if v == disp(p)]
                res.append((w, m2[0] if m2 else 'unexecuted', sorted(t['name'] for t in p['tasks'] if t['action'] == 'rerun')))
        out.setdefault(key, []).append((sel, res))
for k, v in out.items():
    print(k)
    for x in v:
        print('   ', x)
print('KONFLUX per-position selections (3 operators):')
agree = 0
for pos, sels in konflux.items():
    same = len(set(sels)) == 1
    agree += same
    print('   ', pos, 'agree' if same else 'DIFFER', [list(x) for x in sels])
print('konflux positions with identical selections from all three operators:', agree, 'of', len(konflux))
json.dump({'scored': out, 'konflux': konflux}, open(S / 'operator-scores.json', 'w'), indent=1, default=list)
