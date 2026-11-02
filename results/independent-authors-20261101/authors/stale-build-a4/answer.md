Rationale:
- The constraint: `prepare` writes a new `source.txt` on every run (it embeds the TaskRun name), and `build` hashes that file. If `prepare` reruns and `build` is reused, `publish` fails with STALE-ARTIFACT.
- The rules don't let `prepare` name `build` in `retryWith`, because `build` comes later. So every task after `build` that can be the failed task (`gate`, `publish`) must pull `build` in instead.
- An operator can only select `prepare` alongside a failed `gate` or `publish`. With this policy, that recovery always rebuilds from the current `source.txt` on the PVC, so the artifact matches the source.
- If `prepare` itself fails, or `build` fails, nothing after it is reused or the rerun is self-contained. `observe` and `gate` are stateless, so `blocksResume` isn't needed.
- `build` always rewrites both of its files, so rerunning it alone against the persisted `source.txt` is safe. That means `prepare` is never forced to rerun, and `observe` and `prepare` are always reused unless selected.

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
