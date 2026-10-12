#!/usr/bin/env python3
"""Live counterexamples with byte-identical, isolated Workspace forks.

Only creates resources in a new namespace. Never deletes existing resources.
The external oracle reads Workspace bytes and Tekton status, not planner labels.
"""
import argparse
import base64
import hashlib
import json
import pathlib
import random
import subprocess
import time


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--context', default='kind-retry-paper')
    ap.add_argument('--namespace', required=True)
    ap.add_argument('--repetitions', type=int, default=5)
    ap.add_argument('--retryctl', default='/tmp/retryctl')
    ap.add_argument('--output', required=True)
    args = ap.parse_args()
    out = pathlib.Path(args.output)
    out.mkdir(parents=True, exist_ok=True)
    k = ['kubectl', '--context', args.context, '-n', args.namespace]

    def cmd(argv, data=None):
        p = subprocess.run(argv, input=data, text=True, capture_output=True)
        if p.returncode:
            raise RuntimeError(f'{argv}: {p.stderr}\n{p.stdout}')
        return p.stdout

    def save(path, data):
        path.write_text(json.dumps(data, indent=2) + '\n')

    def create(obj):
        return json.loads(cmd(k + ['create', '-f', '-', '-o', 'json'], json.dumps(obj)))

    def get(kind, name=None):
        return json.loads(cmd(k + ['get', kind] + ([name] if name else []) + ['-o', 'json']))

    def wait_run(name, timeout=240):
        deadline = time.monotonic() + timeout
        while time.monotonic() < deadline:
            run = get('pipelinerun', name)
            for c in run.get('status', {}).get('conditions', []):
                if c['type'] == 'Succeeded' and c['status'] in ('True', 'False'):
                    return run, c['status'] == 'True'
            time.sleep(1)
        raise RuntimeError(f'timeout for {name}')

    def pvc(name):
        return create({'apiVersion': 'v1', 'kind': 'PersistentVolumeClaim',
                       'metadata': {'name': name}, 'spec': {'accessModes': ['ReadWriteOnce'],
                       'resources': {'requests': {'storage': '64Mi'}}}})

    def pod(name, claims, script):
        obj = {'apiVersion': 'v1', 'kind': 'Pod', 'metadata': {'name': name},
               'spec': {'restartPolicy': 'Never', 'containers': [{'name': 'io',
                        'image': 'alpine:3.20', 'command': ['sh', '-eu', '-c', script],
                        'volumeMounts': [{'name': f'v{i}', 'mountPath': f'/v{i}'}
                                         for i in range(len(claims))]}],
                        'volumes': [{'name': f'v{i}', 'persistentVolumeClaim': {'claimName': claim}}
                                    for i, claim in enumerate(claims)]}}
        create(obj)
        deadline = time.monotonic() + 120
        while time.monotonic() < deadline:
            phase = get('pod', name).get('status', {}).get('phase')
            if phase in ('Succeeded', 'Failed'):
                log = cmd(k + ['logs', name])
                if phase != 'Succeeded':
                    raise RuntimeError(log)
                return log
            time.sleep(1)
        raise RuntimeError(f'pod timeout {name}')

    files = ['source.txt', 'artifact.txt', 'artifact-input.sha', 'receipt.txt']
    read_script = '\n'.join(f'if [ -f /v0/{f} ]; then printf "{f} "; base64 /v0/{f} | tr -d "\\n"; echo; fi' for f in files)

    def contents(log):
        return {f: base64.b64decode(data) for f, data in (line.split(' ', 1) for line in log.splitlines())}

    def hashes(data):
        return {f: hashlib.sha256(b).hexdigest() for f, b in data.items()}

    def task(name, script, after=None, workspace=True):
        t = {'name': name, 'taskSpec': {'steps': [{'name': 'run', 'image': 'alpine:3.20', 'script': 'set -eu\n' + script}]}}
        if after:
            t['runAfter'] = [after]
        if workspace:
            t['workspaces'] = [{'name': 'data', 'workspace': 'shared'}]
            t['taskSpec']['workspaces'] = [{'name': 'data'}]
        return t

    create({'apiVersion': 'v1', 'kind': 'Namespace', 'metadata': {'name': args.namespace}})
    save(out / 'kubernetes-version.json', json.loads(cmd(k + ['version', '-o', 'json'])))
    save(out / 'tekton-version.json', json.loads(cmd(['kubectl', '--context', args.context, '-n', 'tekton-pipelines', 'get', 'configmap', 'pipelines-info', '-o', 'json'])))
    protocol = {'repetitionsPerMechanism': args.repetitions, 'mechanisms': ['partial-write', 'forward-invalidation'],
                'arms': ['empty', 'closure', 'full'], 'orderSeed': 20261011,
                'sourceSnapshot': 'one failed source per triplet; exact bytes copied to three distinct PVCs',
                'bindingOverride': 'generated recovery manifest is preserved; only claimName and generated name are overridden for isolation',
                'oracle': 'native task assertions plus external SHA-256 over actual Workspace bytes and Tekton status',
                'timing': 'not a timing benchmark; status polling every 1 second',
                'authoring': 'fixtures, policies, and assertions authored with the prototype; no independent author study'}
    save(out / 'protocol.json', protocol)
    rng = random.Random(protocol['orderSeed'])
    rows = []
    for mechanism in protocol['mechanisms']:
        for rep in range(1, args.repetitions + 1):
            case = f'{"pw" if mechanism == "partial-write" else "fi"}-{rep}'
            d = out / case
            d.mkdir()
            pipeline_name = case
            source_name = case + '-source'
            nonce = hashlib.sha256(f'closure-ablation-20261011-{case}'.encode()).hexdigest()[:16]
            w = '$(workspaces.data.path)'
            source_check = f'case "$(context.pipelineRun.name)" in *-source)'
            if mechanism == 'partial-write':
                prepare = f'printf "clean:{nonce}" > {w}/source.txt\nrm -f {w}/artifact.txt {w}/artifact-input.sha {w}/receipt.txt\n'
                mutate = (f'{source_check} printf "dirty:{nonce}" > {w}/source.txt; echo INJECTED-PARTIAL-WRITE; exit 1;; esac\n'
                          f'if [ "`cat {w}/source.txt`" != "clean:{nonce}" ]; then echo DIRTY-INPUT; exit 1; fi\n'
                          f'sha256sum {w}/source.txt | cut -d" " -f1 | tr -d "\\n" > {w}/artifact-input.sha\n'
                          f'printf "built:" > {w}/artifact.txt; cat {w}/artifact-input.sha >> {w}/artifact.txt\n')
                tasks = [task('observe', 'echo OBSERVED\n', workspace=False), task('prepare', prepare, 'observe'),
                         task('mutate', mutate, 'prepare'), task('verify', f'echo VERIFIED-STAGE\n', 'mutate')]
                rule = {'mutate': {'retryWith': ['prepare']}}
                selected = []
            else:
                prepare = (f'printf "input:$(context.pipelineRun.name):{nonce}" > {w}/source.txt\n'
                           f'rm -f {w}/receipt.txt\n')
                build = (f'sha256sum {w}/source.txt | cut -d" " -f1 | tr -d "\\n" > {w}/artifact-input.sha\n'
                         f'printf "built:" > {w}/artifact.txt; cat {w}/artifact-input.sha >> {w}/artifact.txt\n')
                gate = f'{source_check} echo INJECTED-GATE-FAILURE; exit 1;; esac\necho GATE-PASSED\n'
                tasks = [task('observe', 'echo OBSERVED\n', workspace=False), task('prepare', prepare, 'observe'),
                         task('build', build, 'prepare'), task('gate', gate, 'build', workspace=False)]
                rule = {'prepare': {'retryWith': ['build']}}
                selected = ['prepare']
            publish = (f'actual=`sha256sum {w}/source.txt | cut -d" " -f1`\n'
                       f'recorded=`cat {w}/artifact-input.sha`\n'
                       f'if [ "$actual" != "$recorded" ]; then echo STALE-ARTIFACT; exit 1; fi\n'
                       f'if [ "`cat {w}/artifact.txt`" != "built:$actual" ]; then echo BAD-ARTIFACT; exit 1; fi\n'
                       f'artifact=`sha256sum {w}/artifact.txt | cut -d" " -f1`\n'
                       f'printf "%s %s" "$actual" "$artifact" > {w}/receipt.txt\necho ORACLE-CONSISTENT\n')
            tasks.append(task('publish', publish, tasks[-1]['name']))
            pipeline = {'apiVersion': 'tekton.dev/v1', 'kind': 'Pipeline', 'metadata': {'name': pipeline_name},
                        'spec': {'workspaces': [{'name': 'shared'}], 'tasks': tasks}}
            save(d / 'pipeline-input.json', pipeline)
            create(pipeline)
            claims = [case + '-' + arm for arm in ['source', 'empty', 'closure', 'full']]
            for claim in claims:
                pvc(claim)
            source = {'apiVersion': 'tekton.dev/v1', 'kind': 'PipelineRun', 'metadata': {'name': source_name},
                      'spec': {'pipelineRef': {'name': pipeline_name}, 'workspaces': [{'name': 'shared', 'persistentVolumeClaim': {'claimName': claims[0]}}]}}
            create(source)
            source, source_ok = wait_run(source_name)
            assert not source_ok
            save(d / 'source-run.json', source)
            save(d / 'source-taskruns.json', json.loads(cmd(k + ['get', 'taskrun', '-l', 'tekton.dev/pipelineRun=' + source_name, '-o', 'json'])))
            save(d / 'pipeline-live.json', get('pipeline', pipeline_name))
            snapshot_log = pod(case + '-fork', claims, 'cp -a /v0/. /v1/\ncp -a /v0/. /v2/\ncp -a /v0/. /v3/\n' + read_script)
            original = contents(snapshot_log)
            save(d / 'source-snapshot.json', {f: base64.b64encode(b).decode() for f, b in original.items()})
            order = ['empty', 'closure', 'full']
            rng.shuffle(order)
            for arm in order:
                claim = claims[['source', 'empty', 'closure', 'full'].index(arm)]
                before = contents(pod(case + '-' + arm + '-before', [claim], read_script))
                assert before == original, (case, arm, 'unequal fork')
                name = case + '-' + arm + '-recovery'
                if arm == 'full':
                    manifest = {'apiVersion': 'tekton.dev/v1', 'kind': 'PipelineRun', 'metadata': {'name': name},
                                'spec': {'pipelineRef': {'name': pipeline_name}, 'workspaces': [{'name': 'shared', 'persistentVolumeClaim': {'claimName': claim}}]}}
                    plan = None
                else:
                    policy = {'apiVersion': 'retry.tekton.dev/v1alpha1', 'kind': 'PipelineRetryPolicy',
                              'metadata': {'name': case + '-' + arm + '-policy', 'generation': 1},
                              'spec': {'pipelineRef': pipeline_name, 'tasks': rule if arm == 'closure' else {}}}
                    save(d / (arm + '-policy.json'), policy)
                    inputs = ['--pipeline', str(d / 'pipeline-live.json'), '--run', str(d / 'source-run.json'),
                              '--taskruns', str(d / 'source-taskruns.json'), '--policy', str(d / (arm + '-policy.json'))]
                    if selected:
                        inputs += ['--tasks', ','.join(selected)]
                    plan = json.loads(cmd([args.retryctl, 'plan'] + inputs))
                    save(d / (arm + '-plan.json'), plan)
                    generated = cmd([args.retryctl, 'create'] + inputs + ['--confirm-warnings'])
                    (d / (arm + '-generated.yaml')).write_text(generated)
                    manifest = json.loads(cmd(k + ['create', '--dry-run=client', '-f', '-', '-o', 'json'], generated))
                    manifest['metadata'].pop('generateName', None)
                    manifest['metadata']['name'] = name
                    manifest['spec']['workspaces'][0]['persistentVolumeClaim']['claimName'] = claim
                save(d / (arm + '-submitted.json'), manifest)
                create(manifest)
                run, success = wait_run(name)
                save(d / (arm + '-run.json'), run)
                trs = json.loads(cmd(k + ['get', 'taskrun', '-l', 'tekton.dev/pipelineRun=' + name, '-o', 'json']))
                save(d / (arm + '-taskruns.json'), trs)
                logs = {}
                for tr in trs['items']:
                    logs[tr['metadata']['labels']['tekton.dev/pipelineTask']] = cmd(k + ['logs', tr['status']['podName'], '--all-containers=true'])
                save(d / (arm + '-logs.json'), logs)
                after = contents(pod(case + '-' + arm + '-after', [claim], read_script))
                save(d / (arm + '-workspace.json'), {f: base64.b64encode(b).decode() for f, b in after.items()})
                consistent = ('receipt.txt' in after and after.get('artifact-input.sha', b'').decode() == hashlib.sha256(after['source.txt']).hexdigest()
                              and after.get('artifact.txt') == b'built:' + after['artifact-input.sha']
                              and after['receipt.txt'].decode() == hashlib.sha256(after['source.txt']).hexdigest() + ' ' + hashlib.sha256(after['artifact.txt']).hexdigest())
                expected_ok = arm != 'empty'
                assert success == expected_ok and consistent == expected_ok, (case, arm, success, consistent)
                expected_tasks = 5 if arm == 'full' else (4 if arm == 'closure' else (1 if mechanism == 'partial-write' else 3))
                assert len(trs['items']) == expected_tasks, (case, arm, len(trs['items']))
                marker = 'DIRTY-INPUT' if mechanism == 'partial-write' else 'STALE-ARTIFACT'
                if arm == 'empty':
                    assert any(marker in log for log in logs.values())
                row = {'mechanism': mechanism, 'repetition': rep, 'arm': arm, 'armOrder': order,
                       'sourceRun': source_name, 'recoveryRun': name, 'succeeded': success,
                       'externallyConsistent': consistent, 'taskRuns': len(trs['items']),
                       'beforeHashes': hashes(before), 'afterHashes': hashes(after),
                       'emptyFailureMarker': marker if arm == 'empty' else None,
                       'warnings': [x['code'] for x in plan.get('warnings', [])] if plan else [],
                       'selectedTasks': selected if arm != 'full' else []}
                rows.append(row)
                save(out / 'trials.json', rows)
                print(f'{case} {arm}: success={success}, oracle={consistent}, TaskRuns={len(trs["items"])}', flush=True)
    summary = {'protocol': protocol, 'triplets': args.repetitions * 2, 'armsExecuted': len(rows), 'results': {}}
    for mechanism in protocol['mechanisms']:
        summary['results'][mechanism] = {}
        for arm in protocol['arms']:
            chosen = [r for r in rows if r['mechanism'] == mechanism and r['arm'] == arm]
            summary['results'][mechanism][arm] = {'succeeded': sum(r['succeeded'] for r in chosen),
                                                'consistentReceipts': sum(r['externallyConsistent'] for r in chosen),
                                                'n': len(chosen), 'taskRunsPerTrial': sorted(set(r['taskRuns'] for r in chosen))}
    save(out / 'summary.json', summary)
    print(json.dumps(summary, indent=2), flush=True)


if __name__ == '__main__':
    main()
