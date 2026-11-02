No environment note exists here, so I can't tell how the `workspace` Workspace is bound. Konflux normally uses a volumeClaimTemplate, and that would start empty in the recovery run.

The failed `deprecated-base-image-check` only needs `build-image-index`'s string results, which get copied over. But tasks that haven't run yet (`sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile`, and possibly `build-source-image`) read the source tree from that Workspace.

So `clone-repository` must run again to put the source back. If `revision` is a branch rather than a fixed commit, check that the new clone gets the same commit as the image that was built. The image is already pushed, so `init`, `prefetch-dependencies`, `build-container` and `build-image-index` can be reused, since none of the remaining tasks need their files.

RERUN: [clone-repository]
