Rerun clone-repository and reuse init.

- **clone-repository must rerun.** It does its work in the `workspace` workspace, and nothing tells us how that workspace is bound (there is no environment note). Konflux PipelineRuns normally bind it with a `volumeClaimTemplate`, which starts empty in the recovery run. If it is reused, prefetch-dependencies and build-container would find no source. Re-cloning is also the safer choice on a persistentVolumeClaim, because it clears anything the failed prefetch left half-written.
- **init can be reused.** It uses no workspace and only produces string results, which the tool copies into the recovery run.

RERUN: [clone-repository]
