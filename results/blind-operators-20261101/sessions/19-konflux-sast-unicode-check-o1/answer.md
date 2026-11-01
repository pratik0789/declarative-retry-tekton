No environment note exists, so I can't tell how `workspace` is bound. Konflux normally uses a volumeClaimTemplate, which means the recovery run starts with an empty workspace.

The failed `sast-unicode-check` and the not-yet-run `sast-snyk-check`, `sast-shell-check`, `push-dockerfile` and `build-source-image` all read source files (or prefetched dependencies) from that workspace, so `clone-repository` and `prefetch-dependencies` must refill it.

The image built by `build-container` and `build-image-index` already sits in the registry, and the tasks after them use only their copied string results (IMAGE_URL/DIGEST). `init` writes nothing to the workspace, so all three can be reused. This assumes `revision` is a fixed commit SHA (as Konflux sets it); if it's a branch, everything after `init` should rerun so the scans match the image.

RERUN: [clone-repository, prefetch-dependencies]
