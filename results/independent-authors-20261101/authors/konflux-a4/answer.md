Rationale:
- There is no environment note, so I assume the shared `workspace` may be a volumeClaimTemplate, which starts empty in the recovery run. Any rerun task that reads files from it must therefore rerun the tasks that produce those files.
- `prefetch-dependencies`, `sast-*` and `push-dockerfile` need the cloned source. `build-container` and `build-source-image` also need the prefetched dependencies (cachi2 output), and `prefetch-dependencies` brings in the clone transitively.
- Image-consuming tasks (`build-image-index`, scanners, `apply-tags`) only use copied string results plus images already in the registry, so they rerun alone. `init` has no consumers. No task needs `blocksResume`.
- The rules don't allow naming later tasks, so tasks downstream of an operator-selected task get reused as-is. That is only correct if `revision` is a pinned commit SHA, as Pipelines-as-Code supplies; a branch revision would let a re-clone drift from the built image.

```yaml
apiVersion: retry.tekton.dev/v1alpha1
kind: PipelineRetryPolicy
metadata:
  name: template-build-retry
spec:
  pipelineRef: template-build
  tasks:
    prefetch-dependencies:
      retryWith: [clone-repository]
    build-container:
      retryWith: [prefetch-dependencies]
    build-source-image:
      retryWith: [prefetch-dependencies]
    sast-snyk-check:
      retryWith: [clone-repository]
    sast-shell-check:
      retryWith: [clone-repository]
    sast-unicode-check:
      retryWith: [clone-repository]
    push-dockerfile:
      retryWith: [clone-repository]
```
