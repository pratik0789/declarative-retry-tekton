Five tasks that haven't run yet mount `workspace` and need the cloned source: `sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile` and `build-source-image` (when enabled). There's no environment note, so I'm assuming Konflux's usual `volumeClaimTemplate` binding, which means the recovery run starts with an empty workspace.
So `clone-repository` must run again to put the source back. `init`, `build-container` and `build-image-index` only pass on string results (the image URL and digest), which are copied over, so they don't need to run.
`prefetch-dependencies`' files were only used by `build-container`, which is already done (and by `build-source-image`, which is off by default). If the workspace is actually a persistentVolumeClaim, `RERUN: []` is enough.

RERUN: [clone-repository]
