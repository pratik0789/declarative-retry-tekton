Rationale:
- The chain is clone → build → retag → promote. Every failure leaves promote either failed or not yet run, so a later task is never reused on top of an earlier task that reran.
- retag and promote copy `application@<digest>` using build's reused digest. Tags are mutable and the retention job can delete that manifest, so rerunning either copy also reruns build (`retag: [build]`).
- build pulls mutable base tags, so a rebuild can produce a new digest. promote therefore reruns retag too (`[retag]` → build, transitively), which keeps `release:latest` and delivery on the same digest.
- clone and build need no entries. The source workspace is a PVC, so clone's files are still there, and clone is pinned to a fixed revision. `copy-input` (emptyDir) is not used in single-image mode.
- No `blocksResume` is needed: a promote failure from the delivery service being briefly down recovers cleanly by rebuilding and recopying.

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
      retryWith: [retag]
```
