# Locked prospective source assessment: public OCI recovery — refined inputs

## Current assessment and revision boundary

This revision assesses the refined prospective input packet at https://artifacts.internalmeta.com/a/2479895562538338 . It was prepared without access to source/recovery executions, policies, plans, protocol, harness, other assessments, or a revised manuscript. No native experiment or planner ran. This remains an **automated independent source assessment, not a practitioner authoring study**. No venue verdict is inferred.

The supplied brief states that execution still has not started; this is a supplied chronology premise, not an independently inspected experiment status. The initial assessment is preserved below as a clearly labeled historical record, and the artifact also preserves its previous version. **This current addendum supersedes the historical assessment’s missing-base-pin/absent-BASE_PREFIX statements.** It does not revise outcomes retrospectively. Task roles, worlds, minima, native-success predictions, and delivery predicate remain unchanged.

## Final packet integrity

Re-downloaded all 14 files into a separate directory. All 13 file hashes listed in the refined SHA256.json match. Changes from the initially assessed packet are limited to SOURCE_BRIEF.md, both profile JSON files, runtime-images.json, and SHA256.json. The eight pinned public task/application/API bodies and pins.json are byte-identical to the already independently fetched upstream versions.

| Refined file | SHA-256 |
|---|---|
| SHA256.json | `4d26047ce2aa363ecb7b8ef5bfcf356d95a5efbde35bfb59f0be1d82ff005bc7` |
| SOURCE_BRIEF.md | `351674b9c72e91e6bb3a60fa755b6f5bbde2fb05c3fca4a9c6b9450a4e819e6b` |
| profile-tagged.json | `425fdd57fe5498079d3276d024200f85922e2b41b6ab90c522771be3782b0c8b` |
| profile-digest.json | `a9da10f4bf1e54f2d9fccb82102e01a99fd65a269614860f03f096fd202ca8f0` |
| runtime-images.json | `1d86f8d2aabc2a76da043cdd81cb4c5e05b3058c953a1861d6a0542ee9294aed` |

Each final profile’s spec and its last-applied configuration include the same six Kaniko EXTRA_ARGS:

```text
--insecure
--insecure-pull
--cache=false
--reproducible
--force
--build-arg=BASE_PREFIX=registry-source:5000/base/
```

The old profile had four args, without insecure-pull or the BASE_PREFIX override. No task script or command body changed. The unchanged Dockerfile uses its existing BASE_PREFIX argument to select:

```text
registry-source:5000/base/golang:1.18
registry-source:5000/base/alpine:3
```

The refined brief specifies a private frozen mirror, held available and unchanged in both registry worlds. Runtime-images.json additionally identifies the original base inputs by digest:

- Go base: `sha256:50c889275d26f816b5314fc99f55425fa76b18fcaf16af255f5d57f09e1f48da`.
- Alpine final base: `sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a`.

The mirror URLs used by the Dockerfile still contain tags, so their immutability is an **explicit frozen-mirror environmental premise**, not syntax-level `FROM ...@sha256` binding. The supplied digest pins make that premise checkable. They are not execution observations. The source/delivery mutations target application/release/delivery references; the base repository remains available by stipulation. Native serialization defaults are now explicitly documented; they do not alter the compared task behaviors.

The two final Pipeline specs still differ substantively only in promote.srcImageURL: release:latest for tagged, application@build.IMAGE_DIGEST for digest. Their four-task chain and both external-world definitions are unchanged.

## Current locked minima and expected outcomes

Assume the stipulated intact source Workspace, the frozen available base mirror, fixed runtime/platform/build inputs, correct current-versus-saved result binding, successful registry copy/build implementation, and no concurrent mutations. Then:

| Profile/world | Minimum earlier reruns | Executed recovery tasks | Empty earlier selection | Full restart |
|---|---|---|---|---|
| Tagged/world 1: source D0 available, release tag changed | {retag} | retag, promote: 2 | 1 task; native success expected, delivers unrelated Alpine | 4 tasks; native success and desired application delivery expected |
| Tagged/world 2: source D0 unavailable, release tag changed | {build, retag} | build, retag, promote: 3 | 1 task; native success expected, delivers unrelated Alpine | 4 tasks; native success and desired application delivery expected |
| Digest/world 1: source D0 available, release tag changed | empty set | promote: 1 | 1 task; native success and desired delivery expected | 4 tasks; native success and desired application delivery expected |
| Digest/world 2: source D0 unavailable, release tag changed | {build} | build, promote: 2 | 1 attempted task; missing-manifest copy/native failure expected; destination remains unrelated Alpine | 4 tasks; native success and desired application delivery expected |

Counts are workflow tasks in one attempt, excluding source execution, subprocesses, and unforeseen generator helper tasks. No timing or resource-cost measurements are inferred. Native success and desired binary/config delivery remain separate recorded outcomes.

Necessity is unchanged: retag is the only earlier task writing the tagged profile’s promotion reference; build is the only supplied task recreating the missing application-repository source. Clone is unnecessary because the checkout remains present and unchanged. Digest promotion does not consume the release tag, so retag need not repeat there. It may leave release:latest pointing to Alpine while still delivering the correct application; requiring repair of every intermediate tag would change the specified goal.

Sufficiency is now supported by a fixed-base-input premise that the initial packet did not provide. An available D0 directly supplies the original application image. When D0 is missing, build uses the same pinned checkout, frozen Go/Alpine bases, unchanged Dockerfile, and fixed flags; its emitted D1 is consumed by retag/digest promote as applicable. This supplies the intended rebuild route without re-cloning or administrative manifest restoration. D1 need not equal D0: the desired predicate remains original application binary content and effective executable configuration, not original manifest byte identity.

The refinement addresses the particular open input-identity problem identified in the initial assessment. It does **not** itself prove that the pinned Kaniko/Go implementation emits byte-identical binaries or that actual native runs use the stated mirror content. The correct prospective expectation is desired identity under the fixed-input premise, followed by direct binary/config comparison in the execution evidence. Full restart is subject to the same identity check; its native success must not substitute for that check.

## Evidence still required after execution

The earlier registry/task contract derivations and external delivery predicate remain operative. In particular, later immutable evidence should show:

1. Actual private mirror manifests and their verified digest/config/layer chains for the Go and Alpine inputs, before sources and recoveries, with the same platform selected. If mirror content is format-converted, record the original-to-mirror identity mapping and verify the expected image content, not just tag names. Confirm mirror availability and stability across both worlds.
2. Actual resolved BASE_PREFIX/build args, runner images, target architecture, and retained Workspace source-byte hashes. The profile declares these inputs; the execution must establish them.
3. Original successful build binary/config fingerprint captured before external mutation, and each rebuild’s binary/config fingerprint. Require equality of application executable bytes/mode and specified effective CMD/Entrypoint/WorkingDir/environment/User/OS/architecture. Manifest timestamp changes alone need not violate the predicate.
4. Exact application@D0 availability in world 1 and 404 in world 2 with suitable manifest media-type negotiation. Deleting this repository manifest does not establish blob deletion, destruction of other copies, or garbage collection.
5. Generated/native resolved URLs and result values: inherited build uses D0; rerun build’s consumers use D1. Digest binding constrains selection but cannot supply missing bytes.
6. Actual destination manifest/config/layers reconstructed into the effective filesystem, verified independently of planner expectations, plus native conditions and actual TaskRun names/counts. Tagged empty recovery should expose native success with incorrect delivery. Digest empty world 2 should expose missing content as failure.

No additional prerequisite task is inferred from the mirror refinement. No policy, harness behavior, or experimental result is presumed. The four minimum sets and qualitative predictions are prospectively locked against the **final hashes above**. Await actual results and the revised manuscript for the subsequent independent review.

---

# Historical initial assessment — prior packet only

The following text preserves the initial source reasoning and its then-missing build-input premise. Its comments that no BASE_PREFIX override/base pin was supplied describe the **old packet**, not the refined inputs assessed above.

# Locked prospective source assessment: public OCI application recovery

## Status, scope, and inputs

This is an **automated independent source assessment, not a practitioner authoring study**. It contains predictions derived without consulting policies, a protocol, a harness, failed sources, recovery outputs, other assessments, or manuscript revisions. No native experiment, planner, or application code was executed. No source or repository was edited and no person was contacted. No venue recommendation is inferred here.

Input package: https://artifacts.internalmeta.com/a/2479895562538338 . The supplied SOURCE_BRIEF states that no failed sources or recovery arms have executed. That chronology is a supplied premise; this assessment does not independently inspect experiment activity.

Read: SOURCE_BRIEF.md; both complete fixed profile specs and their embedded task bodies; full git-clone/kaniko/skopeo-copy task definitions; the public build/push/GKE pipeline pattern; Skaffold Dockerfile, go.mod, and main.go; image pins; and the registry API sections governing content identity, manifest retrieval/deletion/publication, blob retrieval/upload/mounting, and API errors. All 13 input-file hashes listed in SHA256.json independently matched. SHA256.json itself hashes to `e0400d968b7245d725cca9b0ed15db54ec4003e1cd1bd1fbc99d33980e81f380`.

The eight upstream source files were also fetched directly from the specified commits and matched the supplied bytes exactly:

- Catalog commit `112374ce1262b133a77300bd8133445722605e04`: git-clone 0.10, kaniko 0.6, skopeo-copy 0.4, and build-push-gke-deploy 0.1.
- Skaffold commit `035e3ddc3589c1c64cb69c563b178f34b1cccc5e`: examples/getting-started/{Dockerfile,go.mod,main.go}.
- Distribution API commit `4772604ae973031ab32dd9805a4bccf61d94909f`: docs/spec/api.md.

Upstream provenance is verified for these bodies, not for the application integration as a pre-existing upstream pipeline. The four-task integration is authored. The original public pattern builds/pushes then deploys to GKE; these fixed profiles instead clone, build, retag, and promote between registries. GKE deployment is excluded. The Skaffold application is the public getting-started Go program repeatedly printing “Hello world!”. This is an actual compiled public application, with modest application complexity.

## Exact profile reconstruction

Both profiles have the execution chain:

```text
clone -> build -> retag -> promote
```

There are no finally tasks or when conditions. The only substantive difference between their Pipeline specs is promote.params.srcImageURL.

| Task | Contract and fixed inputs | Effect relevant to recovery |
|---|---|---|
| clone | Catalog git-clone; pinned Skaffold revision; sparse path examples/getting-started; submodules=false; depth=1 | Creates the checked-out source Workspace and records commit/url/committer-date |
| build | Catalog kaniko; source Workspace; Dockerfile and context at examples/getting-started; IMAGE=registry-source:5000/application:build | Builds/pushes application image and records IMAGE_DIGEST; IMAGE_URL is also recorded |
| retag | Catalog skopeo-copy; src=docker://registry-source:5000/application@$(tasks.build.results.IMAGE_DIGEST); dest=docker://registry-source:5000/release:latest | Publishes the built image under the mutable release tag |
| promote, tagged | Catalog skopeo-copy; src=docker://registry-source:5000/release:latest; dest=docker://registry-delivery:5000/delivery:latest | Copies whichever image the release tag currently identifies |
| promote, digest | Catalog skopeo-copy; src=docker://registry-source:5000/application@$(tasks.build.results.IMAGE_DIGEST); same delivery destination | Copies the image addressed by the build result digest |

Clone/build share the declared source Workspace. Retag/promote bind the declared copy-input Workspace to the Catalog images-url Workspace. Their nonempty srcImageURL/destImageURL parameters select the single-copy branch; they do not read url.txt. Skopeo performs a registry copy without a Docker daemon. In this branch the copy command is the final command of the script’s conditional, so its failed exit status is not overwritten by the multi-image branch’s later echo. Multi-image-loop error behavior is irrelevant to these profiles.

Build EXTRA_ARGS are `--insecure`, `--cache=false`, `--reproducible`, `--force`. Both copy tasks set source/destination TLS verification false and use the default multiArch=system. No task implements “is this the desired application?” semantic validation. In particular, skopeo-copy does not reject a valid unrelated Alpine image merely because the application author wanted a different image.

The submitted profiles include server-added empty fields and explicit default string types. Comparing their last-applied embedded task specs to Catalog task specs shows no body changes except the Skopeo runtime image digest substitution; clone uses a supplied immutable gitInitImage parameter, and Kaniko/bash defaults already have digests. This distinguishes API/default normalization from authored behavioral changes.

## State and identity notation

Let:

- C0 = the pinned application checkout, retained unchanged in the source Workspace.
- D0 = the successful source build’s IMAGE_DIGEST in the application repository.
- B0 = the actual original application binary bytes extracted from that built image.
- E0 = its effective execution configuration, including architecture/OS, command, entrypoint, working directory, relevant environment, user, and executable path/mode.
- A = the unrelated Alpine image substituted under release:latest and initially under delivery:latest.
- D1 = a digest emitted by a successful rebuild; D1 need not equal D0.

The desired result is delivery of B0 with E0, as the brief specifies. It is not simply a successful copy, presence of a delivery tag, or exact equality of a newly built manifest digest with D0. Manifest timestamps or metadata may change while the stipulated application identity remains the same.

At the captured failure, clone/build/retag succeeded and promote failed because delivery was unavailable. The endpoint is restored for recovery. Across worlds, source Tekton objects and saved results remain identical. Both worlds have release:latest=A and delivery:latest=A initially.

- **World 1 — source manifest retained:** application@D0 is still fetchable. Correct copying additionally requires its config/layers to remain fetchable and valid.
- **World 2 — source manifest unavailable:** GET application@D0 returns 404. The brief removes that repository’s original manifest, not necessarily its blobs, other repository copies, or all original image data. This is manifest unavailability, not an observed garbage-collection event.

The registry API addresses manifests by repository name plus tag/digest. A surviving release@D0 in another repository would not automatically satisfy a request for application@D0. The profiles contain no alternative source URL or restore task. The minimum sets below are restricted to these unchanged tasks and URLs; a redesigned client able to copy a surviving release@D0 could have different minima.

## Locked minimum earlier-task reselections

These sets exclude failed promote, which must always be attempted again. They optimize **correct delivery**, not mere native success or restoration of every intermediate release tag.

| Profile | World | Minimum earlier tasks | Required execution sequence | Correct-delivery sufficiency |
|---|---|---|---|---|
| tagged | 1: D0 available | {retag} | retag -> promote | Copies existing application@D0 back to release:latest, then to delivery |
| tagged | 2: D0 unavailable | {build, retag} | build -> retag -> promote | Rebuild emits available D1; retag and promote use that new image; conditional on B1=B0 and E1=E0 |
| digest | 1: D0 available | empty set | promote | Promote directly copies application@D0; the mutated release tag is irrelevant |
| digest | 2: D0 unavailable | {build} | build -> promote | Rebuild emits available D1; promote consumes the new result; conditional on B1=B0 and E1=E0 |

**Clone is not necessary in any of these stated worlds** because C0 is assumed present, complete, unchanged, and usable. The worlds alter registry state, not Workspace source bytes. Clone would be needed under a different admitted condition such as missing/corrupted source, which is not the specified intervention.

All added tasks are strict ancestors of promote. Retag is an execution ancestor in the digest profile but not a necessary semantic predecessor for delivery: it writes a release reference that digest promotion does not consume. An inherited retag task may therefore remain omitted while a rerun build orders before promote. Scheduling order alone does not require repeating every intermediate completed task. There are no parallel pairs in either profile.

### Necessity and sufficiency arguments

**Tagged/world 1.** Without retag, promote’s fixed URL resolves A. Neither clone nor build writes release:latest; retag is the only allowed earlier task that does. Thus retag is necessary for correct delivery. Its source is the available D0, so retag followed by promote is sufficient under successful complete copies and absence of races. Build is unnecessary because D0 already supplies the desired bytes.

**Tagged/world 2.** Retag alone requests application@D0 and cannot fetch its missing manifest. Of the supplied tasks, only build can republish an application-repository image. Build is therefore necessary. Build alone changes application:build and records D1, but leaves release:latest=A; promote would still copy A. Retag is also necessary. Build followed by retag with the fresh build result, then promote, is sufficient if the rebuilt application retains the desired identity and all registry operations complete. Adding clone cannot replace either role.

**Digest/world 1.** Promote addresses D0 directly, so no earlier task needs to repeat. It does not consult release:latest. Available verified D0 and all its referenced content are sufficient for the desired delivery; the empty set is the minimum.

**Digest/world 2.** Promote cannot use the unavailable application@D0. Retag cannot recreate that application repository source and would itself fail fetching D0. Clone only recreates already retained checkout bytes. Build is the only allowed earlier task that can produce a new available source image. It is necessary. Promote must use build’s new IMAGE_DIGEST D1, not an inherited/frozen D0. Then build plus promote is sufficient under the rebuild-identity and registry assumptions. Retag need not repeat because neither the delivery predicate nor promote’s source URL depends on release:latest.

These are **conditional sufficient minima**, not an unconditional proof of identical rebuilt binary output from the supplied input pins. The necessity arguments are fixed by the task roles and source URLs. The reproducible-identity assumption in worlds requiring build is an additional premise, discussed below.

## Native success versus actual delivery: predictions

A “native success” here means the retained workflow tasks complete successfully and Tekton reports success, assuming supported syntax, valid bindings, successful image startup, and restored registry access. It does not establish that the desired application was delivered.

| Profile/world | Empty earlier selection | Minimum correct-delivery selection | Full restart |
|---|---|---|---|
| tagged/world 1 | promote executes; native success expected; copies A; incorrect application delivery | retag+promote, 2 tasks; native success and correct delivery expected | clone+build+retag+promote, 4 tasks; native success expected; correct delivery conditional on rebuild identity |
| tagged/world 2 | promote executes; native success expected; copies A despite missing application@D0; incorrect delivery | build+retag+promote, 3 tasks; native success expected; correct delivery conditional on rebuild identity | 4 tasks; native success expected; correct delivery conditional on rebuild identity |
| digest/world 1 | promote executes; native success and correct delivery expected, 1 task | same empty earlier selection, 1 task | 4 tasks; native success expected; correct delivery conditional on rebuild identity |
| digest/world 2 | promote attempts unavailable application@D0; nonzero copy exit/native failure expected; delivery remains A | build+promote, 2 tasks; native success expected; correct delivery conditional on rebuild identity | 4 tasks; native success expected; correct delivery conditional on rebuild identity |

Counts are attempted **workflow tasks in one recovery**, not subprocesses/steps, original-source executions, or measured time/cost. There are no required Pipeline exported-result publishers in these fixed profiles. Actual generator helper counts, if any, must be verified later rather than assumed from unseen implementation outputs. Full restart is not an unconditional binary-identity oracle when build inputs are incompletely pinned.

Diagnostic non-minimum selections also follow prospectively:

- Tagged/world 1, {build} alone: native success may occur, but the mutated release tag still supplies A. A fresh application build does not itself repair the promotion reference.
- Tagged/world 2, {retag} alone: retag fails on missing application@D0 before promote executes. The candidate sequence has two selected tasks but only the failed retag is attempted.
- Digest/world 2, {retag} alone: the same missing-source failure; it cannot substitute for build.
- Digest/world 2, {build}: delivery can be correct while release:latest stays A. A checker that insists every intermediate tag be repaired would be enforcing a different goal and would change this minimum.

A fixed conservative earlier selection covering both worlds can repeat build+retag for tagged promotion and build for digest promotion. That repeats unnecessary work in world 1. These are source-derived conservative sets, not a reading or endorsement of an actual policy. The profiles and external-world state are sufficient to derive the information/work difference without executing a planner.

## Immutable digest binding: guarantee and boundary

Registry API content digests are content-addressed identifiers; retrieval uses `/v2/<name>/manifests/<reference>` with a tag or digest. Blob/config references also have digests. Under honest digest verification and collision-resistance assumptions, D0 constrains the source content to the originally identified manifest and its verified referenced bytes. A mutable tag resolves whatever manifest is bound to it at request time.

Therefore digest binding prevents the specified **release-tag substitution** from silently selecting A. It does not guarantee:

- availability of application@D0 or its config/layers;
- retention/immutability of a repository’s links or permission to pull/push;
- automatic fallback to another repository or a hidden copy;
- application safety, authenticity of the original build, absence of vulnerabilities, or correct source-to-binary provenance;
- absence of races on delivery:latest or an unchanged delivery after the copy;
- exact semantic identity of a new rebuild that emits D1;
- real deployment/execution of the image merely because registry promotion succeeds.

Saved digest identity plus missing bytes should produce failure, rather than a successful substitution of different content, assuming the copy implementation enforces source digest checks. The API’s digest-verification requirement is normative; actual implementation behavior must be checked in the future run. No security-system evaluation is inferred from disabling TLS in this isolated registry composition.

## Identity and availability assumptions requiring evidence

### 1. Runtime pinning is not complete build-input pinning

The task runtime images in these profiles are digest-addressed: git-clone’s configured gitInitImage; Kaniko executor; bash write-url helper; and Skopeo. The registry image and unrelated Alpine reference also have supplied runtime-image pins. These declarations identify intended images; future Pod/container observations are needed to establish which platform images actually ran.

The public Dockerfile remains unchanged:

```dockerfile
ARG BASE_PREFIX=mirror.gcr.io/library/
FROM ${BASE_PREFIX}golang:1.18 as builder
...
RUN go build -gcflags="${SKAFFOLD_GO_GCFLAGS}" -trimpath -o /app main.go
FROM ${BASE_PREFIX}alpine:3
ENV GOTRACEBACK=single
CMD ["./app"]
COPY --from=builder /app .
```

Its **builder and final base images are mutable tag references**, not digest pins in the supplied inputs. No BASE_PREFIX or SKAFFOLD_GO_GCFLAGS override appears in the profile. Pinning Kaniko’s executor does not pin the compiler/standard library or final image filesystem/configuration it downloads. `--reproducible` and `-trimpath` help suppress sources of variability but do not establish complete input identity or determinism.

For rebuild-based correct-delivery predictions, require unchanged resolved Go/Alpine base images, target platform/toolchain, relevant build flags/environment, and complete source bytes, or demonstrate directly that B1=B0 and E1=E0. Native build success alone cannot close this premise. Even full restart may deliver a different binary with a new compiler or changed effective configuration.

If the desired predicate is **exact original manifest digest D0**, rather than the brief’s binary/config identity, a successful rebuild with D1≠D0 is not sufficient. The assessment explicitly adopts the brief’s semantic-identity goal.

### 2. Source Workspace persistence and immutability

The profile has Workspace declarations but no PipelineRun binding. The brief supplies fixed retained source state. The future native input must show how that state is retained/copied, that examples/getting-started source files match the pinned checkout, and that Kaniko can read the same context after clone is inherited. Scratch output and base extraction must not corrupt the retained source. A fresh empty Workspace would invalidate the build-only minima and require clone as well.

The original successful clone’s recorded commit is useful provenance, but it does not prove current Workspace bytes are unchanged. Source snapshot hashes and actual retained files are needed.

### 3. Availability and repository scope

World 1 needs both manifest and all content that Skopeo will read at application@D0, not merely HEAD/GET 200 for the manifest. World 2 requires demonstrated 404 for that exact application-repository digest with compatible Accept headers. Manifest removal is distinct from blob deletion and garbage collection. The API explicitly warns that deleted referenced layers make complete images unresolvable.

Rebuilds require base-image availability plus write access to application:build, and copy operations require source read/destination write access with sufficient storage and compatible media types. Source and delivery endpoints must remain usable throughout each recovery. No task-level retry guarantee or exact HTTP trace is inferred from the bodies alone.

### 4. Concurrency and partial effects

Exclude concurrent mutations of the source checkout, release:latest, application references, and delivery:latest. Tagged/world 1 sufficiency requires that release is not changed again between retag and promote. Digest selection removes that particular source-tag race but delivery remains mutable. Restore each arm’s external state identically before executing it; administrative restoration must not occur during an arm.

The failed promote may have transferred some blobs before the outage. That does not by itself imply a valid delivery manifest/tag. Initial delivery=A must be verified rather than inferred from the failed native condition. Subsequent copy may reuse correct already-uploaded blobs; this does not alter the task-level minima.

### 5. Result inheritance versus result refresh

When build is inherited, digest sources must use saved D0. When build reruns, retag and digest promote must consume the new result D1 and be ordered after that producer. Freezing D0 despite rerunning build would fail world 2 unless rebuilding happened to recreate exactly D0. Inspect future generated manifests/native resolved parameters to distinguish these cases. Both fixed profiles have a string IMAGE_DIGEST reference; array EXTRA_ARGS is an input parameter, not an inherited array result.

### 6. Execution configuration/platform

The application image need not share the manifest digest across rebuilding/format conversion. But equality of only the source revision, a “Hello world!” line, or the presence of an executable file is weaker than the stated identity goal. Record the original binary’s hash and executable mode and the actual execution-relevant config, including CMD/Entrypoint/WorkingDir/environment/User and OS/architecture. Skopeo default multiArch=system selects the running system’s platform, so platform selection must be stable and explicit in the observations.

## Prospective external delivery predicate and evidence needed

Before mutating the failed source’s registries, retain an independent fingerprint of the actual successful build output: expected binary SHA-256 B0_hash, executable path/mode, and E0 from its manifest/config/layers. Tie it to the actual clone commit and source-byte hashes. Do not derive the ground truth from a planner’s predicted action or from whichever image appears at delivery after recovery.

A sufficient checker for the stated goal should:

1. Read the actual delivery:latest manifest with appropriate media-type negotiation and verify its content digest.
2. Fetch/verify its config and referenced layers. Reconstruct the effective filesystem in layer order, including overwrites/whiteouts; locating an old `/app` in any individual layer is insufficient.
3. Require the effective application executable bytes and mode to equal the original B0 identity and require E_delivery=E0 under an explicitly specified configuration comparison. The fixed CMD is `["./app"]` and GOTRACEBACK is `single`; actual inherited base configuration determines the remaining effective fields.
4. Separately report native PipelineRun/TaskRun success and semantic-delivery validity. This separation should expose the tagged empty-selection silent misdelivery rather than classify copy success as application success.
5. Record source D0 availability, release/delivery initial tag identities, actual delivered identity, and each rebuild’s D1. Capture blobs/config/layers before relevant manifest removal so the baseline is not lost.

Executable running can additionally check that the expected binary starts under that configuration, but a log containing “Hello world!” alone is not binary-content identity. Registry-layer/config verification is an external oracle for **artifact delivery**, not evidence that GKE or any production workload was deployed.

Additional native-composition evidence: complete immutable source PipelineRun/TaskRun captures and saved results; current fixed profile specs; actual source/copy Workspace bindings; generated/retained task names and result references; native resolved source/destination URLs; terminal conditions; actual runner/base/platform identities; restored initial states; and destination state after each arm. Declared versus equally informed manual versus restart comparisons can then be assessed without conflating reference repair, rebuild availability repair, and semantic identity.

## Locked conclusion

For the fixed profiles and stated delivery goal, the conditional minimum earlier sets are:

```text
tagged / world 1: {retag}
tagged / world 2: {build, retag}
digest / world 1: {}
digest / world 2: {build}
```

Empty tagged promotion is expected to succeed natively and deliver the unrelated Alpine image in both worlds. Empty digest promotion is expected to deliver the original application in world 1 and fail on unavailable source content in world 2. Clone is unnecessary when the stipulated source Workspace is retained intact. Rebuild-based delivery and full-restart correctness remain conditional on actual binary/config identity; task-runtime pins alone do not establish that condition. Digest binding supplies content selection, not content availability.

These predictions are locked before access to policies, a harness, native failures, recovery arms, or revised manuscript evidence. The next review should compare actual observations to these prospective statements and preserve any mismatch rather than revise the locked expectations to fit the result.


---

## Final prospective offline-transport confirmation (supersedes earlier transport bindings)

The final packet at https://artifacts.internalmeta.com/a/2479895562538338 was downloaded separately, before inspecting any policies, plans, harness, source/recovery outcomes, or revised manuscript. The supplied statement that no source or recovery PipelineRun has executed remains a chronology premise rather than an independently observed execution status. This is an **automated independent source assessment, not a practitioner authoring study**.

All **14/14** file hashes listed in the final SHA256.json reproduce. Its own SHA-256 is `04237d8cf20f8eb3a721cea1ff91c7c529822db5b8a03a24aa31d8f37fe9b245`. The final changed/new file hashes are:

| File | SHA-256 |
|---|---|
| SOURCE_BRIEF.md | `de0cd8f9f2c2b05a2890f1013fe7451b75e2f771d02c4a7abe01c0537f670b7b` |
| profile-tagged.json | `1c780672fac4c634d7d6dd5a1f0cf78a18f2f06d76e27f6d3051d61176f1ca56` |
| profile-digest.json | `282e42fe3b77650b1d1acce0e9b9e2fd2d351fd4e055ea96f9cc3c32252f93c4` |
| offline-transports.json | `53ede5a271053b3314fbbe36d5cbb2b871d51d3a7f5d8d3756f440ed122ffa9d` |

Relative to the preceding frozen-base packet, changes are exactly these four files plus SHA256.json. All other files, including public task/application/API bodies, pins.json, and runtime-images.json, are byte-identical. Recursive comparison of each profile's effective spec finds exactly one change: clone.params.url becomes `git://git-source:9418/skaffold.git`. Every embedded taskSpec remains identical. The only effective-spec difference between tagged and digest profiles is still promote.params.srcImageURL.

The read-only Git daemon is assumed to serve pinned public commit `035e3ddc3589c1c64cb69c563b178f34b1cccc5e`; clone's URL result now identifies that local transport. The final brief additionally specifies pod fsGroup 65532 for workspace access. Availability, read-only behavior, permissions, returned commit, and actual checkout bytes require later native evidence. This transport change introduces no additional earlier-task necessity under the fixed intact source-workspace premise.

The offline record's Go/Alpine source digests agree with the already locked runtime pins and record linux/amd64 configurations and rootfs layer identities. It records Git and three image archive checksums, sizes for image archives, and docker-archive transport names. Those archives and raw destination mirror captures are not included in this packet; their byte checksums and source-to-mirror equivalence have **not** been independently recomputed here. The coordinator's successful verification statement is therefore retained as a supplied premise. Packet-file integrity and internal consistency have been verified independently. Later evidence should expose actual local Git commit/checkout identity and effective mirror configuration/layer identities, with transport conversion or recompression explained where manifest or compressed-layer digests differ.

All four minimal earlier rerun sets remain locked: tagged/world1 `{retag}`, tagged/world2 `{build, retag}`, digest/world1 `{}`, digest/world2 `{build}`. The prior native-success predictions, semantic delivery predicate, frozen-base assumption, result-rebinding requirements, and binary/configuration identity qualifications remain in force. Public-network clone and image-pull latency is excluded; later timing can support conclusions about this local frozen environment only. No native experiment was run and no venue verdict is inferred. Await actual results and the revised manuscript.


---

## Post-calibration explicit Kaniko entrypoint qualification

This bounded assessment responds to the coordinator's subsequent report that the first attempted source completed clone but reached build PodCreationFailed when Tekton's controller could not access gcr.io to discover the image entrypoint. It also reports that no build or recovery executed in that calibration attempt, that evaluated sources use distinct `source-evaluated-{profile}` names, and that native execution has now started. These are supplied facts pending the promised raw calibration and image-configuration evidence. No evaluated recovery outcome has been supplied or inspected here.

The previously inspected Catalog Kaniko task's first step sets image and args but omits command. The reported evaluated adaptation adds `build.taskSpec.steps[0].command: ["/kaniko/executor"]`, said to equal the same pinned image's Docker-inspected Entrypoint. Under Kubernetes container invocation semantics, explicitly supplying the same entrypoint while preserving the existing args invokes the same executor with the same inputs. It removes this controller-side need to discover that entrypoint through the public registry. It does not establish registry availability, image-pull success, build success, or correct delivery.

**The four previously derived minimum earlier rerun sets are unchanged, conditional on this being the sole effective task-spec change and the pinned image's effective entrypoint actually being exactly that command:** tagged/world1 `{retag}`, tagged/world2 `{build, retag}`, digest/world1 `{}`, digest/world2 `{build}`. The command adds no producer, consumer, state mutation, result, or dependency beyond the existing build operation. Only build still publishes the application source image; only retag still repairs the mutable release locator. Necessity and sufficiency arguments for the intended captured state—clone/build/retag completed, promote failed—therefore remain the same, with the existing availability and binary/configuration-identity qualifications.

These four minima **do not describe recovery from the retained calibration failure**: its build never completed and its source state lacks the assumed successful build/retag outputs. The calibration attempt should remain separately identified rather than counted as an evaluated promote-failure source or successful build observation.

This command binding is a **post-calibration adaptation**. It was not present in the locked prospective profiles and must not be called preregistered or described as a byte-identical embedded Catalog task body. The original semantic minima were locked beforehand; this qualification is written after native execution began, without evaluated outcome inspection. Earlier claims that embedded task bodies were unchanged apply to those earlier packets, not to the newly adapted evaluated profiles.

Final evidence should permit comparison of both evaluated profile specs against the locked transport packet, confirming that command is the sole effective task-spec delta, and supply the pinned image configuration showing Entrypoint together with Cmd. It should retain the raw calibration source/TaskRun and PodCreationFailed evidence, distinguish evaluated source identities, and verify that captured evaluated states actually satisfy the original promote-failure premises. The public-registry entrypoint-discovery failure is a concrete offline execution prerequisite; this assessment does not generalize it into a retry-policy failure or infer a venue verdict. No experiments were run or contacts made. Await full results and the revised manuscript.


---

## Second calibration and revised builder binding: independent bounded qualification

Re-downloaded the updated packet into a separate directory. **14/14** listed file hashes reproduce. The final SHA256.json itself hashes to `7fc07793ad424bbcb5a326df6c20d2769bf894fdba60ac1261aebbce5f5cf1b3`. Changed files relative to the previously verified offline-transport packet are exactly SHA256.json, SOURCE_BRIEF.md, and the two profiles:

| File | SHA-256 |
|---|---|
| SOURCE_BRIEF.md | `ebca8caddc85df1edc10296f72f839ee6a581a9094143f57a213cc504823f09b` |
| profile-tagged.json | `fa3bd1cd3c8b19c1183e72db8bbdd44a1f4586c358764140bd26c267ce376037` |
| profile-digest.json | `7b74739b7b1cbdd7edc112670bef271903ba2c2794cd72864b249cb3d94628ef` |

Effective-spec comparison finds exactly the previously reported explicit first build-step command `[/kaniko/executor]`, a build.params.BUILDER_IMAGE override to `gcr.io/kaniko-project/executor@sha256:9e69fd4330ec887829c780f5126dd80edc663df6def362cd22e79bcdf00ac53f`, and one additional EXTRA_ARGS element `--ignore-path=/product_uuid`. Catalog's existing parameter interfaces support passing a builder image and additional arguments; the implementation's precise treatment of the ignore-path option is not independently verified from this packet. Task scripts and ordinary args, application bytes, Dockerfile, other task bodies, ordering, workspaces, and promotion locators are unchanged. The sole effective-spec distinction between tagged and digest profiles remains promote.params.srcImageURL.

**The four semantic minimum earlier rerun sets remain unchanged:** tagged/world1 `{retag}`, tagged/world2 `{build, retag}`, digest/world1 `{}`, digest/world2 `{build}`. Their necessity derives from which existing task can repair the locator or recreate the application source image. The new builder does not add another writer of either registry repository or alter the dependency/results interfaces. Sufficiency remains conditional on a successful final-builder source capture at the intended promote failure, a working final-builder rebuild where required, fixed source/base inputs, current build-result substitution, available required blobs, and delivery binary/configuration identity. The new executor is a different implementation version; unchanged Dockerfile and scripts alone do not prove identical output. No assertion of byte-identical images or cross-version build equivalence is made.

The supplied report says the second calibration compiled the application but Kaniko 1.5.1 failed during cleanup at /product_uuid, published no image, and ran no recovery arm. Raw evidence for this chronology and behavior has not yet been supplied. This calibration is distinct from the assumed source with successful build/retag and failed promote, so the locked promote-recovery minima do not apply to it. Compilation alone is neither a successful build task nor a published application result. Keep both calibration failures separate from evaluated sources and recovery denominators.

The /product_uuid exclusion is acceptable to the preceding reasoning only insofar as it concerns an injected runtime file outside the application source and desired runtime behavior. Later evidence must verify that it does not suppress a required application artifact or change the chosen semantic oracle. The original binary/configuration baseline should come from the actual successful source built with the final pinned executor, not from the unpublished calibration compilation. Every build-containing comparator must use the same final builder/args/command and frozen inputs; retain actual rebuilt binary and execution-configuration checks.

The final profiles and brief record the new digest, but **runtime-images.json is unchanged and still lists the old 1.5.1 executor rather than this new image**. Thus packet integrity verifies the declared override, not the actual new image's identity, version, entrypoint, or transport. The final evidence should provide the new pinned image configuration and actual Pod/container image identity, update or clearly label the historical inventory, preserve calibration records, and expose the complete evaluated profile/spec hashes. The brief's opening statement that inputs precede any failed source is superseded by its explicit post-calibration paragraph and this chronology qualification.

The source-only semantic predictions were locked before attempted runs; these builder and ignore-path bindings are **post-calibration adaptations, not preregistration**. The current assessment was written after source attempts began and without access to evaluated recovery outcomes, plans, policies, harness, or a revised manuscript. This remains an automated independent source assessment, not a practitioner authoring study. No experiments were executed, contacts made, or venue verdict inferred. Await full evidence and the revised manuscript.
