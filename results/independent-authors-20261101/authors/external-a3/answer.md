Every service record only counts for the current epoch, and operators can advance the epoch between the failed run and the recovery run. So a task that is rerun alone may get a 409 because the records it depends on are now stale.
To be safe, rerunning any service task also reruns the whole chain before it, back to `acquire`. That way acquire, publish, sign and scan records are all rewritten in the new epoch before the rerun task uses them.
Each task names only the task directly before it; the transitive closure covers the rest, and only earlier tasks are named, so the policy is valid. The Pipeline is one linear chain, so any task an operator selects is already inside the failed task's closure.
`warmup` makes no service calls and holds no epoch state, so it is never named and its work is always reused.
`blocksResume` isn't needed, because the closure always rebuilds a consistent chain from `acquire` onward.

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
