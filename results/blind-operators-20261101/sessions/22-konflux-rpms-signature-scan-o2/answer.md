No environment note exists, so I can't tell how `workspace` is bound. Konflux normally binds it with a volumeClaimTemplate, which means the recovery run starts with an empty volume.

Several tasks that haven't run yet read source files from that workspace: the three SAST checks, `push-dockerfile` and, if enabled, `build-source-image`. So `clone-repository` has to run again to put the source back. `prefetch-dependencies` also writes into the workspace, and `build-source-image` uses its output, so it must run again after the clone.

`init`, `build-container` and `build-image-index` can be reused. They only pass along string results, and the image they built is already in the registry.

RERUN: [clone-repository, prefetch-dependencies]
