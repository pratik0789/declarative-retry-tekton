No environment note was present, so the workspace binding is unknown. Konflux PipelineRuns normally bind `workspace` with a volumeClaimTemplate, which starts empty in a new run, so I'm assuming that here.
Several downstream tasks that haven't run yet read source files from `workspace`: the three sast-* checks, push-dockerfile, and build-source-image if enabled. So clone-repository and prefetch-dependencies must run again to put the source tree and prefetched dependencies back.
build-container, build-image-index and init can be reused because their results are just strings (the image URL and digest point to images already pushed to the registry). Re-cloning gives the same source as long as `revision` is a pinned commit SHA, which is the Konflux default.

RERUN: [clone-repository, prefetch-dependencies]
