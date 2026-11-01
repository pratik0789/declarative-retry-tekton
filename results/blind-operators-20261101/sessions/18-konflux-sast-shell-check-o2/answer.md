I'm assuming the workspace uses a `volumeClaimTemplate` (the Konflux default), since this directory has no environment note saying otherwise.

- **The workspace starts empty.** With a `volumeClaimTemplate`, the recovery run gets a fresh volume. `sast-shell-check` scans the cloned source in `workspace`, and so do the tasks that never ran (`sast-snyk-check`, `sast-unicode-check`, `push-dockerfile`, `build-source-image`). So `clone-repository` has to run again.
- **Prefetched dependencies must come back too.** `prefetch-dependencies` writes them into the same workspace, and `build-source-image` needs them, so it reruns as well.
- **The rest can be reused.** `init`, `build-container` and `build-image-index` only pass string results (image URL and digest) to later tasks, and the image is already in the registry. Reusing them is safe as long as `revision` is a pinned commit SHA. If it's a branch, the re-clone could pick up a different commit, and the build tasks would need to run again too.

RERUN: [clone-repository, prefetch-dependencies]
