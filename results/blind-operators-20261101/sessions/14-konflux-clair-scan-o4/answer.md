No environment note exists, so I can't tell how the `workspace` is bound. Konflux normally uses a volumeClaimTemplate, and that starts empty in the recovery run.

Tasks that will run next and read the source from `workspace` are sast-snyk-check, sast-shell-check, sast-unicode-check and push-dockerfile. Without a re-clone they would scan an empty directory or fail to find the Dockerfile.

init, build-container and build-image-index only pass string results (IMAGE_URL/IMAGE_DIGEST) to later tasks, so they can be reused. The image is already pushed. I'm reusing prefetch-dependencies on the assumption that nothing downstream reads its output; if your sast checks scan the prefetched dependencies, add it to the list.

RERUN: [clone-repository]
