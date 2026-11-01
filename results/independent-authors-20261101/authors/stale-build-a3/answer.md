Rationale:
- The pipeline is a straight chain: observe → prepare → build → gate → publish. The only stale-data risk is `prepare` rerunning, which writes a new `source.txt` because its content includes the TaskRun name, while `build` is reused. Then the old `artifact-input.sha` stays on the PVC and `publish` fails with STALE-ARTIFACT.
- `retryWith` can only name earlier tasks, so `prepare` can't force `build` to rerun. The fix has to go on the tasks after `build`.
- If observe, prepare or build fails, `build` runs anyway (rerun or continue). If gate or publish fails, `build` only reruns when that failed task pulls it in, so both name it.
- `build` rebuilds from whatever `source.txt` is on the PVC, so it doesn't need `prepare`. Partial writes are overwritten when a task reruns, and `observe` is harmless, so no `blocksResume` is needed.
- The cost: when only gate or publish fails, `build` reruns even if nobody selected `prepare`. The policy can't check what the operator selected, so this is the least rerunning that is always correct.

```yaml
apiVersion: retry.tekton.dev/v1alpha1
kind: PipelineRetryPolicy
metadata: {name: fi-1-retry}
spec:
  pipelineRef: fi-1
  tasks:
    gate:
      retryWith: [build]
    publish:
      retryWith: [build]
```
