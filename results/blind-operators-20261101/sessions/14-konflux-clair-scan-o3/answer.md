No environment note exists, so the workspace binding is unknown. Konflux PipelineRuns usually bind `workspace` with a volumeClaimTemplate, which means the recovery run starts with an empty workspace.

Several tasks that never ran read files from that workspace: sast-snyk-check, sast-shell-check, sast-unicode-check, push-dockerfile (needs the Dockerfile) and build-source-image (needs the source plus prefetched dependencies). So clone-repository and prefetch-dependencies must run again to refill it. This stays consistent only if `revision` is a fixed commit SHA, which Konflux normally sets; if it's empty or a branch, the re-clone could fetch newer code than what was built.

init, build-container and build-image-index produce only string results and a pushed image. Those results are copied into the recovery run, so clair-scan and the other checks scan the same image digest. Rebuilding would create a new digest for no benefit.

RERUN: [clone-repository, prefetch-dependencies]
