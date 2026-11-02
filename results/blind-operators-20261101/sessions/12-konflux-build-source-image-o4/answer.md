No environment note was present, so I couldn't confirm how the `workspace` is bound. I assumed the Konflux default, a `volumeClaimTemplate`, which starts empty in the recovery run.

The failed `build-source-image` reads the cloned source and the prefetched dependencies from that workspace. So `clone-repository` and `prefetch-dependencies` have to run again to refill it. The SAST checks and `push-dockerfile` also need the source and haven't run yet.

`init`, `build-container` and `build-image-index` can be reused, because the next tasks only use their string results (proxy settings, `IMAGE_URL`/`IMAGE_DIGEST`). That only holds if `revision` is a fixed commit SHA: a branch that has moved would give a source image that doesn't match the binary.

RERUN: [clone-repository, prefetch-dependencies]
