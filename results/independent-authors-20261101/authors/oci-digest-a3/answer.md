Rationale:
- `retag` and `promote` copy the image by `build`'s IMAGE_DIGEST result. A retention job may have deleted that manifest, so a reused digest isn't safe. Rerunning either task also reruns `build`.
- `build` pulls its base image through a mutable tag (`BASE_PREFIX=registry-source:5000/base/`), so a rebuild can produce a new digest. Another pipeline may also have moved `release:latest`. So `promote` reruns `retag` too, which keeps both tags on the same image.
- `build` doesn't need `clone`. The `source` PVC keeps the cloned files, the clone is pinned to a commit SHA, kaniko doesn't write into the build context, and no task reads `clone`'s results. `clone` is always reused unless it failed or an operator selects it.
- With this policy: if `promote` fails, the recovery run reruns `build`, `retag` and `promote`. If `retag` fails, it reruns `build` and `retag`. If `build` or `clone` fails, it reruns only that task. Operator selections pull in the same closures.
- No `blocksResume` is needed. A failed `clone` empties its directory before cloning again, and the `copy-input` emptyDir isn't used because both copy tasks get their image URLs as params.

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
      retryWith: [retag, build]
```
