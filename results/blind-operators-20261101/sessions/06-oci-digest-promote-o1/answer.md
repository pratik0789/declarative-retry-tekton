Promote copies `application@<IMAGE_DIGEST>`. A digest is a content address, so if someone moves a tag, promote still copies exactly the image this run built. The reused `IMAGE_DIGEST` result is therefore safe to keep.

If the retention job has deleted that manifest, skopeo fails with a clear error and does not push the wrong image. Rebuilding first would be worse: the build pulls its base image by a mutable tag (`registry-source:5000/base/`), so it could produce a new digest that no longer matches the image retag already published as `release:latest`.

Promote passes its images as parameters, not through the `url.txt` file, so starting with an empty `copy-input` scratch volume is fine. A brief outage of the delivery service only needs promote retried. Re-running retag could overwrite a tag another pipeline has since moved, and nothing downstream uses its output.

RERUN: []
