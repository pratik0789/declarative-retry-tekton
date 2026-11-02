No environment note was provided. Konflux PipelineRuns normally bind `workspace` with a `volumeClaimTemplate`, so I'm assuming the recovery run starts with an empty source workspace.
Several tasks that never ran read that workspace: `sast-snyk-check`, `sast-shell-check`, `sast-unicode-check` and `push-dockerfile` need the cloned source, and `build-source-image` (when enabled) also needs what `prefetch-dependencies` wrote there. So `clone-repository` and `prefetch-dependencies` must run again; with a pinned `revision` they recreate the same tree.
`init`, `build-container` and `build-image-index` can be reused: the image is already in the registry, and their string results (`IMAGE_URL`/`IMAGE_DIGEST`) are copied into the recovery run.

RERUN: [clone-repository, prefetch-dependencies]
