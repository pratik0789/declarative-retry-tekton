Every task from `publish` to `deploy` lists the task right before it in `retryWith`. Followed transitively, a rerun of any of them also reruns everything back to `acquire`.
Why: the epoch can advance between the failed run and the recovery run, and the service then rejects records from the old epoch. Rerunning the chain from `acquire` rebuilds all records in the current epoch, whether the failure was a 409 or a 503.
Each entry names only the earlier task, as the rules require. Tasks after the failed one never ran, so they run normally anyway.
`warmup` only echoes and the service doesn't depend on it, so it is always reused. `acquire` needs no entries.
`blocksResume` isn't needed, because the rerun group always rebuilds a consistent set of records.

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
