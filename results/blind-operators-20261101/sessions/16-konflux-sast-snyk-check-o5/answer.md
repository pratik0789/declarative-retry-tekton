No environment note exists, so I assumed Konflux's usual binding: `workspace` comes from a volumeClaimTemplate, which starts empty in the recovery run. That is an assumption, not something I could check.

`sast-snyk-check` scans source files in `workspace`. So do `sast-shell-check`, `sast-unicode-check`, `push-dockerfile` and `build-source-image`, which haven't run yet. So `clone-repository` must run again to put the source back, and `prefetch-dependencies` must run again to put back the dependencies it saved there, which the source-image build uses.

`init`, `build-container` and `build-image-index` only pass on string results (image URL and digest), and the image is already in the registry, so they can be reused. This is only safe if `revision` is a fixed commit, so the re-clone matches what was built. If it's a branch, or the workspace is a persistentVolumeClaim, the answer changes.

RERUN: [clone-repository, prefetch-dependencies]
