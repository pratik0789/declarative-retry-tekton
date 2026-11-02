Rationale:
1. The service only accepts records from the current epoch, and the epoch can advance between the failed run and the recovery run. Any rerun of publish, sign, scan or deploy must therefore rebuild the whole chain from a fresh `acquire`, or it gets a 409 again.
2. Each of these tasks names its predecessor in `retryWith`. Because the closure is transitive, rerunning any of them reruns everything back to `acquire`.
3. Every `retryWith` points backwards in `runAfter` order, so the policy is valid.
4. `warmup` only runs `echo` and has no state on the service, so it is always reused. Rerunning `acquire` on its own is harmless: within one epoch it doesn't invalidate later records, and later tasks never reuse records across an epoch change.
5. `blocksResume` isn't needed, because every failure can be recovered through the closure.

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
      retryWith: [publish, acquire]
    scan:
      retryWith: [sign, publish, acquire]
    deploy:
      retryWith: [scan, sign, publish, acquire]
```
