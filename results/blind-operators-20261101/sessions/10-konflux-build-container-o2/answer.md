No environment note exists, so the run doesn't say how `workspace` is bound. Konflux PipelineRuns normally bind it with a volumeClaimTemplate, which starts empty in the recovery run, so the cloned source and prefetched dependencies that `build-container` needs would be gone.
`clone-repository` and `prefetch-dependencies` must therefore run again. Even if `workspace` were a persistentVolumeClaim, rerunning them is the only safe choice when the binding isn't known.
`init` only produces string results and writes nothing to a workspace, so it can be reused safely.
Rerunning the clone also keeps `CHAINS-GIT_COMMIT` matching the source that actually gets built.

RERUN: [clone-repository, prefetch-dependencies]
