No environment note exists, so the workspace binding is unknown. Konflux PipelineRuns usually bind `workspace` with a volumeClaimTemplate, which starts empty in the recovery run. In that case `prefetch-dependencies` would have no source tree and the build would fail or come out wrong.

Rerunning `clone-repository` fixes that. It also replaces any half-written prefetch output if the workspace is a PVC, and it keeps the cloned files in step with its `url`/`commit` results, which the pipeline reports as CHAINS-GIT_*.

`init` doesn't touch any workspace and only produces string results (proxy settings), so copying them into the recovery run is safe.

RERUN: [clone-repository]
