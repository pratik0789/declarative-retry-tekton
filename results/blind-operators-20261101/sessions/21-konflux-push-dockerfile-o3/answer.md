Without an environment note I can't tell how `workspace` is bound. Konflux PipelineRuns normally use a `volumeClaimTemplate`, which would mean the recovery run starts with an empty source volume.

`push-dockerfile` reads the Dockerfile from that workspace, and so do the SAST checks and `build-source-image`, which haven't run yet. So `clone-repository` must refill it, and `prefetch-dependencies` must put its prefetched files back too.

`init`, `build-container` and `build-image-index` only pass on string results (the image URL and digest), and the image is already in the registry, so they can be reused. This assumes `revision` is a fixed commit, so re-cloning gives the same source the image was built from.

RERUN: [clone-repository, prefetch-dependencies]
