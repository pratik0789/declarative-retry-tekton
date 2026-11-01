No environment note exists, so I assumed the Konflux default: the `workspace` workspace is a volumeClaimTemplate, which means the recovery run starts with it empty.

`build-source-image` reads that workspace, and so do the not-yet-run `sast-*` and `push-dockerfile` tasks. So `clone-repository` and `prefetch-dependencies` must run again to refill it with the source and the prefetched dependencies.

`init`, `build-container` and `build-image-index` can be reused: the copied `IMAGE_URL` and `IMAGE_DIGEST` still point to the image already pushed to the registry. That holds only if `revision` is pinned to a commit SHA, as Konflux usually does. If it names a moving branch, the re-clone could fetch newer code than what was built, and `build-container` and `build-image-index` would need to run again too.

RERUN: [clone-repository, prefetch-dependencies]
