I'd rerun `build` and `retag` and reuse `clone`.

- **Why `retag` must rerun:** `promote` copies the mutable tag `release:latest`, which other pipelines may have moved since the failure. Rerunning `retag` points it back at this run's image just before `promote` reads it.
- **Why `build` must rerun:** `retag` reads this run's image by digest (`application@<digest>`), and the retention job may have deleted that manifest by now, which would make `retag` fail. Rebuilding pushes the image again and produces a fresh digest.
- **Why `clone` can be reused:** the `source` workspace is a persistentVolumeClaim, so the checkout of the pinned commit (`035e3dd…`) is still there for `build`.

RERUN: [build, retag]
