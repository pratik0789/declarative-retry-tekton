No environment note is present, so I assumed standard Konflux: `workspace` comes from a `volumeClaimTemplate`, which starts empty in the recovery run, and `revision` is a pinned commit SHA, so re-cloning gives the same commit.

**Rationale:**
1. Any task that reads the shared `workspace` must rerun the steps that filled it: `clone-repository`, then `prefetch-dependencies`. `build-container`, `build-source-image` and the three SAST tasks need both, because the prefetched deps sit in the source tree they build or scan. `push-dockerfile` needs only the clone.
2. The image-only tasks (`build-image-index`, clair, clamav, deprecated-image, preflight, rpms-signature, `apply-tags`) take image refs from copied results, and the image is still in the registry. They have no `retryWith` and get full reuse.
3. `retryWith` can only name earlier tasks. So if an operator selects an already-successful `build-container` or `build-image-index`, the reused tasks after it keep the old digest. This policy can't prevent that, so don't select those tasks without the ones after them.
4. If `revision` can be a moving branch, re-cloning could fetch a different commit. Every workspace task would then also need the whole build and scan set in its `retryWith`, which means almost no reuse.
5. No task needs `blocksResume`.

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
      retryWith: [prefetch-dependencies]          # -> clone-repository (transitive)
    build-source-image:
      retryWith: [prefetch-dependencies, clone-repository]
    sast-snyk-check:
      retryWith: [prefetch-dependencies, clone-repository]
    sast-shell-check:
      retryWith: [prefetch-dependencies, clone-repository]
    sast-unicode-check:
      retryWith: [prefetch-dependencies, clone-repository]
    push-dockerfile:
      retryWith: [clone-repository]
```
