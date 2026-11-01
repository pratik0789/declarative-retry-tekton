No environment note was present, so I assumed Konflux's usual `volumeClaimTemplate` workspace. That means the recovery run starts with an empty `workspace`.

These tasks that never ran read source files from that workspace: `sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile` (needs the Dockerfile) and `build-source-image` (needs the source and prefetched dependencies). So `clone-repository` and `prefetch-dependencies` must run again to refill it.

`init`, `build-container` and `build-image-index` only pass on string results (the image URL and digest), and the image is already in the registry, so they can be reused. If `revision` is a branch rather than a commit SHA, it could move between runs and the new clone would no longer match the built image; this answer assumes it is pinned.

RERUN: [clone-repository, prefetch-dependencies]
