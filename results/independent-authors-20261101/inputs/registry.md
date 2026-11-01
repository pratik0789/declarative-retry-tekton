# Environment: the container registry

clone fetches the application source; build compiles it and pushes an image to a Docker Distribution (v2) registry; retag copies the built image to a release tag; promote copies an image to the delivery repository. Tags in the registry are mutable and other pipelines may move them. A retention job may delete old image manifests. The delivery service promote talks to can be briefly unavailable.
The "source" Workspace is bound to a persistentVolumeClaim (the recovery run sees the failed run's files); "copy-input" is an emptyDir scratch volume.
