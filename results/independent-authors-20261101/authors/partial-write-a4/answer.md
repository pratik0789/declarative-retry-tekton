Rationale:
- `prepare` overwrites `source.txt` and deletes `artifact.txt`, `artifact-input.sha` and `receipt.txt` on the shared PVC. If an operator reruns `prepare` while `mutate` is reused, `publish` breaks. A task can't name a later task in `retryWith`, so every task after `mutate` that touches the workspace (`verify`, `publish`) pulls `mutate` back in.
- `mutate` can safely rerun alone. It only reads `source.txt`, which only `prepare` writes, and it rewrites both of its outputs from scratch. So a partial write by `mutate` heals itself and `prepare` can be reused.
- `publish` writes its receipt last, `verify` only echoes, and `observe` and `prepare` are covered by "continue" for the tasks after them. No `blocksResume` is needed.
- The cost is that a failure in `verify` or `publish` also reruns `mutate`. That's the price of staying correct when an operator selects `prepare`.

```yaml
apiVersion: retry.tekton.dev/v1alpha1
kind: PipelineRetryPolicy
metadata:
  name: pw-1-retry
spec:
  pipelineRef: pw-1
  tasks:
    verify:
      retryWith: [mutate]
    publish:
      retryWith: [mutate]
```
