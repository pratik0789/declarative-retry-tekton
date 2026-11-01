Rationale:
- `build-source-image` reads source code and prefetched dependencies from the shared `workspace`. The `sast-*` checks and `push-dockerfile`, which haven't run yet, also read from it.
- There's no environment note, so the workspace binding is unknown. Konflux normally binds it with a volumeClaimTemplate, which starts empty in the recovery run, so `clone-repository` and `prefetch-dependencies` have to fill it again.
- The image tasks (`build-container`, `build-image-index`) hand over only string results (IMAGE_URL/DIGEST), which get copied, so reuse them. `init` doesn't touch the workspace. This assumes `revision` is a pinned commit, so the re-clone matches the image that was built.

RERUN: [clone-repository, prefetch-dependencies]
