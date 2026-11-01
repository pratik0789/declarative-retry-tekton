No environment note is present, so I'm assuming the standard Konflux setup: the `workspace` is bound through a volumeClaimTemplate. That means the recovery run starts with an empty workspace, but the reused tasks' string results are still copied over.

`push-dockerfile` reads the Dockerfile from the cloned source in `workspace`, and the SAST checks and `build-source-image` read from it too. So `clone-repository` has to repopulate it. `prefetch-dependencies` also has to run again, because its prefetched output in `workspace` is used by `build-source-image`.

`init`, `build-container` and `build-image-index` don't need to run again. The image is already in the registry, and their copied results (IMAGE_URL/IMAGE_DIGEST) are still valid; rebuilding would only create a different digest.

RERUN: [clone-repository, prefetch-dependencies]
