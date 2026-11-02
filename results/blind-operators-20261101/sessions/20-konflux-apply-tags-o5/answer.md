No environment note here, so I can't tell how the `workspace` is bound. Konflux normally uses a `volumeClaimTemplate`, so I'm assuming the recovery run gets an empty workspace. Tasks that never ran need its files: the three sast checks and `push-dockerfile` read the source and Dockerfile, and `build-source-image` reads the source plus the prefetched dependencies.
So `clone-repository` and `prefetch-dependencies` must run again to refill it. With the same `revision` they produce the same commit, so their results still match the image that was already pushed.
`init`, `build-container` and `build-image-index` only pass string results to later tasks (the image URL and digest), and those are copied over, so reusing them is safe and the image isn't rebuilt.

RERUN: [clone-repository, prefetch-dependencies]
