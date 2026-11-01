No environment note exists, so I assumed the `workspace` binding follows the Konflux default, a `volumeClaimTemplate`. That means the recovery run starts with an empty workspace.

Several tasks that never ran read source files from `workspace`: `sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile`, and `build-source-image` if enabled. So `clone-repository` must run again, and so must `prefetch-dependencies`, which also writes its prefetched dependencies into that workspace.

`init`, `build-container` and `build-image-index` can be reused, because their outputs are strings (the image is already in the registry) and the copied values stay valid.

RERUN: [clone-repository, prefetch-dependencies]
