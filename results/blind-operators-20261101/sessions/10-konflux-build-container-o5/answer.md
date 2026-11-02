No environment note exists, so I can't confirm the `workspace` binding is a persistentVolumeClaim. Konflux PipelineRuns normally bind it with a volumeClaimTemplate, and a recovery run would mount that empty.
`build-container` reads the source tree and prefetched dependencies from `workspace`. An empty volume means the clone and prefetch outputs are missing, so both of those tasks must run again.
`prefetch-dependencies` works on the files `clone-repository` checks out, so rerunning the prefetch alone would not be enough.
`init` doesn't mount any workspace, and its results only need to be copied, so it can be reused.

RERUN: [clone-repository, prefetch-dependencies]
