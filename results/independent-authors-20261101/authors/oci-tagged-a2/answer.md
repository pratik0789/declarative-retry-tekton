Rationale:
1. `clone` → `build` → `retag` → `promote` is a single chain. `source` is a PVC, so a rerun `build` still sees the pinned-revision checkout and doesn't need `clone`. A rerun `clone` re-fetches the same commit, so the `build` digest it would reuse is still valid.
2. `retag` copies `application@<IMAGE_DIGEST>`. If that digest is reused from the old run, the retention job may already have deleted its manifest. So `retag` must rerun `build` to get a digest that exists now.
3. `promote` reads the mutable tag `release:latest`, and other pipelines may have moved it since. So `promote` must rerun `retag` (and therefore `build`) so the tag points at our image again. When the delivery service is briefly down, rerunning that chain is safe, so no `blocksResume`.
4. `copy-input` (emptyDir) is never read, because both skopeo tasks receive their image URLs as params. Every `retryWith` entry names an earlier task, as the rules require.

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
