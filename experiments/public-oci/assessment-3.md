# Locked prospective assessment: public Catalog build/registry-promotion application

**Version 5 — post-calibration Kaniko 1.23.2 binding qualification.** Versions 1–3 recorded source-only minima before native execution. Version 2 incorporated the frozen-base mirror premise; Version 3 confirmed offline transport bindings. Version 4 appended an entrypoint qualification after calibration/native execution began. Version 5 verifies the newly supplied evaluated profile hashes and qualifies the subsequent builder/ignore-path change below. The earlier pre-execution input hashes describe the Version 3 packet and do not identify this amended evaluated profile. Original-image identity, mirror stability and availability remain explicit, testable premises. Earlier versions remain in the artifact history.

## Assessment status and inputs

This is an automated independent source assessment, not a practitioner authoring study or a venue recommendation. The predictions are derived from supplied task/application/API contracts before access to failed-source captures, policies, recovery plans, harness code, or execution outcomes. No other assessment or review was consulted for this case. No native experiment was run and no application source was edited.

Input artifact: https://artifacts.internalmeta.com/a/2479895562538338 . All 14 files listed in the final `SHA256.json` reproduce their listed SHA-256 hashes. The two profile hashes are:

- `profile-tagged.json`: `1c780672fac4c634d7d6dd5a1f0cf78a18f2f06d76e27f6d3051d61176f1ca56`.
- `profile-digest.json`: `282e42fe3b77650b1d1acce0e9b9e2fd2d351fd4e055ea96f9cc3c32252f93c4`.

Final SOURCE_BRIEF.md hash: `de0cd8f9f2c2b05a2890f1013fe7451b75e2f771d02c4a7abe01c0537f670b7b`. Final runtime-images.json hash: `1d86f8d2aabc2a76da043cdd81cb4c5e05b3058c953a1861d6a0542ee9294aed`. Version 2 changed the base-input bindings. Version 3 changes the two profiles only at clone’s URL, updates SOURCE_BRIEF.md/SHA256.json, and adds offline-transports.json; runtime-images.json is unchanged from Version 2. Catalog task bodies, application files, public pipeline pattern, and API source are unchanged. Both profile specs still differ only at promote’s source URL.

The inputs include the source brief, both complete resolved profiles, complete git-clone/Kaniko/Skopeo-copy Catalog bodies, the public build/push/GKE pattern, the Skaffold Dockerfile/main.go/go.mod, image pins, and the pinned Distribution API text. Public provenance is specified as Catalog commit `112374ce1262b133a77300bd8133445722605e04`, Skaffold commit `035e3ddc3589c1c64cb69c563b178f34b1cccc5e`, and registry API commit `4772604ae973031ab32dd9805a4bccf61d94909f`. Hash checking authenticates consistency with the supplied inventory; it does not independently certify that the supplied files were fetched from those remote commits.

This assessment uses the supplied task bodies and fixed profiles as the execution contract. A field-level comparison shows identical Catalog scripts, arguments, descriptions, volumes and workspace declarations. Differences are the Skopeo runtime-image digest pin plus native serialization/default fields: empty metadata/computeResources, null spec, and explicit string types where the Catalog omits the default type. Clone’s pinned git-init image is a parameter value; the Kaniko executor and write-url images are already digest-pinned in the Catalog body. GKE deployment is excluded; registry promotion is an authored integration of public components.

## The resolved application contract

The original pipeline order is `clone -> build -> retag -> promote`. All three earlier tasks have succeeded when promote fails because the delivery endpoint is unavailable. Recovery occurs after that endpoint is restored, with the original source observations/results and an intact fixed source workspace.

The two profile specs differ in exactly one field: promote’s `srcImageURL`.

| Task | Relevant supplied behavior |
|---|---|
| clone | Fetches the pinned Skaffold revision with submodules disabled and sparse checkout `examples/getting-started`; writes commit, URL and committer-date results. `deleteExisting` defaults true and the script cleans the output directory before cloning. |
| build | Reads `examples/getting-started/Dockerfile` and that context from the source workspace; Kaniko pushes `registry-source:5000/application:build` and writes `IMAGE_DIGEST`. Final extra arguments are `--insecure`, `--insecure-pull`, `--cache=false`, `--reproducible`, `--force`, and `--build-arg=BASE_PREFIX=registry-source:5000/base/`. A second step publishes IMAGE_URL. |
| retag | Skopeo copies `docker://registry-source:5000/application@$(tasks.build.results.IMAGE_DIGEST)` to `docker://registry-source:5000/release:latest`. |
| promote, tagged | Skopeo copies `docker://registry-source:5000/release:latest` to `docker://registry-delivery:5000/delivery:latest`. |
| promote, digest | Skopeo copies `docker://registry-source:5000/application@$(tasks.build.results.IMAGE_DIGEST)` directly to the same delivery destination. |

Both copy tasks provide nonempty source/destination parameters, so they take the single-image Skopeo branch, not the `url.txt` loop. They still require their declared workspace bindings for native validation, but that branch does not read `url.txt`. `multiArch` defaults to `system`, and source/destination TLS verification is disabled in the fixed profiles.

The direct branch’s final command is `skopeo copy`; its unsuccessful exit is not replaced by a later unconditional success command. Successful native copying means the specified reference was copied, not that it was the intended application. The task has no assertion comparing the delivered binary or execution configuration with the original build.

Define:

- `D0`: the original successful build’s saved IMAGE_DIGEST.
- `B0`: the original built application binary bytes.
- `E0`: the original image’s relevant executable configuration, fixed by an oracle before recovery outcomes.
- `A`: the unrelated Alpine image now selected by `release:latest` and initially present at the delivery destination.
- `D1`: a new successful build’s result if build repeats; this may equal D0 or may differ.

**World 1 — drift with source manifest available:** release selects A, while GET of `application@D0` remains available.

**World 2 — drift with source manifest unavailable:** release selects A and GET of `application@D0` returns 404. This is the specified manifest-unavailability intervention. It does not establish actual registry garbage collection or deletion of every copy of the image.

The scoped delivery objective is B0 with E0 in the delivery image. It does not require restoring every intermediate release tag. That distinction matters for digest-profile recovery, which can deliver correctly while leaving release:latest pointing to A.

## Smallest sufficient earlier-task selections

Promote must run in every recovery and is excluded from “earlier reruns.” Counts below are predicted pipeline-task executions, not step/container counts or measured cost. Build has two steps but is one task. The profiles declare no Pipeline results needing extra publication helpers.

**Identity qualification:** the sets below are the smallest repair sets under intact source-workspace, usable registries/blobs, and fixed build-input/platform assumptions. The final packet now explicitly supplies a frozen mirror of the original digest-pinned Dockerfile bases, available in both worlds. Under that premise and repeatable compilation/configuration, selections containing build are expected to reproduce B0/E0. These are prospective predictions to verify against actual bytes/configuration, not observed equality or an unconditional reproducibility theorem. Full restart has the same operational premises.

| Profile | World | Minimum earlier reruns | Retry execution sequence | Tasks | Native outcome under availability assumptions | Delivery outcome |
|---|---|---|---|---:|---|---|
| tagged | 1 | {retag} | retag, promote | 2 | success | B0/E0, using retained original image |
| tagged | 2 | {build, retag} | build, retag, promote | 3 | success | B0/E0 if rebuilding preserves application identity |
| digest | 1 | empty set | promote | 1 | success | B0/E0, using retained original image |
| digest | 2 | {build} | build, promote | 2 | success | B0/E0 if rebuilding preserves application identity |

### Necessity and sufficiency: tagged, World 1

Without retag, promote reads the changed release tag and copies A. Repeating clone or build alone does not change that release tag: build’s destination is a different repository/reference. Retag is therefore necessary for correct delivery using these task actions.

Retag alone can retrieve the retained `application@D0`, copy it to release:latest, and let promote copy that release to delivery. It is sufficient if the original config/layers remain retrievable and the release reference stays stable between retag and promote. Clone and build add no necessary repair under this world’s stated conditions.

### Necessity and sufficiency: tagged, World 2

Retag without build still uses the inherited D0 reference, whose source manifest is absent, and fails. Without retag, repeating build does not replace the unrelated release image, so promote still copies A. Both build and retag are necessary.

Build can reconstruct/push an image from the intact workspace and produce D1. Retag must consume that new build result, not a frozen D0, and set release:latest to the new image. Promote then copies that reference to delivery. This is a sufficient state-repair sequence; matching B0/E0 additionally requires stable/equivalent build inputs and executable configuration. Clone is unnecessary when the fixed workspace still contains the original correct checkout.

### Necessity and sufficiency: digest, World 1

Promote already addresses the retained original image by D0. The changed release tag is not an input to its copy. Zero earlier repeats is sufficient and is the minimum possible set. Retag is an execution predecessor in the original graph, but repeating it is not a delivery-repair prerequisite in this world.

### Necessity and sufficiency: digest, World 2

Promote cannot retrieve `application@D0`. Clone does not publish an image; retag also requires that absent source reference and only writes the release repository. Among the earlier workflow actions, build is the one capable of reconstructing/publishing a source image and supplying a usable result. Thus build is necessary.

After build produces D1, promote must use `application@D1` and can copy it directly. Retag need not run: promote’s source bypasses release:latest. The needed scheduling is build before promote. If a recovery generator removes retag, it must preserve that ordering and the new result reference. The original reference to build’s IMAGE_DIGEST already supplies a direct result dependency in this profile. Matching B0/E0 remains conditional on rebuild identity.

These minima are relative to the supplied task actions and the delivery predicate. Administrative manifest restoration, changing copy parameters to use another repository, retrieving an original image from delivery storage, or modifying the Dockerfile are different actions outside the assessed recovery interface. The possibility that another repository still has an original digest does not make retag sufficient: its fixed source URL names the application repository.

## Empty selection and full restart

“Empty” retains only failed promote. “Full restart” executes clone/build/retag/promote in order. Predictions assume the repaired endpoint and required source/base/runtime images are reachable, permissions/bindings are valid, and copy/build tools operate as their contracts describe.

| Profile | World | Empty task count | Empty native outcome | Empty actual delivery | Full task count | Full native outcome | Full actual delivery |
|---|---|---:|---|---|---:|---|---|
| tagged | 1 | 1 | success | incorrect: A copied | 4 | success | correct if rebuild identity holds |
| tagged | 2 | 1 | success | incorrect: A copied | 4 | success | correct if rebuild identity holds |
| digest | 1 | 1 | success | correct: original B0/E0 copied | 4 | success | correct if rebuild identity holds |
| digest | 2 | 1 | failure on absent source manifest | incorrect: desired image not delivered | 4 | success | correct if rebuild identity holds |

For digest/World 2, source resolution fails before a successful desired manifest can be delivered; the existing unrelated delivery image does not become correct. Partial destination uploads or other leftover effects should be checked separately rather than inferred absent from failed native status.

A fixed conservative earlier selection covering both supplied worlds is {build, retag} for tagged and {build} for digest, with the same rebuild qualifications. It repeats unnecessary work in World 1. These are source-derived selections, not an inspected or generated policy. No counts here establish latency, compute, energy, or authoring savings.

Useful negative controls follow directly: tagged/World 2 retag-only fails while fetching D0; build-only can succeed natively yet deliver A because it leaves release unchanged. Digest/World 2 retag-only does not repair the source manifest and fails before promote. These controls distinguish source-availability repair from release-reference repair.

## Identity and availability assumptions

### What digest binding establishes

Distribution’s **Content Digests** section defines content-addressed identifiers and requires clients to verify against the digest used to request content rather than trust Docker-Content-Digest alone. **Pulling an Image Manifest** and **GET Manifest** permit tag or digest references. The manifest’s config/layer references provide the subsequent content chain; blob downloads are separate repository-scoped requests.

Given a trustworthy saved D0, compatible media-type handling, collision-resistant hashes, client integrity verification, and available referenced bytes, `application@D0` selects original content independently of what release:latest now names. The digest profile therefore removes this particular release-tag drift dependency.

It does not retain or recreate absent manifests, make missing blobs available, certify the image’s source provenance, freeze build toolchain inputs, establish application behavior, or protect delivery:latest against a later concurrent overwrite. API digest addressing is scoped by repository name as well as digest. A readable manifest by itself is not proof that every required config/layer blob is available.

### Final frozen-base input contract

The unchanged public Dockerfile writes `FROM ${BASE_PREFIX}golang:1.18` (line 2) and `FROM ${BASE_PREFIX}alpine:3` (line 10). It compiles with `go build -gcflags="${SKAFFOLD_GO_GCFLAGS}" -trimpath -o /app main.go`, sets GOTRACEBACK=single, and uses CMD ["./app"]. Its supplied main.go repeatedly prints “Hello world!”; go.mod introduces no nonstandard module dependency.

The final profiles pass the Dockerfile’s supported argument `--build-arg=BASE_PREFIX=registry-source:5000/base/`, plus `--insecure-pull` for that local registry. Thus the actual FROM references are `registry-source:5000/base/golang:1.18` and `registry-source:5000/base/alpine:3`, not the public mutable tags under the Dockerfile’s default prefix. The refined brief specifies those private aliases as a frozen mirror, kept available in both worlds. Updated runtime-images.json identifies their source content:

- Go 1.18 base: `sha256:50c889275d26f816b5314fc99f55425fa76b18fcaf16af255f5d57f09e1f48da`.
- Alpine 3 base: `sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a`.

Together with the unchanged digest-pinned task runtimes, original source bytes, and stable platform/build arguments, this supplies the missing input-stability premise identified in Version 1. The mirror shares the registry-source endpoint but uses separate base repositories; World 2 removes the original application manifest and explicitly leaves the base mirror available. It therefore does not prevent build from reconstructing the application.

The task passes tag-shaped references to the frozen mirror rather than literal digest-shaped FROM references. Their fixed binding is a supplied experiment contract, not cryptographically enforced by the Dockerfile syntax. It must be verified and kept invariant. This refinement closes the public mutable-base gap for the specified case; it does not make `--reproducible` or `-trimpath` universal guarantees.

For build-containing selections, correct B0/E0 delivery is predicted under:

1. Original intact main.go, go.mod and Dockerfile bytes in the fixed workspace, with no concurrent invalidation.
2. The private aliases serving the specified original base content throughout source and recovery, and unchanged target platform, build arguments, compiler environment and relevant configuration.
3. Repeatable compilation and effective executable configuration under those fixed inputs, followed by an independent comparison of delivered binary bytes, mode/location and E0. New D1 may still differ because image metadata differs; exact manifest-digest equality is not the stated application-identity criterion.

The packet’s base bindings and availability are accepted prospective premises. No actual mirror contents or rebuilt image were queried. Future raw evidence should establish those premises and B0/E0 equality. Repeating clone remains unnecessary under intact-workspace conditions and adds Git availability requirements; it does not replace verification of build inputs.

### Other availability/composition premises

- Fixed source observations must contain the original successful build’s actual D0 and coherent completed-task records. A repeated build must replace result consumption with D1 where selected; inheriting stale D0 after rebuilding is a composition error.
- The source workspace must survive recovery with the correct context. Clone is not necessary under that premise; loss or corruption of the checkout would change the minima and could require clone.
- All required runtime/base images, original or rebuilt config/layer blobs, DNS routes, registry read/write operations, and Git access for full restart must be available. A manifest 404 is not a model of authentication failure or all possible registry outages.
- The images-url workspace still must be legally bound even though the single-image branch does not consume its files. Task service-account, filesystem, architecture, and root/nonroot requirements must be satisfied.
- Multi-architecture images use Skopeo’s `system` selection; the future evidence must record platform/media types. This assessment does not assume all-platform replication or byte-identical destination manifest representation.
- No source deletion, release mutation between retag and tagged promote, or delivery overwrite may race the relevant successful copy and final-state check. Digest input prevents release-name drift from selecting A, but not loss of the referenced object or later delivery mutation.
- Workflow tasks cannot call the administrative restore path. Resetting original manifests between matched arms is an experimental starting-state operation, not an available recovery task.

## Required evidence for a later composition check

The following list is prospective; no such execution observations were read for this assessment.

1. **Source identity:** complete original failed PipelineRun/TaskRuns and saved build result D0; checked-out commit and hashes of main.go, go.mod and Dockerfile; original manifest/config/layer bytes; B0 and the pre-specified E0 projection.
2. **World construction:** immutable per-arm starting snapshots; release:latest resolving to A in both worlds; `application@D0` readable in World 1 and explicitly returning 404 in World 2; documented availability of needed blobs. API **Deleting an Image/DELETE Manifest** states deletion is by digest and a successful operation returns 202, but removal of a repository manifest should be distinguished from garbage collection and from deletion of all other repository copies.
3. **Native actions and arguments:** actual TaskRuns/statuses plus resolved source/destination URLs. For rebuilt runs, retag and/or promote must consume D1 as prescribed. Verify order rather than infer it from requested selection. Check that the original Catalog scripts and fixed profile specs were not edited to force the expected outcomes.
4. **Actual delivery:** resolve delivery:latest immediately after the arm, retrieve and independently verify its manifest/config/blobs, reconstruct the effective image filesystem with layer/whiteout semantics, and compare the delivered executable bytes/mode and E0. A tag name, copy exit code, source commit string, or reported digest alone is not that oracle.
5. **Execution configuration:** define the relevant comparison before outcomes, including command, entrypoint, working directory, environment, user and platform as applicable. The intended `./app` must resolve to the original executable. An optional recorded runtime check can complement byte/config identity; it is not substituted for them by the current source contracts.
6. **Rebuild inputs:** verify the frozen private base aliases against the supplied original base digests and retain build-platform/argument records. Record their availability in both worlds and show that application-manifest deletion does not remove base inputs. Post-run equality of B0/E0 checks the predicted identity under these fixed inputs; it is not implied by source-file hashes alone.
7. **Matched isolation:** no successful recovery arm supplies state to a later arm; reset/restoration occurs before each arm and is separated from task effects. Record any original image still available in another repository, without silently altering the assessed copy sources.
8. **Cost accounting:** count actual task/step executions and any helper tasks separately. Measure useful build/copy work or runtime if benefit is claimed; the source-only counts above are not performance measurements.

## Locked conclusions

The smallest delivery-repair selections are tagged/World 1 {retag}, tagged/World 2 {build, retag}, digest/World 1 empty, and digest/World 2 {build}, under the explicit availability and rebuild-identity premises. Empty tagged recovery can succeed natively while delivering the unrelated image in either world. Empty digest recovery should deliver original content when D0 remains retrievable, and fail when that source manifest is unavailable.

Immutable digest binding addresses selection identity when content is available; it does not establish retention or deterministic rebuilding. The final packet supplies frozen original Dockerfile-base bindings through BASE_PREFIX and keeps them available in both worlds. Correct original-binary delivery after rebuilding is therefore expected under that contract and fixed platform/build inputs, and remains a measured byte/configuration check for later execution. Full restart shares those operational premises. These statements are locked source-derived predictions and conditions for a later evidence check, without a venue verdict.


## Version 3 addendum — final offline-transport confirmation

The final transport packet was re-downloaded from the same input artifact before source/recovery execution. All **14** inventory hashes reproduce. New offline-transports.json has SHA-256 `53ede5a271053b3314fbbe36d5cbb2b871d51d3a7f5d8d3756f440ed122ffa9d`. The current profile and SOURCE_BRIEF hashes are recorded above.

Relative to Version 2, the only profile-spec change in either variant is clone’s `url`: `https://github.com/GoogleContainerTools/skaffold` becomes `git://git-source:9418/skaffold.git`. The revision remains `035e3ddc3589c1c64cb69c563b178f34b1cccc5e`. Task bodies, build arguments including the frozen BASE_PREFIX, task ordering, promotion URLs, application bytes and both worlds are unchanged. Thus all four minimum earlier rerun sets, native-success/delivery predictions and task counts remain as derived. Full restart now requires the local read-only Git service and its pinned commit to be available; recovery reusing clone does not require a fresh Git fetch. The clone URL result will correctly identify the local transport, not the public origin URL; public provenance must be established by commit/content identity and transport records.

The new transport index identifies linux/amd64 bases, including Go 1.18.10 with its environment/configuration, seven Go base rootfs layer identities, and the one-layer Alpine runtime/foreign-image roots. It records a bare-Git archive checksum and Docker-archive checksums for the Go, Alpine-base and unrelated Alpine images. These make the frozen-input premise more concrete. Archive transport or manifest representation can change a registry manifest digest while preserving the intended rootfs/configuration; verify effective layer/configuration identity, not only public and mirrored reference names. The rootfs layer identities in this metadata are not automatically identical to registry compressed-blob digests.

**Verification boundary:** the source-packet file hashes and profile differences were independently checked. The actual Git/image archive payloads and running mirror are not included in this packet and were not fetched or executed for this assessment. The reported upload verification and mirror/source equality therefore remain supplied premises to reconcile with raw transport/effective-config evidence later. No policy, planner output, failed-source capture or recovery outcome was consulted. The stated fsGroup=65532 should be confirmed from the eventual native run/pod configuration, since these input Pipeline profiles alone do not establish that runtime setting.

Public-network clone and image-pull latency is explicitly excluded. Any later timing result characterizes this local/offline transport environment, including whichever local clone/build/copy work is actually timed. It cannot be attributed to avoiding public Git or public registry network latency. Binary/configuration identity, mirror stability, platform compatibility, manifest-versus-blob availability, correct D0/D1 result consumption, arm isolation and post-copy race exclusions retain their existing qualifications.


## Version 4 addendum — explicit Kaniko entrypoint after startup calibration

**Timing/provenance.** The coordinator reports that an initial attempted source completed clone, but build had PodCreationFailed because Tekton could not contact gcr.io to discover the Kaniko image entrypoint. No build step or recovery arm executed in that calibration attempt. The evaluated profile subsequently adds `build.taskSpec.steps[0].command = ["/kaniko/executor"]`; execution of separately named `source-evaluated-{profile}` sources has now begun. This binding is a calibration-driven amendment, not preregistered or part of the pre-execution hash lock. The earlier semantic minima were locked before these observations.

**Conditional semantic confirmation.** If the supplied assertion is verified that the same pinned Kaniko image has Entrypoint exactly `["/kaniko/executor"]`, and image, arguments, scripts, working directory, source workspace, BASE_PREFIX inputs and copy locators are unchanged, making that command explicit launches the same build executable with the same arguments. It removes the controller’s need to discover the entrypoint remotely. It neither adds a registry-writing producer nor changes build’s source/destination/result contract, retag’s source reference, promote’s profile distinction, or either post-promote-failure world.

Consequently, the four previously derived earlier-task minima remain:

- Tagged, retained original source manifest: **{retag}**.
- Tagged, missing original source manifest: **{build, retag}**.
- Digest, retained original source manifest: **empty**.
- Digest, missing original source manifest: **{build}**.

All previously stated delivery-identity, result-consumption, availability, isolation and ordering qualifications remain. Explicit entrypoint resolution makes native startup feasible under the reported controller restriction; it does not itself demonstrate binary/configuration equality or a successful recovery. These minima concern a source where clone/build/retag have actually succeeded and promote has failed. They do not describe the separate calibration source whose build never started.

**Evidence boundary and later checks.** This addendum derives the consequence of the coordinator-supplied command binding; it does not claim independent inspection of the raw calibration, amended profile hashes, or Docker inspect record, which were not supplied here. Final evidence should retain: the calibration PipelineRun/TaskRuns and PodCreationFailed reason; the pinned image’s exact Entrypoint/Cmd and digest; the evaluated profile hashes and explicit command/args; the distinct evaluated source names and their actual task statuses; and the chronology of calibration, amendment and recovery trials. Compare the evaluated source’s saved Pipeline specification against the amended profile, rather than treating the pre-amendment capture or Version 3 hashes as identical. Keep the failed calibration separate from the promote-failure recovery cohort and any reported task-execution or timing denominator.

The explicit command is an adaptation beyond runtime-image pinning/native serialization and should be recorded as such in any later statement about unchanged Catalog task bodies. It is expected to preserve executable behavior if the image configuration equality is established. No evaluated recovery outcome or venue verdict is inferred in this addendum. No experiment was run and no other party was contacted for it.


## Version 5 addendum — Kaniko 1.23.2 and runtime-path exclusion

**Latest packet fingerprints.** The packet at https://artifacts.internalmeta.com/a/2479895562538338 was re-downloaded. All 14 hashes in its SHA256.json match. The latest profile hashes are `fa3bd1cd3c8b19c1183e72db8bbdd44a1f4586c358764140bd26c267ce376037` (tagged) and `7b74739b7b1cbdd7edc112670bef271903ba2c2794cd72864b249cb3d94628ef` (digest). Latest SOURCE_BRIEF.md hash is `ebca8caddc85df1edc10296f72f839ee6a581a9094143f57a213cc504823f09b`. Earlier fingerprints elsewhere in this document identify their historical packets, not these evaluated profiles.

**Independently checked changes.** Relative to the Version 3 transport packet, each profile adds:

- A build `BUILDER_IMAGE` parameter value `gcr.io/kaniko-project/executor@sha256:9e69fd4330ec887829c780f5126dd80edc663df6def362cd22e79bcdf00ac53f`, identified by the source brief as Kaniko 1.23.2.
- Build EXTRA_ARGS `--ignore-path=/product_uuid`, retaining the existing frozen BASE_PREFIX and other build flags.
- `build.taskSpec.steps[0].command=["/kaniko/executor"]`, the entrypoint amendment previously discussed in Version 4.

The Catalog YAML, Dockerfile, Go application, registry API, frozen-base/source-transport metadata and other task bodies are unchanged. The changes are parameter binding plus explicit command resolution; they are not an unchanged complete evaluated Task specification. Both variants still differ only at promote’s source locator. The additional runtime-path flag and executor version must be reported as calibration-driven execution adaptations.

**Timing and cohort.** The supplied brief reports two calibration sources: one whose builder pod could not be created, and one which compiled the app with Kaniko 1.5.1 but failed during cleanup at /product_uuid without publishing an image. No recovery arm ran in either calibration. The source-only minima preceded these attempts. The new builder and exclusion flag were selected afterward and are not preregistration. The brief’s initial pre-execution framing must be read together with its explicit post-calibration section; the latest packet as a whole was not locked before the calibration failures.

**Effect on minima.** The builder parameter is already part of the public Task’s interface. Assuming the new pinned executor implements the stated Dockerfile build/push/result contract and the exclusion concerns only the injected /product_uuid runtime file, none of these changes modifies which workflow action repairs the missing application manifest or the drifted release tag. The Dockerfile and application do not reference /product_uuid. The four earlier-task minima therefore remain **{retag}** for tagged/manifest-present, **{build, retag}** for tagged/manifest-missing, **empty** for digest/manifest-present, and **{build}** for digest/manifest-missing. Their native task-count predictions remain 2, 3, 1, and 2 respectively under the existing operational premises.

This is confirmation of the producer/reference dependency argument. It is not a proof that arbitrary Kaniko versions produce equivalent images or have identical snapshot behavior. A runtime-path exclusion could affect image contents outside the scoped binary/configuration predicate and should be checked in raw image evidence. No source-only argument converts a successful build/copy exit into verified original-binary delivery.

**Identity baseline and remaining evidence.** Define D0/B0/E0 from the successful published build preceding promote failure in the **evaluated Kaniko 1.23.2 source**. The unpublished calibration compilation is not that baseline. Keep the executor pin, arguments, frozen bases and platform fixed across that source and its recovery/restart comparisons, and verify rebuilt/delivered executable bytes and configuration against that source. No equality between Kaniko 1.5.1 calibration output and 1.23.2 evaluated output is assumed.

The latest runtime-images.json retains its Version 2 hash and still documents Kaniko 1.5.1; it does not contain an inspect/configuration record for the new executor. The new digest is independently verified as the literal profile binding, but its image payload, claimed version, Entrypoint/Cmd, supported flags and snapshot behavior are not independently established by this packet alone. Final evidence should supply the new pinned image configuration/digest and native resolved command/arguments, raw calibration reasons and publication state, evaluated source identities/specs, and byte/configuration results. Use evaluated source captures matching the final amended profile in strict-definition checks. Separate both calibrations from the promote-failure recovery cohort and disclose any compilation/resource work when reporting overall experimental cost.

No native experiment, planner or policy was executed or consulted for this addendum. No other party was contacted. The unchanged minima are conditional source-contract predictions; actual outcomes and the eventual venue assessment await full evidence.
