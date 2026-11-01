Rationale:
- `prepare` writes a new `source.txt` every time it runs, because it embeds the taskRun name. `publish` fails with STALE-ARTIFACT unless `build` rehashes it in the same run. `prepare` can't name the later `build` task in `retryWith`, so the dependency goes the other way.
- The PVC keeps old files, so if an operator selects `prepare` while `gate` or `publish` is being rerun or continued, a reused `build` leaves a stale artifact. That's why `gate` and `publish` each carry `retryWith: [build]`.
- Rerunning `build` alone is safe: it rehashes whatever `source.txt` is on the PVC. So `build` doesn't pull in `prepare`, and `observe` and `prepare` are reused unless they fail or are selected.
- If `observe`, `prepare` or `build` fails, everything after it never ran and just continues, so no `blocksResume` is needed. Every task is safe to rerun after a partial write.
- Cost: a failure in `gate` or `publish` reruns the cheap `build` task even when nothing was selected. That's the price of covering an operator selecting `prepare`.

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
