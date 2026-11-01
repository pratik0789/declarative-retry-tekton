No environment note exists, so I don't know how the `workspace` is bound. I assumed Konflux's usual `volumeClaimTemplate`, which means a recovery run starts with an empty workspace.

Tasks that never ran and read the cloned source (`sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile`) would then find no files, so `clone-repository` must run again to restore them; this only reproduces the built source if `revision` is a pinned commit.

`build-container` and `build-image-index` don't need to rerun: the image is already in the registry, and their reused `IMAGE_URL`/`IMAGE_DIGEST` results are all that `apply-tags` and the scans need. `init` only produces results, and the prefetched dependencies are something none of the tasks still to run depend on.

If the workspace is actually a persistentVolumeClaim, the files survive and `RERUN: []` is enough.

RERUN: [clone-repository]
