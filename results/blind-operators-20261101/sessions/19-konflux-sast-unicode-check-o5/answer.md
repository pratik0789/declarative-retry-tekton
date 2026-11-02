No environment note is present, so I don't know how the `workspace` workspace is bound. Konflux normally uses a `volumeClaimTemplate`, which means the recovery run starts with an empty workspace.

`sast-unicode-check` and the other tasks that never ran (`sast-snyk-check`, `sast-shell-check`, `push-dockerfile`, `build-source-image`) read the source files and prefetched dependencies from that workspace. So `clone-repository` and `prefetch-dependencies` must run again to fill it. This is safe if `revision` is a pinned commit SHA, which is how Konflux usually passes it.

`init`, `build-container` and `build-image-index` only pass on string results (image URL and digest, which get copied over), so they can be reused. If the workspace turns out to be a PVC, nothing needs to rerun.

RERUN: [clone-repository, prefetch-dependencies]
