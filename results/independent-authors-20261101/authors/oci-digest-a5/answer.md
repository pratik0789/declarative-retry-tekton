I wrote a policy where reruns of `retag` or `promote` pull in `build`, and reruns of `promote` also pull in `retag`.

- **`build` is never reused under `retag` or `promote`.** Both copy `application@<IMAGE_DIGEST>`, and the retention job may delete that old manifest. So whenever they rerun, `build` reruns first to push a fresh image and digest.
- **`promote` also reruns `retag`.** Other pipelines can move `release:latest`, and a rebuild can produce a new digest. Rerunning `retag` makes `release:latest` and `delivery:latest` point at the same new image.
- **`clone` is always reused.** `source` is a persistentVolumeClaim, so the checkout of the pinned revision survives, and nothing reads `clone`'s results. `build` doesn't need it rerun. `copy-input` being empty doesn't matter, because both copy tasks take their source and destination from params.
- **No `blocksResume`.** A failed `promote` from the delivery service being briefly down is transient and safe to retry. Every case where one task fails, plus any tasks the operator selects, ends with both tags on the same digest.

```yaml
apiVersion: retry.tekton.dev/v1alpha1
kind: PipelineRetryPolicy
metadata:
  name: public-oci-digest-retry
spec:
  pipelineRef: public-oci-digest
  tasks:
    retag:
      retryWith: [build]
    promote:
      retryWith: [build, retag]
```
