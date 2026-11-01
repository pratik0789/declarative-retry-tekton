No environment note exists, so the workspace binding is unknown. Konflux PipelineRuns normally bind `workspace` with a volumeClaimTemplate, which means the recovery run would start with an empty volume.
Tasks that have not run yet read source files from that workspace: push-dockerfile, sast-snyk-check, sast-shell-check, sast-unicode-check, and build-source-image (which also needs the prefetched dependencies). So clone-repository and prefetch-dependencies must run again to refill it.
init, build-container and build-image-index pass only string results, which are copied over, so the image URL and digest stay the same.
This assumes `revision` is pinned to a commit SHA, as Konflux does. If it is a branch, clone could fetch a newer commit than the one in the built image.

RERUN: [clone-repository, prefetch-dependencies]
