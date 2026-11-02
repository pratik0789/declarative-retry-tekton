Order is clone → build → retag → promote. The source PVC keeps its files, so a rerun of build can use the reused clone checkout, and clone needs no entries.

- **retag needs build:** retag copies `application@<IMAGE_DIGEST>`, and that digest comes from build's reused result. The retention job may have deleted that manifest, so build must rerun to push a live digest.
- **promote needs retag:** promote reads the mutable `release:latest` tag, which other pipelines may have moved since the failed run. retag must rerun so promote copies our image; through retag this also reruns build.
- Rules stop us naming later tasks, so an operator-selected build can't force retag to rerun. `--reproducible` keeps the digest the same, which makes that acceptable. No `blocksResume` is needed.

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
