No environment note exists in this directory, so I don't know how the `workspace` is bound. Konflux PipelineRuns normally use a volumeClaimTemplate, which the recovery run would start empty.

`push-dockerfile` reads the Dockerfile from that workspace. So do the tasks that haven't run yet: the three SAST checks, plus `build-source-image`, which also needs the prefetched dependencies. So `clone-repository` and `prefetch-dependencies` must run again to refill it.

`init`, `build-container` and `build-image-index` only hand on string results (the image URL and digest), which are copied over, so reuse them. This assumes `revision` points to the same commit that was built.

RERUN: [clone-repository, prefetch-dependencies]
