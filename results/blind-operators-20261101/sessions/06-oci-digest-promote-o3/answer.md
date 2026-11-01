The recovery run can reuse clone, build and retag; only the failed promote task needs to run again.

- promote copies `application@$(tasks.build.results.IMAGE_DIGEST)`. That is a fixed digest, so if another pipeline moves a tag, promote still copies the image we built.
- promote doesn't read retag's `release:latest` tag or anything in the `copy-input` volume, which starts empty anyway. clone's files are kept on the "source" claim.
- If the retention job has deleted that image since the failure, the copy fails with an error instead of delivering the wrong image. Rebuilding first wouldn't help: the base image tag may have moved, so a rebuild could produce a different digest.
- If the delivery service is briefly down, rerunning promote is enough.

RERUN: []
