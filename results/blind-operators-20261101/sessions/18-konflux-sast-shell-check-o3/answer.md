Re-run `clone-repository` and `prefetch-dependencies`; reuse the rest.

1. There's no environment note, and Konflux PipelineRuns usually bind `workspace` with a `volumeClaimTemplate`. That volume starts empty in the recovery run, so `clone-repository` must refill it.
2. The tasks that still have to run read files from `workspace`: `sast-shell-check`, `sast-snyk-check`, `sast-unicode-check`, `push-dockerfile`, and `build-source-image` if enabled. `build-source-image` also needs the prefetched dependencies, so `prefetch-dependencies` runs again too.
3. `build-container` and `build-image-index` are reused, so the image stays the same and every check scans it. `init`'s results aren't needed by anything that runs again. If `revision` is a branch rather than a commit SHA, check that the clone lands on the commit that was built.

RERUN: [clone-repository, prefetch-dependencies]
