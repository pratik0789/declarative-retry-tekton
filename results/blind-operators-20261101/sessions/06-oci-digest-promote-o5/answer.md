Promote copies `application@$(tasks.build.results.IMAGE_DIGEST)`. A digest always names the same content, so if anything has moved the tags since the failure, the reused build result still points at exactly the image that was built.

If the retention job has since deleted that manifest, promote fails with an error rather than publishing the wrong image. The likely cause of this failure was the delivery service being briefly down, so a promote-only retry is the right first step.

Rebuilding is not "safe extra work": the base image comes through a tag (`registry-source:5000/base/`) that other pipelines can move, so a rebuild could produce a new digest. That would no longer match `release:latest`, so retag would need to run again, and re-pushing `release:latest` could overwrite a newer image another pipeline put there. clone's files are still on the persistent volume, and retag's output isn't an input to promote.

RERUN: []
