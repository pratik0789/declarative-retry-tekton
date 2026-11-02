Any 409 can mean the oracle's epoch moved on, and once that happens every stored record (credential, artifact, signature, scan) is stale. So rerunning a failed task means rerunning everything from `acquire` up to it.
Each task's `retryWith` names only the task right before it (always an earlier task, so the policy is valid). Because `retryWith` is followed transitively, a failure at step N reruns `acquire` through N.
An operator-selected success is always upstream of the failure or covered by its closure, so no reused task depends on a rerun one. Tasks after the failure never ran, so they continue normally.
`warmup` has no external state and is never in a closure, so it is always reused. No `blocksResume` is needed: rerunning the chain always works, and a 503 is fixed by rerunning too.

```yaml
apiVersion: retry.tekton.dev/v1alpha1
kind: PipelineRetryPolicy
metadata:
  name: external-recovery-retry
spec:
  pipelineRef: external-recovery
  tasks:
    publish:
      retryWith: [acquire]
    sign:
      retryWith: [publish]
    scan:
      retryWith: [sign]
    deploy:
      retryWith: [scan]
```
