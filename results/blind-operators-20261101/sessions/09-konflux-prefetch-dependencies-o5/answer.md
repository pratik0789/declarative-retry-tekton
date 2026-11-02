No environment note exists, so how the `workspace` is bound is unknown. Konflux PipelineRuns normally use a volumeClaimTemplate, and that would start the recovery run with an empty workspace.

prefetch-dependencies and build-container need the cloned source in that workspace. So clone-repository must run again. Even with a PVC, re-cloning (git-clone deletes the existing copy by default) clears out anything the failed prefetch half-wrote.

init writes nothing to a workspace and only produces string results (e.g. cache-proxy settings), which get copied over, so it can be reused safely.

RERUN: [clone-repository]
