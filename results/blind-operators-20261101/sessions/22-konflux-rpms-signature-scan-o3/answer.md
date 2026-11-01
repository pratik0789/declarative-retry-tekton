No environment note exists, so I'm assuming the Konflux default: the `workspace` is bound through a volumeClaimTemplate. That means the recovery run starts with an empty workspace, even though the copied string results (IMAGE_URL/DIGEST, url/commit) are still valid.

`rpms-signature-scan` only needs `build-image-index` string results, but some tasks that never ran read source files from the workspace: `sast-snyk-check`, `sast-shell-check`, `sast-unicode-check`, `push-dockerfile` (the Dockerfile), and `build-source-image` (source plus prefetched deps). On an empty workspace they would quietly scan or push nothing.

So rerun `clone-repository` to repopulate the source, and `prefetch-dependencies`, which writes its output into that same workspace. `init`, `build-container` and `build-image-index` can be reused because only their string results are consumed and the image is already pushed. If the workspace is actually a persistentVolumeClaim, the answer would be RERUN: [].

RERUN: [clone-repository, prefetch-dependencies]
