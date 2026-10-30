#!/usr/bin/env python3
"""Wait for actual Tekton admission using a server dry-run; persist no resource."""
import argparse
import json
import subprocess
import time


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--context', required=True)
    ap.add_argument('--kubectl', default='kubectl')
    ap.add_argument('--timeout', type=int, default=120)
    args = ap.parse_args()
    probe = {'apiVersion': 'tekton.dev/v1', 'kind': 'Pipeline',
             'metadata': {'generateName': 'retryctl-admission-ready-', 'namespace': 'default'},
             'spec': {'tasks': [{'name': 'probe', 'taskSpec': {
                 'steps': [{'name': 'probe', 'image': 'alpine:3.20', 'script': 'exit 0'}]}}]}}
    deadline = time.monotonic() + args.timeout
    last_error = ''
    while time.monotonic() < deadline:
        p = subprocess.run([args.kubectl, '--context', args.context, '--request-timeout=10s',
                            'create', '--dry-run=server', '-f', '-', '-o', 'name'],
                           input=json.dumps(probe), text=True, capture_output=True)
        if p.returncode == 0:
            print('Tekton admission webhook is ready (server dry-run).')
            return
        last_error = p.stderr
        time.sleep(1)
    raise SystemExit('Tekton admission webhook did not become ready: ' + last_error)


if __name__ == '__main__':
    main()
