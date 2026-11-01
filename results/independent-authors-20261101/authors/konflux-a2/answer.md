I assumed the Konflux default `volumeClaimTemplate` workspace, so the recovery run starts with an empty workspace. Any rerun task that reads files must therefore rerun the tasks that put them there: `clone-repository`, plus `prefetch-dependencies`, which adds deps and can modify the source.
Tasks that only use results (`init`, `build-image-index`, image scans, `apply-tags`) get no entries. Copied results and the image already pushed to the registry are enough for them.
The builds and SAST scans need the full source plus deps (through `prefetch-dependencies`). `push-dockerfile` needs only the Dockerfile, so `clone-repository` is enough for it.
These rules can't force reruns of later tasks. So if an operator selects `clone-repository` or `build-container` alone, later tasks keep their old results. Correctness also assumes `revision` is a fixed commit SHA, so a re-clone can't pick up a newer commit.

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
      retryWith: [prefetch-dependencies]
    sast-shell-check:
      retryWith: [prefetch-dependencies]
    sast-unicode-check:
      retryWith: [prefetch-dependencies]
    push-dockerfile:
      retryWith: [clone-repository]
```
