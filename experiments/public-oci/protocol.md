# Prospective public-task OCI recovery composition

This bounded application case grounds recovery requirements in pre-existing public
Tekton Catalog tasks, an unchanged public Skaffold Dockerfile/Go application, and
Docker Distribution registry semantics. It is an authored integration and fault
injection, not a naturally occurring production incident or practitioner study.
The exact source inputs and task-body hashes are pinned before execution. The
Dockerfile’s supported BASE_PREFIX argument points at a private frozen mirror
of digest-pinned Go/Alpine base images, available in both registry worlds. This
keeps the public Dockerfile unchanged while fixing rebuild inputs. Pod egress
to public endpoints is unavailable in the test environment. Host-fetched pinned
image archives are mirrored locally, and a read-only Git daemon serves the exact
public bare commit. The native clone uses that local transport; public-network
clone latency is not measured. Archive hashes and source/private-mirror image
configuration/layer identities are checked. Pod fsGroup is 65532 for shared
workspace access. No task script or application file is changed. The Kaniko step explicitly resolves its pinned image entrypoint as command [/kaniko/executor], avoiding an otherwise required controller-side public-registry lookup. One prior source failed at PodCreationFailed before build execution and is retained separately as calibration/controller-entrypoint; it is not an evaluated source or arm.

The public build-push-gke-deploy pipeline supplies the build/push pattern. Its
GKE-specific deployment is excluded; the resolved application slice uses Catalog
git-clone, Kaniko, and Skopeo-copy tasks for clone -> build -> retag -> promote.
Retag copies the built digest to a release tag. Promotion copies either that
mutable tag (tagged profile) or the build's immutable digest (digest profile) to
a delivery repository. All writes target one fresh, isolated local registry.
Task bodies are retained; runtime images are pinned through supported parameters
or a recorded literal image pin. No GKE or public-registry write occurs.

A disabled delivery Service injects a connection failure at promotion while
clone/build/retag succeed. Two full native failed sources are captured, one per
profile. The Service is restored before recovery. No source object is edited.
For each source, fixed captured Tekton inputs are used in two registry worlds:
(1) release tag moved to a different image while the original source manifest
remains available by digest; (2) release tag moved and the original source
repository manifest removed, so its original digest lookup returns 404. The
registry is not garbage-collected; manifest unavailability is the tested event.
The original manifest and content are retained by the harness for restoring
paired worlds. Workflow tasks cannot access a harness recovery path.

Tagged profile controls are empty selection, fixed retag-only policy,
policy-informed manual selection of retag, fixed rebuild-and-retag policy, and
full restart: 2 worlds x 5 arms = 10 native executions. The digest profile compares empty selection, a fixed build-only policy, and
policy-informed manual selection of build in each world: 6 native executions.
There are 16 execution arms and 2 failed sources, one execution per cell. Arm order uses a fixed shuffled
seed. No probability, policy-authoring superiority or causal timing effect is
inferred. The two policies are fixed across both worlds; the retag-only policy
is conditional on source-digest availability. Its failure outside that assumption
is evidence of a boundary, not a claimed successful repair.

Before each arm, the harness restores the source image reference, release tag,
and delivery target to a recorded common initial state. Delivery starts with a
foreign image to prevent a stale correct destination from passing the audit.
When testing missing-manifest state, source-digest GET must return 404. Restoration
uses raw manifests/blobs or the private harness backup, never planner output.

The semantic predicate independently reads actual registry manifests and layer
bytes at delivery, resolves whiteouts, and hashes the delivered /app binary.
It compares that hash to the binary in the source build image, with pinned Git
commit, original public source files, Dockerfile and source TaskRun output/logs
retained. Native success alone is insufficient. Current task names/conditions,
source snapshot identity, restored registry state, and generated plans/manifests
are audited separately. Expected content is not inferred from labels or the
plan. Full rebuilds may change manifest/config timestamps; binary content,
configuration and source identity determine the application predicate.

The evaluated lesson is the relationship between identity, retention, and repair:
immutable references can avoid a mutable-tag repair while the image remains
available; they do not ensure continued availability. A policy can repeat
publication without rebuilding retained content, while missing source manifests
can require a build and publication. Equal-information manual controls remain.
This is one externally grounded application composition, not broad deployment
validation or a new registry/graph theorem.

## Builder calibration and final binding
A second source compiled the public Go app but Kaniko 1.5.1 failed during multistage cleanup at /product_uuid (device busy), publishing no image. It is separately retained as calibration/multistage-filesystem. Final profiles select pinned Kaniko 1.23.2 (digest 9e69fd4330ec887829c780f5126dd80edc663df6def362cd22e79bcdf00ac53f) through the existing BUILDER_IMAGE parameter and supported EXTRA_ARGS --ignore-path=/product_uuid. This ignores an injected runtime file, not an application input. Public scripts and Dockerfile are unchanged; the pinned entrypoint is resolved explicitly. Source-final-* are the two evaluated sources. Both failed calibration objects and logs remain; neither counts as an evaluated cell. Prospective source-only minima preceded all attempted executions; final builder refinements followed calibration and must not be described as preregistration.
