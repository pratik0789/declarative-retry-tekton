Rationale:
- **retag → build**: a reused build only hands over its digest string, and the retention job may already have deleted that manifest. Rerunning the reproducible build pushes it again.
- **promote → retag, build**: promote also needs a live digest. `BASE_PREFIX` pulls base images by mutable tag, so a rebuild can produce a new digest. Rerunning retag with build keeps `release:latest` and `delivery:latest` on the same image.
- **build has no retryWith**: clone checks out a pinned SHA into the PVC, so a rerun build sees the same source. If an operator reruns clone, it wipes and re-clones the same SHA, so a reused build stays valid.
- **No blocksResume**: each task can be retried safely, including a promote that fails because the delivery service is briefly down. The `copy-input` emptyDir is never read because both copy tasks get the URLs as params.
- Limit: retryWith can't name later tasks, so if an operator selects build and nothing after it fails, retag and promote are reused with the old digest. Select them together with build.

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
