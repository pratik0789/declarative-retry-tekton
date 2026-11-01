Nothing in this directory says how the `workspace` workspace is bound; Konflux normally uses a `volumeClaimTemplate`, which would start empty in the recovery run. That would leave the `build-container` and `push-dockerfile`/SAST steps with no files, so `clone-repository` and `prefetch-dependencies` must rerun to put back the source and prefetched dependencies.
Even if the workspace were a persistent volume claim, the failed `build-container` may have left changes behind, and the tool doesn't check what tasks do, so running both again is the safe choice. `clone-repository` also supplies the `CHAINS-GIT_*` results, and rerunning it keeps them matched to the fresh checkout.
`init` writes no files to any workspace and only produces string results from a parameter. Those results are copied into the recovery run, so it can be reused.

RERUN: [clone-repository, prefetch-dependencies]
