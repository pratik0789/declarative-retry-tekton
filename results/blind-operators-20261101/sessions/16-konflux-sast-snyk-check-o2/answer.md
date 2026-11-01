Without an environment note, assume the standard Konflux binding: `workspace` comes from a volumeClaimTemplate, so the recovery run starts with an empty workspace. The failed `sast-snyk-check` and the tasks still to run (`sast-shell-check`, `sast-unicode-check`, `push-dockerfile`, `build-source-image`) all read source files from that workspace.
`clone-repository` must run again to put the source back, and `prefetch-dependencies` must run again to restore the dependency files that the source scans and `source-build` depend on.
`init`, `build-container` and `build-image-index` can be reused. Their image URL, digest and proxy results are copied over, and the image already exists in the registry.
This assumes `revision` names a fixed commit, so the re-clone matches what was built.

RERUN: [clone-repository, prefetch-dependencies]
