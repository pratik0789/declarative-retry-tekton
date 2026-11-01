I found no environment note, so I'm assuming the usual Konflux setup: the `workspace` binding is a volumeClaimTemplate. Under that binding the recovery run starts with an empty workspace.

Several tasks that never ran need the cloned source in that workspace: `sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile`, and `build-source-image` if enabled. `build-source-image` also needs the prefetched dependencies. So `clone-repository` and `prefetch-dependencies` must run again to refill it.

`init`, `build-container` and `build-image-index` don't need to rerun: the image is already pushed, and the remaining tasks only read its URL and digest results, which get copied over. This is only safe if `revision` is pinned to a commit SHA, as Konflux normally does; otherwise the re-clone could fetch code that doesn't match the built image.

RERUN: [clone-repository, prefetch-dependencies]
