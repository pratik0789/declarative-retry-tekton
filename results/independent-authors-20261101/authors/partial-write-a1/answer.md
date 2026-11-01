**Rationale:**
- Each task overwrites its own outputs, so a task that failed partway can safely be rerun alone. On the shared PVC it just replaces its half-written files, and `source.txt` is only ever written by `prepare`.
- The risk is an operator selecting `prepare` for fresh input. `prepare` deletes `artifact.txt`, `artifact-input.sha` and `receipt.txt`. A policy can't make `prepare` pull later tasks back in, so any later failed task must bring `mutate` back with it.
- `verify` → `mutate`: if `verify` fails while `prepare` is selected, `mutate` gets rebuilt before `publish` continues. `publish` → `mutate` covers the same case when `publish` is the one that failed.
- `prepare` stays reused, since `mutate` only reads its unchanged `source.txt`. `observe` stays reused. A failure in `mutate` or `prepare` needs no entries: rerunning the failed task is enough, because the tasks after it never ran and run normally.
- No `blocksResume` is needed, because every failure point can be recovered. Both entries name `mutate`, which runs before `verify` and `publish`, so the policy is valid.

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
