No environment note exists, so I assumed the default Konflux binding: `workspace` comes from a volumeClaimTemplate. Under that binding the recovery run starts with an empty workspace.

`push-dockerfile` reads the Dockerfile from that workspace, and so do the SAST checks and `build-source-image`, which also need the files `prefetch-dependencies` writes there. So `clone-repository` and `prefetch-dependencies` must run again to refill it.

`init`, `build-container` and `build-image-index` pass only string results, and the image they built is already in the registry. Their copied results stay valid, so they can be reused.

RERUN: [clone-repository, prefetch-dependencies]
