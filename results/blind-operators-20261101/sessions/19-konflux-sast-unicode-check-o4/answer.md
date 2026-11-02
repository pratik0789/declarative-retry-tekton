No environment note exists, so I can't confirm how `workspace` is bound. Konflux PipelineRuns normally bind it with `volumeClaimTemplate`, so I'm assuming the recovery run starts with it empty.
Four not-yet-run tasks read source files from `workspace`: `sast-unicode-check`, `sast-snyk-check`, `sast-shell-check` and `push-dockerfile`. `clone-repository` must run again to put the source back, and `prefetch-dependencies` must run again to restore the dependencies it added there.
`init`, `build-container` and `build-image-index` only need their string results (proxy settings, IMAGE_URL/DIGEST) carried over, and the image is already in the registry. Reusing them is safe.
If `workspace` turns out to be a persistentVolumeClaim, its files are kept and RERUN: [] would be enough.
RERUN: [clone-repository, prefetch-dependencies]
