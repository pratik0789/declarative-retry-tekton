Rationale:
- `sast-snyk-check` scans the source in the `workspace` workspace, and so do the not-yet-run `sast-shell-check`, `sast-unicode-check`, `push-dockerfile` and possibly `build-source-image`. There's no environment note, so I'm assuming the usual Konflux PipelineRun binding, a volumeClaimTemplate, which means the recovery run starts with an empty workspace.
- So `clone-repository` must run again to restore the checkout. `prefetch-dependencies` must also run again to put back the dependencies it fetched into the workspace (source-build uses them). Both rebuild the same tree as long as `revision` is a fixed commit.
- `init`, `build-container` and `build-image-index` only pass on string results (image URL and digest), which get copied over. Reuse them so the already-pushed image digest stays the same.

RERUN: [clone-repository, prefetch-dependencies]
