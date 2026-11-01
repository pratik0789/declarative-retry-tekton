Rationale:
- Any rerun of `prepare` rewrites `source.txt`, using the new TaskRun name. If `build` is reused after that, `publish` sees a stale artifact and fails with STALE-ARTIFACT. So every rerun of `prepare` has to include `build`.
- `prepare` can't name `build` in `retryWith` because `build` comes later. The only way to guarantee it is for every task after `build` that can fail (`gate`, `publish`) to name `build`. The operator can select `prepare` only when one of those has failed. If `observe`, `prepare` or `build` fails, `build` either runs normally or is the failed task.
- `build` recomputes everything from the `source.txt` left on the PVC, and the operator can still select `prepare`, so `build` doesn't need to pull in `prepare`. `observe` and `gate` only echo, so they are never pulled in. No `blocksResume` is needed.

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
