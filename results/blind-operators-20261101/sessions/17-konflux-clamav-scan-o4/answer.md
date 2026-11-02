Konflux usually binds `workspace` with a volumeClaimTemplate, and there's no environment note saying otherwise. If so, a reused run starts with an empty source volume.
Several tasks that never ran read source from that volume: sast-snyk-check, sast-shell-check, sast-unicode-check, push-dockerfile, and build-source-image if enabled. So clone-repository has to run again, and prefetch-dependencies too, since it also writes into that volume.
build-container, build-image-index and init only pass string results (image URL/digest), which get copied over. Reusing them keeps the image that was already scanned and pushed, and doesn't rebuild it.
This assumes `revision` is a fixed commit SHA, so the re-clone matches the source of the built image.

RERUN: [clone-repository, prefetch-dependencies]
