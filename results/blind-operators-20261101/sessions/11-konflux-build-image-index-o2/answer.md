No environment note exists here, so I can't tell whether the `workspace` binding is a PVC (keeps files) or a `volumeClaimTemplate` (starts empty, the usual Konflux setup). I'm treating it as possibly empty.

`build-image-index` only needs `build-container`'s string results, which get copied. But tasks that never ran need the source files in `workspace`: `sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile` and `build-source-image`. So `clone-repository` and then `prefetch-dependencies` must run again to refill it.

`revision` defaults to `""` (whatever the branch head is), so a fresh clone may get a newer commit than the one the image was built from. Rerunning `build-container` keeps the image, scans and the `CHAINS-GIT_COMMIT` result on the same commit. `init` only passes on string results, so it can be reused.

RERUN: [clone-repository, prefetch-dependencies, build-container]
