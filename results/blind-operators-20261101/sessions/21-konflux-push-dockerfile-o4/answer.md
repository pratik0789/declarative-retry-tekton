No environment note is present, so I'm assuming the standard Konflux setup: the `workspace` is a volumeClaimTemplate, so the recovery run starts with an empty volume, and `revision` pins a commit.

`push-dockerfile` needs the source checkout (the Dockerfile at `path-context`/`dockerfile`), and so do the never-run SAST tasks and `build-source-image`. With an empty volume, `clone-repository` must run again to put it back.

`init`, `build-container` and `build-image-index` can be reused: the image is already in the registry, and their string results (`IMAGE_URL`/`IMAGE_DIGEST`) are copied over. `prefetch-dependencies` only feeds the build, and no remaining task uses its output. If `revision` is a branch rather than a pinned commit, the clone could pick up newer code, and everything from `clone-repository` onward would need to rerun.

RERUN: [clone-repository]
