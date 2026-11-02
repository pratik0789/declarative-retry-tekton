- clone pins a fixed commit and "source" is a PVC, so the recovery run still sees the checkout. build can reuse clone, and clone can rerun on its own (it checks out the same files again).
- retag and promote copy `application@<IMAGE_DIGEST>` from a reused build result, but the retention job may already have deleted that manifest. So both must rerun build to get a fresh, existing digest.
- build pulls base images by mutable tags, so a rebuild can produce a different digest. promote therefore also reruns retag, so release:latest and delivery:latest point at the same digest. In this linear chain, anything after a rerun task is either rerun or never ran.
- copy-input (emptyDir) is only read when the URL params are empty, and here they are always set, so it has no effect. A brief delivery outage is just a rerun of promote, so no task needs blocksResume.

```yaml
apiVersion: retry.tekton.dev/v1alpha1
kind: PipelineRetryPolicy
metadata: {name: public-oci-digest-retry}
spec:
  pipelineRef: public-oci-digest
  tasks:
    retag:
      retryWith: [build]
    promote:
      retryWith: [build, retag]
```
