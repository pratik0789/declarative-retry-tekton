No environment note exists, so I'm assuming the Konflux default: the `workspace` workspace is bound through a volumeClaimTemplate and starts empty in the recovery run.
`push-dockerfile`, the three `sast-*` checks and (if enabled) `build-source-image` all read source files and prefetched dependencies from that workspace. So `clone-repository` and `prefetch-dependencies` must run again to refill it.
`build-container` only hands on string results (IMAGE_URL/IMAGE_DIGEST), which get copied, and the image is already in the registry. `init` only produces strings too, so both can be reused. This assumes `revision` is a fixed commit SHA, so the re-clone matches the built image.

RERUN: [clone-repository, prefetch-dependencies]
