No environment note exists, so I can't confirm how `workspace` is bound. In Konflux it is normally a `volumeClaimTemplate`, which means the recovery run starts with an empty volume.

Several tasks that never ran read files from `workspace`: `sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile`, and `build-source-image` if enabled. They need the cloned source and the prefetched dependencies. So `clone-repository` and `prefetch-dependencies` must run again to refill the volume. `revision` is the same pinned value, so the source will match.

`init`, `build-container` and `build-image-index` only pass on string results (image URL and digest), which are copied over, so they can be reused.

RERUN: [clone-repository, prefetch-dependencies]
