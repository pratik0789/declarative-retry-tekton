Rationale:
- There's no environment note, so I'm assuming the standard Konflux binding: `workspace` is a volumeClaimTemplate. Under that binding the recovery run starts with an empty source workspace.
- `sast-snyk-check` scans the source tree in `workspace`, so `clone-repository` has to fill it again. `prefetch-dependencies` also has to rerun, because it writes the prefetched dependencies into that same workspace and the scan should see the same tree as the original run.
- `init`, `build-container` and `build-image-index` only hand on string results (image URL and digest), which get copied over. Rerunning them would just rebuild and push a new image that no longer matches the reused results. This assumes `revision` is a fixed commit, so the re-clone matches the source the image was built from.

RERUN: [clone-repository, prefetch-dependencies]
