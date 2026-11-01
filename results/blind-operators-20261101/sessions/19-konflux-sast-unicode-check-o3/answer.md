No environment note exists, so I can't confirm how `workspace` is bound. Konflux normally uses a `volumeClaimTemplate`, which starts empty in the recovery run.

That matters because `sast-unicode-check` scans the source in `workspace`, as do `sast-snyk-check`, `sast-shell-check`, `push-dockerfile` and `build-source-image`. `clone-repository` and `prefetch-dependencies` must run again to put that source back (assuming `revision` is pinned to the same commit). `build-container`, `build-image-index` and `init` can be reused: the image is already in the registry and the later tasks only need their copied IMAGE_URL/IMAGE_DIGEST results.

RERUN: [clone-repository, prefetch-dependencies]
