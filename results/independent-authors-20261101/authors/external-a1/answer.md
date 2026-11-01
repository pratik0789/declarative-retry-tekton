Every service call checks that the records from the previous step belong to the current epoch, and the epoch can move forward between the failed run and the recovery run. So a task can't be rerun on its own if the tasks before it are reused: their records may be from an old epoch, and the call would get a 409.
- Each service task therefore names the service task just before it (`deploy`→`scan`→`sign`→`publish`→`acquire`). Because `retryWith` is followed transitively, rerunning any of them reruns everything back to `acquire`, which gets a credential for the current epoch.
- Tasks after the failed one never ran, so they run normally anyway.
- `warmup` only echoes and stores nothing in the service, so it is never forced to rerun and is always reused.
- Every entry names an earlier task, so the policy is valid. `blocksResume` isn't needed, because rerunning back from `acquire` is always a correct recovery, including after a 503.

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
