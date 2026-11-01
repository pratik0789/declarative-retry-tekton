No environment note exists, so I'm assuming the usual Konflux binding: `workspace` comes from a volumeClaimTemplate. A reused task's files would then be gone, and the recovery run would start with an empty source volume.

Several tasks that never ran read that workspace for the cloned source: `sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile`, and `build-source-image` if enabled. So `clone-repository` must run again, and `prefetch-dependencies` too, because its prefetched output lives in the same volume (`build-source-image` would need it).

The image is already in the registry, and `build-container` and `build-image-index` pass it on only as string results (IMAGE_URL/DIGEST), which get copied over. `init` produces only string results. All three can be reused, assuming `revision` is a fixed commit so the re-clone gets the same source.

RERUN: [clone-repository, prefetch-dependencies]
