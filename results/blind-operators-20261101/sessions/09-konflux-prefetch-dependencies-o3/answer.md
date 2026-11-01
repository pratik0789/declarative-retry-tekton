No environment note exists, so the `workspace` binding is unknown. Konflux PipelineRuns normally give `workspace` a fresh volume on each run (`volumeClaimTemplate`), which means the recovery run's workspace would start empty.

`clone-repository` writes the source into that workspace, and `prefetch-dependencies` and every later task read it from there. So `clone-repository` has to run again. Rerunning it also avoids the same-commit result (`CHAINS-GIT_COMMIT`) being reported for files that aren't actually on disk.

`init` only produces string results from a parameter and writes nothing to the workspace, so its copied results are still valid and it can be reused.

RERUN: [clone-repository]
