Rationale:
- Only `prepare`, `mutate` and `publish` touch the shared PVC. `observe` and `verify` keep no state. `mutate` only reads `source.txt` and never changes it, so a rerun of `mutate` alone is safe. Its output files are overwritten from scratch each time, so a partial write is simply replaced.
- The risk is an operator selecting `prepare`, which deletes `artifact.txt` and `artifact-input.sha`. If `mutate` is then reused, the next `publish` fails. retryWith can't name a later task, so `prepare` can't pull `mutate` in. Instead, every task that leads to `publish` running without `mutate` must pull `mutate` in. Those are `verify`, which fails before `publish` continues, and `publish` itself.
- With `verify: [mutate]` and `publish: [mutate]`, each recovery reruns `mutate` and `publish`. `prepare` and `observe` stay reused unless selected, and `verify` is reused when `publish` fails.
- If `prepare` or `mutate` fails, everything after them never ran and just continues, so they need no entries. No partial write blocks recovery, so `blocksResume` isn't needed.

```yaml
apiVersion: retry.tekton.dev/v1alpha1
kind: PipelineRetryPolicy
metadata:
  name: pw-1-retry
spec:
  pipelineRef: pw-1
  tasks:
    verify:
      retryWith: [mutate]
    publish:
      retryWith: [mutate]
```
