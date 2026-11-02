Nothing here says how the `workspace` is bound, and there's no environment note. Konflux PipelineRuns use a `volumeClaimTemplate` by default, so I'm assuming the recovery run starts with an empty workspace.
`init`, `build-container` and `build-image-index` only pass string results (image URL/digest, cache-proxy settings) to later tasks, and the built image is already in the registry. Reusing them is safe.
Several tasks that never ran read the source from the workspace: `sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile` and possibly `build-source-image`. So `clone-repository` must run again, and I'm also rerunning `prefetch-dependencies` so the workspace matches what the first run had. This assumes `revision` points to the same commit that was built. If the workspace is actually a `persistentVolumeClaim`, the answer would be RERUN: [].

RERUN: [clone-repository, prefetch-dependencies]
