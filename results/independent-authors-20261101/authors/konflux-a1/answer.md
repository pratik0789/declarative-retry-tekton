Rationale:
1. I assumed the `workspace` binding is a volumeClaimTemplate (the Konflux default), so a recovery run starts with an empty workspace. Every rerun task that reads it has to rerun the tasks that fill it: `clone-repository` for source, then `prefetch-dependencies` for deps.
2. `build-container` and `build-source-image` need both the source and the prefetched deps. `push-dockerfile` and the three `sast-*` tasks only read source, so they rerun `clone-repository` alone.
3. Tasks that only use the image (`build-image-index`, clair/clamav/preflight/deprecated/rpms scans, `apply-tags`) get their inputs from copied string results, and the image stays in the registry, so they need no `retryWith`. `init` has no consumers.
4. This assumes `revision` is a fixed commit SHA (Pipelines-as-Code sets it that way). If it were a branch, re-cloning could fetch a newer commit than the one that was built.
5. `retryWith` can only name earlier tasks. So when an operator selects a successful build task, the policy can't force the later tasks to rerun, and they keep stale results. `blocksResume` doesn't help here because it only applies to the task that failed.

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
      retryWith: [prefetch-dependencies, clone-repository]
    build-source-image:
      retryWith: [prefetch-dependencies, clone-repository]
    push-dockerfile:
      retryWith: [clone-repository]
    sast-snyk-check:
      retryWith: [clone-repository]
    sast-shell-check:
      retryWith: [clone-repository]
    sast-unicode-check:
      retryWith: [clone-repository]
```
