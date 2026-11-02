Rationale:
- `promote` copies from `release:latest`, which is a mutable tag. Other pipelines may have moved it since the failure, so `retag` has to run again to point it back at this run's image.
- `retag` uses the reused `build` digest to copy `application@<digest>`. The retention job may have deleted that manifest since, so `build` should run again to push the image back.
- `clone` can be reused. The `source` workspace is a persistentVolumeClaim that still holds the checkout at the pinned commit, which is all `build` needs. `copy-input` is an emptyDir, but these tasks pass all their inputs as params and never read it.

RERUN: [build, retag]
