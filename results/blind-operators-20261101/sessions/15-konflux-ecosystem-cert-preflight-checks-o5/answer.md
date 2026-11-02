Rationale:
- The failed task, ecosystem-cert-preflight-checks, needs only the string result `build-image-index.results.IMAGE_URL`. Reused tasks copy their string results, so build-container, build-image-index and init can be reused, and the built image and digest stay the same.
- Five tasks that never ran read files from the shared `workspace`: sast-snyk-check, sast-shell-check, sast-unicode-check, push-dockerfile and build-source-image (if enabled). They need the cloned source and the prefetched dependencies.
- There is no environment note, so I'm assuming the Konflux default: `workspace` is bound with a volumeClaimTemplate, which starts empty in the recovery run. clone-repository and prefetch-dependencies must run again to refill it. This assumes `revision` points to a fixed commit; if it names a branch that has moved, the re-clone would not match the built image.

RERUN: [clone-repository, prefetch-dependencies]
