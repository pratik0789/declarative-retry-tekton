No environment note exists, so the workspace binding is unknown. Konflux PipelineRuns normally use a `volumeClaimTemplate` for `workspace`, which means the recovery run's workspace starts empty.

sast-shell-check, sast-snyk-check, sast-unicode-check, push-dockerfile and build-source-image all read source files from that workspace, so clone-repository and prefetch-dependencies must run again to refill it.

init, build-container and build-image-index only pass on string results (proxy config, image URL and digest), which are copied over, so they can be reused.

This assumes `revision` is a pinned commit, as Konflux sets it. If it's a branch name, the re-clone could pick up newer code than the built image, and build-container and build-image-index would also need to rerun.

RERUN: [clone-repository, prefetch-dependencies]
