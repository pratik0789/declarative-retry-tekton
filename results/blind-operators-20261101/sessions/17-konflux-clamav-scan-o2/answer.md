Workspace storage is my main assumption: there's no environment note, so I'm assuming `workspace` uses the usual Konflux `volumeClaimTemplate`. If it's actually a persistentVolumeClaim, the files survive and `RERUN: []` is enough.

- **Why the workspace matters:** reused tasks keep their string results, but a `volumeClaimTemplate` workspace starts empty in the recovery run. The tasks that never ran (`sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile`, and possibly `build-source-image`) read source files from that workspace.
- **What must rerun:** `clone-repository` refills the source tree. `prefetch-dependencies` refills the prefetched dependencies, which `build-source-image` packages if it's enabled. I'm assuming `revision` is pinned to a commit, so the re-clone matches the image that was already built.
- **What can be reused:** `build-container` and `build-image-index` are only consumed through their IMAGE_URL/IMAGE_DIGEST results, which are copied over. `init`'s results aren't used by anything downstream.

RERUN: [clone-repository, prefetch-dependencies]
