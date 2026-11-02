No environment note was provided. Konflux PipelineRuns normally bind `workspace` through a volumeClaimTemplate, so the recovery run would start with an empty workspace.

`build-source-image` reads that workspace, needing the cloned source and the prefetched dependencies, and so do later tasks like the SAST checks and `push-dockerfile`. So `clone-repository` and `prefetch-dependencies` must run again to refill it.

`init`, `build-container` and `build-image-index` can be reused: the image is already in the registry, and the string results (IMAGE_URL/DIGEST, proxy settings) are copied over. This assumes `revision` is a fixed commit SHA (Konflux's default), so the re-clone matches the source that was built.

RERUN: [clone-repository, prefetch-dependencies]
