No environment note exists, so I'm assuming the standard Konflux setup: `workspace` is bound through a volumeClaimTemplate. A reused run would then start with an empty source directory.

Several tasks that haven't run yet read source files from `workspace`: sast-snyk-check, sast-shell-check, sast-unicode-check, push-dockerfile (it needs the Dockerfile) and possibly build-source-image. So clone-repository has to run again, and prefetch-dependencies too, so the scans see the same tree the image was built from.

init, build-container and build-image-index don't need to run again. Their string results (IMAGE_URL/IMAGE_DIGEST) are copied over and still point to the image that was already pushed, which clamav-scan and the other image checks use.

RERUN: [clone-repository, prefetch-dependencies]
