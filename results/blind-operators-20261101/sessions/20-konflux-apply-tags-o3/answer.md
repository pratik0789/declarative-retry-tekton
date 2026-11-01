No environment note exists, so I can't tell how `workspace` is bound. Konflux PipelineRuns normally use a volumeClaimTemplate, and per operator-semantics.md that volume starts empty in the recovery run.
Tasks that never ran still need files in that workspace: the three SAST checks and push-dockerfile need the source tree, and build-source-image, if enabled, needs the prefetched dependencies. So clone-repository and prefetch-dependencies must run again to refill it.
Don't rerun build-container or build-image-index. The image is already pushed, and their copied IMAGE_URL/IMAGE_DIGEST results are what apply-tags and the scans should use. A rebuild could produce a new digest. init only returns a string result, so reusing it is safe.
Caveat: if `revision` is a branch rather than a commit SHA, the re-clone could fetch newer code than the image was built from.

RERUN: [clone-repository, prefetch-dependencies]
