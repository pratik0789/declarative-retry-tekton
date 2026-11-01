No environment note exists here, so I assumed the Konflux default: the `workspace` binding is a `volumeClaimTemplate`. That means the recovery run starts with an empty source volume. If it's actually a persistentVolumeClaim, the files survive and `RERUN: []` is enough.

`sast-snyk-check` and the other workspace users that never ran (`sast-shell-check`, `sast-unicode-check`, `push-dockerfile`, and `build-source-image` if enabled) need the cloned source. `clone-repository` refills it, and `prefetch-dependencies` puts back the prefetched deps output that `build-source-image` reads.

`build-container` and `build-image-index` can be reused: the image is already in the registry and their copied `IMAGE_URL`/`IMAGE_DIGEST` results stay valid. `init` has no workspace. This assumes `revision` is pinned to a commit; if it's a moving branch, the re-clone could scan different code than was built.

RERUN: [clone-repository, prefetch-dependencies]
