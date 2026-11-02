No environment note exists, so I can't tell whether the `workspace` binding keeps its files. Konflux normally uses a volumeClaimTemplate, so I'm assuming it starts empty in the recovery run.

`build-image-index` only needs `build-container`'s IMAGE_URL/IMAGE_DIGEST. Those are string results that get copied over, and the image is already in the registry, so neither `build-container` nor `init` has to run again.

Several downstream tasks read the source from the workspace: `push-dockerfile`, the three `sast-*` checks, and `source-build`, which also needs the prefetched dependencies. With an empty workspace they'd fail or scan nothing, so `clone-repository` and `prefetch-dependencies` must run again. Konflux normally passes a commit SHA as `revision`, so the re-clone gets the same code the image was built from. If the workspace turns out to be a persistentVolumeClaim, RERUN: [] would be enough.

RERUN: [clone-repository, prefetch-dependencies]
