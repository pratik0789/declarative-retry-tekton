No environment note exists, so the PipelineRun's workspace binding is unknown. Konflux normally binds `workspace` with a volumeClaimTemplate, which means the recovery run starts with an empty volume.

`ecosystem-cert-preflight-checks` only needs the string result `build-image-index.IMAGE_URL`, which gets copied over. Several tasks that never ran do read source files from the workspace, though: the three `sast-*` checks, `push-dockerfile` (it needs the Dockerfile) and `build-source-image` (it needs the source plus the prefetched dependencies).

So `clone-repository` and `prefetch-dependencies` must run again to refill the workspace. `init`, `build-container` and `build-image-index` produce the image and its URL/digest, and those results are copied, so they can be reused. Rebuilding would only create a different digest. One risk: if `revision` is a branch rather than a fixed commit, the re-clone could pick up newer code than what was built.

RERUN: [clone-repository, prefetch-dependencies]
