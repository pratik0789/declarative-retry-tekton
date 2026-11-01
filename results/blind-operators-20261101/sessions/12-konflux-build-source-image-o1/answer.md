RERUN: [clone-repository, prefetch-dependencies]

1. **Why the workspace matters:** `build-source-image` needs the shared `workspace` to contain the cloned source and the prefetched dependencies. The four workspace-based tasks that never ran (`push-dockerfile`, `sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`) need those files too.
2. **Assumption about storage:** there's no environment note, so I'm assuming Konflux's usual `volumeClaimTemplate` binding. That volume starts empty in the recovery run, which means `clone-repository` and `prefetch-dependencies` have to fill it again.
3. **What can be reused:** `init` doesn't use the workspace. `build-container` and `build-image-index` only pass on string results (the image URL and digest) for an image that's already pushed, and copying those is safe. Rebuilding would produce a new digest for no benefit.
4. **Risk:** this is only correct if `revision` is pinned to a commit SHA. If it's a branch name, the re-clone could pick up newer code that no longer matches the built image. If the workspace is actually a persistent volume claim, RERUN: [] would be enough.
