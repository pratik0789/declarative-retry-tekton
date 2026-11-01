No environment note was present, so I assumed the usual Konflux setup, where `workspace` is a `volumeClaimTemplate`. Under that assumption the recovery run starts with an empty source volume.
Several tasks that never ran read that workspace: `sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile`, and `build-source-image` (which needs the prefetched dependencies). So the source must be cloned and prefetched again. In Konflux, `revision` is a fixed commit SHA, so the re-clone matches what was built.
`clair-scan` and the other image checks only use the image URL and digest from `build-image-index`. Those results are copied over, so `build-container`, `build-image-index` and `init` can be reused.
If the workspace is actually a persistentVolumeClaim, the files are still there and RERUN: [] would be enough.
RERUN: [clone-repository, prefetch-dependencies]
