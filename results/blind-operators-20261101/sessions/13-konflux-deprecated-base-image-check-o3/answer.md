Without an environment note, nothing says how the `workspace` is bound. Konflux PipelineRuns normally use a volumeClaimTemplate, which starts empty in the recovery run, so I've assumed that.
Five tasks that haven't run yet read source files from that workspace: sast-snyk-check, sast-shell-check, sast-unicode-check, push-dockerfile and build-source-image. So clone-repository has to run again, and prefetch-dependencies too, because source-build needs the prefetched dependencies in the workspace.
init, build-container and build-image-index only hand on string results (IMAGE_URL/IMAGE_DIGEST), which get copied over, and the image is already in the registry. Rebuilding it would only risk a different digest, so those three are reused.

RERUN: [clone-repository, prefetch-dependencies]
