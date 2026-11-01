Rationale:
- `promote` copies the mutable `release:latest` tag, and other pipelines may move it. A lone rerun could ship someone else's image, so `promote` reruns `retag` to re-point the tag at our digest first.
- `retag` reads `build`'s `IMAGE_DIGEST` result. The retention job may delete that manifest, so a reused digest could be dangling. `retag` therefore reruns `build` to push a fresh image and digest.
- `build` reuses `clone`'s files because `source` is a PVC (pinned revision, no result references). `copy-input` (emptyDir) is unused because the copy URLs are passed as params.
- No task's failure leaves state a rerun can't overwrite. Promote outages are transient. So no task sets `blocksResume`.

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
