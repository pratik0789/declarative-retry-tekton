The `promote` task reads `registry-source:5000/release:latest`. That tag can be changed by other pipelines, and the run failed long enough ago that it may now point at someone else's image, so `retag` has to run again to point it back at our build.

`retag` copies `application@<digest>` using the digest it gets from `build`. The retention job may have deleted that image manifest since, so `build` should run again to push a fresh image and digest. `--reproducible` and the pinned commit should give the same image, though base images pulled through `BASE_PREFIX` may have changed since the first run.

`clone` can be reused. It checks out a fixed commit SHA onto the `source` persistentVolumeClaim, which keeps its files, so `build` will find the source tree. `promote` fails because the delivery service is briefly down, and running it again handles that.

RERUN: [build, retag]
