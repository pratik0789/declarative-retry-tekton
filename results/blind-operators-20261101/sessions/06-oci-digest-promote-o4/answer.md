Promote copies `application@<IMAGE_DIGEST>`, using the digest result reused from the old build. The run failed a while ago, and a retention job may have deleted that manifest since. So build must re-push the image and produce a digest that exists now.

The base image is pulled by a mutable tag (`BASE_PREFIX`), so the rebuilt digest may differ from the first one. Retag must run again so `release:latest` points to the same image that promote delivers.

Clone can be reused: its source is pinned to a commit SHA and the files persist on the "source" PVC. The promote failure itself was probably the delivery service's brief outage, which needs nothing beyond the automatic rerun of promote.

RERUN: [build, retag]
