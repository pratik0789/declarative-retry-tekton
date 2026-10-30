#!/usr/bin/env python3
"""Verify finally-only, upstream, and parallel-closure recovery on live Tekton."""
import argparse
import hashlib
import json
import pathlib
import subprocess
import time


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--context', default='kind-retry-paper')
    ap.add_argument('--namespace', required=True)
    ap.add_argument('--retryctl', default='/tmp/retryctl')
    ap.add_argument('--output', required=True)
    args = ap.parse_args()
    out = pathlib.Path(args.output)
    out.mkdir(parents=True, exist_ok=False)
    k = ['kubectl', '--context', args.context, '-n', args.namespace]

    def cmd(argv, data=None):
        for attempt in range(5):
            p = subprocess.run(argv, input=data, capture_output=True, text=True)
            if p.returncode == 0:
                return p.stdout
            read = data is None and 'get' in argv
            if not (read and 'timed out' in p.stderr and attempt < 4):
                raise RuntimeError(f'{argv}: {p.stderr}')
            time.sleep(attempt + 1)

    def save(path, data):
        path.write_text(json.dumps(data, indent=2) + '\n')

    def create(obj):
        return json.loads(cmd(k + ['create', '-f', '-', '-o', 'json'], json.dumps(obj)))

    def wait(name):
        deadline = time.monotonic() + 240
        while time.monotonic() < deadline:
            run = json.loads(cmd(k + ['get', 'pipelinerun', name, '-o', 'json']))
            for condition in run.get('status', {}).get('conditions', []):
                if condition['type'] == 'Succeeded' and condition['status'] in ('True', 'False'):
                    return run, condition['status'] == 'True'
            time.sleep(1)
        raise RuntimeError('PipelineRun timeout: ' + name)

    def task(name, script, result=None, params=None, after=None):
        spec = {'steps': [{'name': 'run', 'image': 'alpine:3.20', 'script': 'set -eu\n' + script}]}
        if result:
            spec['results'] = [{'name': result}]
        obj = {'name': name, 'taskSpec': spec}
        if params:
            obj['params'] = params
            spec['params'] = [{'name': p['name'], 'type': 'string'} for p in params]
        if after:
            obj['runAfter'] = [after]
        return obj

    create({'apiVersion': 'v1', 'kind': 'Namespace', 'metadata': {'name': args.namespace}})
    save(out / 'build-metadata.json', {'context': args.context, 'namespace': args.namespace,
         'binarySHA256': hashlib.sha256(pathlib.Path(args.retryctl).read_bytes()).hexdigest(),
         'harnessSHA256': hashlib.sha256(pathlib.Path(__file__).read_bytes()).hexdigest()})
    save(out / 'kubernetes-version.json', json.loads(cmd(k + ['version', '-o', 'json'])))
    save(out / 'tekton-version.json', json.loads(cmd(['kubectl', '--context', args.context, '-n', 'tekton-pipelines', 'get', 'configmap', 'pipelines-info', '-o', 'json'])))
    cases = [
        ('only', {}, {'notify'}),
        ('upstream', {'notify': {'retryWith': ['build']}, 'build': {'retryWith': ['prepare']}}, {'prepare', 'build', 'notify'}),
        ('parallel', {'notify': {'retryWith': ['audit']}}, {'notify', 'audit'}),
    ]
    summary = []
    for name, rules, expected in cases:
        d = out / name
        d.mkdir()
        source_name, recovery_name = 'finally-' + name + '-source', 'finally-' + name + '-recovery'
        pipeline_name = 'finally-' + name
        tasks = [
            task('prepare', 'printf "%s" "$(context.taskRun.name)" > "$(results.revision.path)"\n', 'revision'),
            task('build', 'printf "image:%s" "$(params.revision)" > "$(results.image.path)"\n', 'image',
                 [{'name': 'revision', 'value': '$(tasks.prepare.results.revision)'}], 'prepare'),
        ]
        final = [
            task('notify', 'case "$(context.taskRun.name)" in *-source-*) echo INJECTED-NOTIFICATION-FAILURE; exit 1;; esac\n'
                 'printf "%s" "$(params.image)" > "$(results.receipt.path)"\necho NOTIFIED\n', 'receipt',
                 [{'name': 'image', 'value': '$(tasks.build.results.image)'}]),
            task('audit', 'echo AUDITED\n'),
        ]
        pipeline = create({'apiVersion': 'tekton.dev/v1', 'kind': 'Pipeline', 'metadata': {'name': pipeline_name},
                           'spec': {'tasks': tasks, 'finally': final}})
        save(d / 'pipeline.json', pipeline)
        create({'apiVersion': 'tekton.dev/v1', 'kind': 'PipelineRun', 'metadata': {'name': source_name},
                'spec': {'pipelineRef': {'name': pipeline_name}}})
        source, source_ok = wait(source_name)
        assert not source_ok
        source_trs = json.loads(cmd(k + ['get', 'taskruns', '-l', 'tekton.dev/pipelineRun=' + source_name, '-o', 'json']))
        statuses = {t['metadata']['labels']['tekton.dev/pipelineTask']: next(c['status'] for c in t['status']['conditions'] if c['type'] == 'Succeeded') for t in source_trs['items']}
        assert statuses == {'prepare': 'True', 'build': 'True', 'notify': 'False', 'audit': 'True'}, statuses
        save(d / 'source-run.json', source)
        save(d / 'source-taskruns.json', source_trs)
        policy = {'apiVersion': 'retry.tekton.dev/v1alpha1', 'kind': 'PipelineRetryPolicy', 'metadata': {'name': name},
                  'spec': {'pipelineRef': pipeline_name, 'tasks': rules}}
        save(d / 'policy.json', policy)
        inputs = ['--pipeline', str(d / 'pipeline.json'), '--run', str(d / 'source-run.json'),
                  '--taskruns', str(d / 'source-taskruns.json'), '--policy', str(d / 'policy.json')]
        plan = json.loads(cmd([args.retryctl, 'plan'] + inputs))
        assert plan['decision'] == 'recover'
        assert {t['name'] for t in plan['tasks'] if t['action'] != 'inherit'} == expected
        save(d / 'plan.json', plan)
        manifest = cmd([args.retryctl, 'create'] + inputs + ['--confirm-warnings'])
        (d / 'generated.yaml').write_text(manifest)
        submitted = json.loads(cmd(k + ['create', '--dry-run=client', '-f', '-', '-o', 'json'], manifest))
        submitted['metadata'].pop('generateName', None)
        submitted['metadata']['name'] = recovery_name
        save(d / 'submitted.json', submitted)
        create(submitted)
        recovered, success = wait(recovery_name)
        assert success, recovered['status']['conditions']
        trs = json.loads(cmd(k + ['get', 'taskruns', '-l', 'tekton.dev/pipelineRun=' + recovery_name, '-o', 'json']))
        by_name = {t['metadata']['labels']['tekton.dev/pipelineTask']: t for t in trs['items']}
        assert set(by_name) == expected, set(by_name)
        receipt = next(r['value'] for r in by_name['notify']['status']['results'] if r['name'] == 'receipt')
        revision_trs = by_name if name == 'upstream' else {t['metadata']['labels']['tekton.dev/pipelineTask']: t for t in source_trs['items']}
        expected_receipt = 'image:' + revision_trs['prepare']['metadata']['name']
        assert receipt == expected_receipt, (receipt, expected_receipt)
        if name == 'upstream':
            assert by_name['notify']['status']['startTime'] >= by_name['build']['status']['completionTime']
        else:
            anchor = submitted['metadata']['annotations']['retry.tekton.dev/finally-anchor']
            assert {t['name'] for t in recovered['status']['skippedTasks']} == {anchor}
        save(d / 'recovery-run.json', recovered)
        save(d / 'recovery-taskruns.json', trs)
        save(d / 'logs.json', {task: cmd(k + ['logs', tr['status']['podName'], '--all-containers=true']) for task, tr in by_name.items()})
        row = {'case': name, 'sourceFailedOnlyInFinally': True, 'recoverySucceeded': success,
               'executedTasks': sorted(by_name), 'taskRuns': len(by_name), 'receipt': receipt,
               'expectedReceipt': expected_receipt, 'finalPhasePreserved': True}
        summary.append(row)
        print(json.dumps(row), flush=True)
    save(out / 'summary.json', {'cases': summary, 'scope': 'Three synthetic finally-task recovery cases; no timing or production effectiveness claim.'})


if __name__ == '__main__':
    main()
