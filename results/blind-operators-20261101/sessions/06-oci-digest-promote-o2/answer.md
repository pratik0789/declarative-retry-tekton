Promote copies `application@<IMAGE_DIGEST>`. A digest always names the same image, so moved tags can't change what gets promoted, but the retention job may have deleted that manifest since the run failed. I can't check the registry, so reusing build's digest could leave promote with nothing to copy.
Rerunning build pushes a fresh image and gives a new digest. Retag then has to run again so `release:latest` points at the image this run promotes, not the old and possibly deleted one.
Clone can be reused: `source` is a PVC that still holds the checkout at the pinned commit, and no task uses clone's results.
The delivery outage only needs promote to run again, which it does anyway.

RERUN: [build, retag]
