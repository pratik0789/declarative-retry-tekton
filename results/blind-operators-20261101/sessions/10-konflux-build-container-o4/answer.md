No environment note exists, so the PipelineRun's binding for `workspace` is unknown. Standard Konflux PipelineRuns bind it with a `volumeClaimTemplate`, which means the recovery run gets an empty volume.
`build-container` reads the source and prefetched dependencies from that workspace. `clone-repository` writes the source there and `prefetch-dependencies` writes the dependencies there, so both must run again to refill it, in Pipeline order.
Re-cloning also keeps the `CHAINS-GIT_COMMIT` result matched to the source that actually gets built. A reused result plus a fresh clone of a moving branch could disagree.
`init` uses no workspace and none of its results are consumed downstream, so reusing it is safe. If the binding were a `persistentVolumeClaim`, RERUN: [] would be enough.

RERUN: [clone-repository, prefetch-dependencies]
