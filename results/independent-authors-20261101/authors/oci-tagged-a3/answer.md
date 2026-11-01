Rationale:
- `promote` copies the mutable tag `release:latest`, and other pipelines may move it. A rerun `promote` must have `retag` rerun with it so it copies this run's image.
- `retag` copies by the reused `IMAGE_DIGEST`, which the retention job may have deleted. It must rerun `build` so the manifest is pushed again. The rule is transitive, so a `promote` rerun means `promote`, `retag` and `build` all rerun.
- `build` doesn't need `clone`: the `source` PVC keeps the cloned files and the revision is pinned. Rerunning `clone` alone is safe because later tasks only use the same pinned source.
- `blocksResume` isn't needed: each `skopeo copy` and `kaniko` push just overwrites its target, so running it again is safe. That covers a `promote` that failed because the delivery service was briefly down. `copy-input` (emptyDir) is never read, because the src and dest params are always set.

```yaml
apiVersion: retry.tekton.dev/v1alpha1
kind: PipelineRetryPolicy
metadata:
  name: public-oci-tagged-retry
spec:
  pipelineRef: public-oci-tagged
  tasks:
    retag:
      retryWith: [build]
    promote:
      retryWith: [retag]
```
