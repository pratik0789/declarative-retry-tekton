Downstream tasks that haven't run yet (sast-snyk-check, sast-shell-check, sast-unicode-check, push-dockerfile, and build-source-image if it's enabled) read the source files from the `workspace` workspace.
There's no environment note, so I'm assuming Konflux's usual `volumeClaimTemplate` binding. That means the recovery run starts with an empty workspace, so clone-repository and prefetch-dependencies must run again to put the source and prefetched dependencies back (source-build needs both).
The image already in the registry is reused through the copied IMAGE_URL/IMAGE_DIGEST results, so build-container and build-image-index stay reused. init has no workspace and only string results, so it is reused too. This assumes `revision` is a pinned commit, so the re-clone matches the built image.
If the workspace is actually a persistentVolumeClaim, the files are still there and RERUN: [] would be enough.
RERUN: [clone-repository, prefetch-dependencies]
