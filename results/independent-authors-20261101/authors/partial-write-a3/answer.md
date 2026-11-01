Rationale:
- `prepare` deletes `artifact.txt`, `artifact-input.sha` and `receipt.txt`. If an operator reruns `prepare` while `mutate` is reused, those files stay deleted on the PVC and `publish` fails. `prepare` can't name the later `mutate` in retryWith, so the fix has to come from the downstream tasks.
- So `verify` and `publish` each name `mutate`. When either one fails, `mutate` rebuilds the artifacts from whatever `source.txt` holds at that point, which keeps them consistent even if `prepare` was rerun.
- If `mutate` fails partway, rerunning it alone is safe. `source.txt` is still clean, because only `prepare` writes it, and `mutate` fully overwrites both of its outputs.
- `publish` overwrites `receipt.txt` on rerun, and `observe` touches no state. So neither needs anything else, and `blocksResume` isn't needed anywhere.

```yaml
apiVersion: retry.tekton.dev/v1alpha1
kind: PipelineRetryPolicy
metadata: {name: pw-1-retry}
spec:
  pipelineRef: pw-1
  tasks:
    verify:
      retryWith: [mutate]
    publish:
      retryWith: [mutate]
```
