No environment note exists, so I can't tell how the shared `workspace` is bound. Konflux normally binds it with a `volumeClaimTemplate`, and if so the recovery run starts with an empty volume.

`sast-shell-check` scans the source in that workspace, and so do the not-yet-run `sast-snyk-check`, `sast-unicode-check`, `push-dockerfile` and `build-source-image`. Rerunning `clone-repository` puts the source back, and rerunning `prefetch-dependencies` restores the prefetched content, which `build-source-image` packages into the source image.

The other succeeded tasks are safe to reuse: `init`, `build-container` and `build-image-index` only pass on string results (the image URL and digest), and the image is already in the registry.

RERUN: [clone-repository, prefetch-dependencies]
