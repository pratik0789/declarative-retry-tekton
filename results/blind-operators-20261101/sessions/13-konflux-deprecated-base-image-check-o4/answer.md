No environment note exists, so I can't confirm how `workspace` is bound. Konflux normally uses a volumeClaimTemplate, and reused tasks would then leave it empty in the recovery run.
The sast-snyk, sast-shell, sast-unicode and push-dockerfile tasks have never run, and they read the cloned source and Dockerfile from that workspace. So clone-repository must run again, and prefetch-dependencies too, to rebuild the workspace as the original build left it.
build-container and build-image-index can be reused: the image is already in the registry, and their copied IMAGE_URL/IMAGE_DIGEST results are all the remaining tasks need.
init has no workspace state, so it is safe to reuse. This assumes `revision` points to the same commit (a SHA), so re-cloning gets the source that was built.

RERUN: [clone-repository, prefetch-dependencies]
