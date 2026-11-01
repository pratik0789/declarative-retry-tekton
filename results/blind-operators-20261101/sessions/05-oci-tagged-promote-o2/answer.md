Promote copies `release:latest`. That tag is mutable, so another pipeline may have moved it since the failure. retag has to run again to point it back at our image.

retag copies `application@<digest>` using build's saved digest. The retention job may have deleted that manifest since the run failed "a while ago". Rerunning build pushes the image again, so that digest definitely exists.

clone can be reused because the "source" workspace is a persistentVolumeClaim and still holds the checkout at the pinned commit. The emptyDir "copy-input" workspace doesn't matter, because both skopeo tasks get their image URLs as params and never read `url.txt`.

RERUN: [build, retag]
