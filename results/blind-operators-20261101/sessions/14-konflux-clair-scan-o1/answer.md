No environment note exists, so I'm assuming the Konflux default: the `workspace` is bound with a volumeClaimTemplate, which means the recovery run starts with it empty.

The tasks that never ran still need the cloned source in that workspace: sast-snyk-check, sast-shell-check, sast-unicode-check, push-dockerfile, and build-source-image if it's enabled. So clone-repository has to run again, and prefetch-dependencies too, because its prefetched deps also live in the workspace (build-source-image uses them).

build-container and build-image-index only pass on string results (the image URL and digest), which the recovery run copies over, so the image isn't rebuilt. init doesn't use the workspace. This assumes `revision` is a fixed commit, so the re-clone gets the same code that was built.

RERUN: [clone-repository, prefetch-dependencies]
