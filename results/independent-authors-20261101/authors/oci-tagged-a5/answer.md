Rationale:
- `promote` reads the mutable `release:latest` tag, and other pipelines may move it, so rerunning `promote` alone could ship someone else's image. `promote` therefore reruns with `retag`.
- `retag` copies `application@<digest>` using a digest copied from the failed run. The `application:build` tag can move, and the retention job may then delete that old manifest. So a rerun of `retag` also rebuilds, and `build` uses `--reproducible`.
- `build` can reuse `clone`: "source" is a PVC that still holds the checkout of the pinned commit, and kaniko doesn't change it. A rerun of `clone` re-fetches the same tree. `build` and `retag` both start from the same `clone` checkout, so neither needs `clone` rerun with it.
- `retag` and `promote` take their source and destination from params, so the empty `copy-input` volume doesn't matter. No failure requires `blocksResume`.

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
