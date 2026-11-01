Pipeline order is `observe → prepare → mutate → verify → publish`. Nothing earlier can name `verify` or `publish`, because `retryWith` only points backward.

Rationale:
- **Reruns that are safe alone:** `prepare`, `mutate` and `publish` overwrite or recreate everything they write. So a rerun of any of them repairs its own partial writes, and `mutate` never changes `source.txt`.
- **The risk:** if an operator reruns `prepare`, it deletes `artifact.txt` and `artifact-input.sha`, and a reused `mutate` would not rebuild them. `publish` would then fail on the missing files.
- **The fix:** whenever `verify` or `publish` is rerun (failed or selected), `mutate` reruns too and rebuilds the artifact from the current `source.txt`.
- **Reuse kept:** `mutate` doesn't need `prepare` because `source.txt` stays clean on the persistent volume. `observe` writes nothing, so it stays reused, and no task needs `blocksResume`.

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
