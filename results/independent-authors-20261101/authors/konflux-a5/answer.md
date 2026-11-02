No environment note is present, so I assumed the `workspace` binding is a volumeClaimTemplate (the Konflux default). That means it starts empty in a recovery run, and any task that reads it needs `clone-repository` rerun, plus `prefetch-dependencies` if it uses the prefetched dependencies. Tasks that only read result values (`build-image-index` and the image scans) can reuse copied results, which still point to the same image digest.
I also assumed the `revision` parameter is a fixed commit SHA, as Konflux's Pipelines-as-Code sets it, so a re-clone checks out the same commit. If it is a branch, a re-clone could fetch a newer commit and give the wrong `CHAINS-GIT_COMMIT`. The policy can't stop that if an operator selects an upstream task, because `retryWith` may not name later tasks.
`init` produces no results that other tasks use, so it is always reused. No task needs `blocksResume`.

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
