The pipeline runs in a straight line: observe → prepare → build → gate → publish. Only build can make the workspace stale: it records a hash of `source.txt`, and publish checks that hash.

- Rules only let a task name earlier tasks, so "if prepare reruns, also rerun build" can't be written. The fix is to have the tasks after build pull build back in instead.
- If an operator reruns prepare to get fresh input, build has already succeeded, so the failure must be in build, gate or publish. A build failure reruns build anyway. Gate and publish each get `retryWith: [build]`, so build always recomputes from the current `source.txt`.
- prepare is safe to reuse: the volume keeps `source.txt`, and rerunning build from it is idempotent. observe and gate have no side effects. publish writes `receipt.txt` only after its checks pass. A half-finished run of any task is overwritten when it reruns, so `blocksResume` isn't needed.

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
