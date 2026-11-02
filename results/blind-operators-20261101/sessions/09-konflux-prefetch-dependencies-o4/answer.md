No environment note exists, so I don't know how the `workspace` is bound. Konflux PipelineRuns normally use a `volumeClaimTemplate`, which would mean the recovery run starts with an empty workspace. In that case, reusing `clone-repository` would leave `prefetch-dependencies` and the build with no source, while still passing on the copied `url`/`commit` results.
Even if the workspace is a PVC, the failed prefetch may have half-written files into that source tree. Re-cloning is cheap and gives a clean checkout of the same revision.
`init` only produces string results (cache-proxy settings) and uses no workspace, so reusing it is safe.
RERUN: [clone-repository]
