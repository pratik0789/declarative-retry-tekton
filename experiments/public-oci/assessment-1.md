# Locked prospective source assessment: public-task OCI promotion — refined packet

## Revision record

This assessment incorporates the refined packet before receiving any native source-failure or recovery outcomes. It supersedes the original packet assessment in this artifact; the original is retained in this artifact's historical revision (assessment SHA-256 `380e34cd4f7e9f3b3c04f2a4bf25b8e9120cf8bbffe6dd4283bcf03e880a0fd1`). The earlier qualification about mutable upstream base tags applies to that original packet. The final packet supplies a frozen private base mirror, so base-tag drift is now excluded by its stated invariant.

## Scope and input identity

This is an **automated independent source assessment, not a practitioner authoring study or a venue verdict**. It predicts behavior from the supplied task/application/registry contracts before receiving any source-failure or recovery evidence for this case. No shared-repository protocol, policies, harness, future outputs, other assessment, or other review was consulted. No source was modified, no native execution was launched, and no person or session was contacted.

Input artifact: https://artifacts.internalmeta.com/a/2479895562538338. All **13 file hashes** listed in `SHA256.json` matched downloaded bytes. The fixed profile hashes are:

- Tagged: `425fdd57fe5498079d3276d024200f85922e2b41b6ab90c522771be3782b0c8b`.
- Digest: `a9da10f4bf1e54f2d9fccb82102e01a99fd65a269614860f03f096fd202ca8f0`.

The package pins Catalog commit `112374ce1262b133a77300bd8133445722605e04`, Skaffold commit `035e3ddc3589c1c64cb69c563b178f34b1cccc5e`, and Registry API document commit `4772604ae973031ab32dd9805a4bccf61d94909f`. Assessment used the supplied copies; independently fetching those upstream commits was not part of this pass.

Final `SOURCE_BRIEF.md` hash: `351674b9c72e91e6bb3a60fa755b6f5bbde2fb05c3fca4a9c6b9450a4e819e6b`; final `runtime-images.json` hash: `1d86f8d2aabc2a76da043cdd81cb4c5e05b3058c953a1861d6a0542ee9294aed`. Compared with the original packet, exactly SOURCE_BRIEF.md, both profiles, runtime-images.json and SHA256.json changed. The entire Pipeline spec changed only in build EXTRA_ARGS; all four embedded Task specs are unchanged.

The two final profiles have identical Pipeline specs except promotion's `srcImageURL`. Their different Kubernetes names/UIDs/resourceVersions/last-applied annotations are not recovery-semantic differences. For both profiles, all four embedded Task bodies match the supplied complete Catalog specs after removing Kubernetes-added empty/default fields and accounting for immutable Skopeo image pinning. Clone's resolved `gitInitImage` parameter is also pinned. Scripts and build commands are otherwise unchanged.

## 1. Relevant existing contracts

Execution is serial **clone → build → retag → promote**.

**Clone:** Catalog git-clone checks out the named Skaffold commit, disables submodules, and uses the application sparse checkout. Its default `deleteExisting=true` cleans the checkout before cloning. It reports the actual commit via `git rev-parse HEAD`. It writes to the `source` Workspace.

**Build:** Catalog Kaniko reads `examples/getting-started/Dockerfile` and context from the `source` Workspace, pushes `registry-source:5000/application:build`, and publishes `IMAGE_DIGEST`. Final extra arguments are `--insecure`, `--insecure-pull`, `--cache=false`, `--reproducible`, `--force`, and `--build-arg=BASE_PREFIX=registry-source:5000/base/`. This uses the unchanged Dockerfile's existing BASE_PREFIX argument. Both admitted worlds keep the private Go/Alpine mirror frozen and available; no repeatable build path needs to fetch mutable upstream base tags under that assumption. Exact original binary/configuration equality is still an execution observation to verify, rather than a result already established by reading task source.

**Retag:** Catalog Skopeo-copy reads `docker://registry-source:5000/application@$(tasks.build.results.IMAGE_DIGEST)` and copies it to `docker://registry-source:5000/release:latest`.

**Promote:** The tagged profile copies `docker://registry-source:5000/release:latest`; the digest profile copies `docker://registry-source:5000/application@$(tasks.build.results.IMAGE_DIGEST)`. Both write `docker://registry-delivery:5000/delivery:latest`. Both use direct nonempty source/destination parameters, so the task's alternate file-driven `copyimages` loop is not used. The direct Skopeo invocation is the terminal command in its branch: a successful copy is expected to yield task success, and a failed copy a nonzero task exit. Neither task checks the source image against the expected application binary.

**Registry API:** `GET /v2/<name>/manifests/<reference>` accepts a tag or digest; an unknown image returns 404. Manifest content identifies further blobs, fetched by digest with repository-scoped access. PUT accepts a manifest under a tag or digest. DELETE removes a manifest by digest. Availability of a manifest and availability of its referenced blobs are distinct; the API explicitly notes that missing referenced layers make an image unresolvable. Content fetched by a digest must be verified against that requested digest, not merely a response header.

The original Catalog build/push/GKE example supplies a pattern, not a completed deployment evaluation here. GKE is excluded; the four-task profile is an authored integration of public components.

## 2. State, objective, and boundaries

Let **d₀** be the successful source build's saved digest, **B₀** its actual compiled `/app` bytes, and **C₀** its execution configuration. Neither d₀ nor B₀ is yet present as an execution observation in these inputs.

At the assumed failed source, clone/build/retag succeeded and promote failed during a delivery-endpoint outage. Before each recovery the endpoint is restored, the destination initially contains unrelated Alpine, source Tekton objects/results are unchanged, and the source Workspace is fixed.

- **World 1 (tag drift, source available):** `release:latest` resolves to unrelated Alpine; `application@d₀` remains available.
- **World 2 (tag drift, source manifest unavailable):** `release:latest` resolves to Alpine; GET of `application@d₀` returns 404.

The objective is **delivery of B₀ with C₀**, not equality of a rebuilt manifest digest and not restoration of every intermediate registry tag. Earlier reruns are selected from the completed tasks `{clone, build, retag}`; failed promote runs after the selected tasks in original order. No task can invoke administrative raw-manifest restoration, alter its fixed source locator, or copy from a different repository as an unlisted recovery operation.

“Sufficient” below is conditional on reachable/configured registries, available referenced blobs, coherent result inheritance, no concurrent tag/writer changes, and—for rebuilding—reproduction of B₀/C₀. The refined packet materially supports the rebuild condition by fixing its base inputs. Under the frozen mirror, fixed checkout, same platform/compiler configuration and ordinary reproducible-build behavior, original B₀/C₀ delivery is the predicted result of the World 2 minimal selections and full restart. Source inspection does not substitute for the later binary/configuration comparison; all sufficiency statements retain these explicit operational/build assumptions.

## 3. Smallest sufficient earlier selections

| Profile | World | Minimum earlier reruns | Execution including promote | Basis |
|---|---|---|---|---|
| Tagged | 1 | `{retag}` | retag → promote; 2 tasks | Repair the mutable source tag from available `application@d₀`. |
| Digest | 1 | `{}` | promote; 1 task | The fixed source digest already selects the desired available image; release-tag drift is irrelevant. |
| Tagged | 2 | `{build, retag}` **conditional** | build → retag → promote; 3 tasks | Recreate an application manifest, then replace the promotion source tag. |
| Digest | 2 | `{build}` **conditional** | build → promote; 2 tasks | Recreate an application manifest and bind promote to the new build result. Retag is irrelevant to this delivery path. |

### Necessity and sufficiency

**Tagged / World 1:** Without retag, promotion reads Alpine. Clone or build alone cannot overwrite `release:latest`; only retag has that destination among the earlier fixed tasks. Retag reads the saved d₀, copies the original image to release, and promotion then reads that repaired tag. Thus `{retag}` is necessary and sufficient when d₀ and all referenced data are accessible and the tag is not changed again between the two copies. Clone/build need not rerun.

**Digest / World 1:** Empty earlier selection is sufficient because promote reads `application@d₀`, not release. No smaller set than empty exists. Retag could repair release for a separate goal, but contributes nothing to the stated delivery predicate.

**Tagged / World 2:** Retag alone attempts to read missing `application@d₀` and cannot produce the desired image. Build is the only earlier task capable of publishing a fresh manifest at the required application repository from the retained checkout; clone alone does not publish anything. Build alone still leaves release pointing to Alpine. Thus build and retag are both necessary. Build emits d₁, retag must consume **the rerun build's d₁**, and promotion reads the repaired tag. This is sufficient to deliver a rebuilt application image. It is sufficient for **original** B₀/C₀ only if the rebuild reproduces that identity.

**Digest / World 2:** Empty selection reads missing d₀. Retag alone also reads missing d₀, and clone alone leaves the required manifest unavailable. Build is necessary, emits a reachable d₁, and promote must use **d₁ rather than inherited d₀**. Build plus promote can deliver the rebuilt application without touching release. Clone is unnecessary if retained source bytes are intact. Retag's stale derived state does not matter to this final-delivery goal; omitting it does not establish consistency of all intermediate effects or full-pipeline semantic equivalence.

Possible retained original manifests under some other repository do not invalidate this necessity argument: these fixed tasks cannot switch to that locator. Administrative restore or an additional copy task would change the permitted operation set.

All required earlier selections are strict ancestors of promote. No parallel behavior is exercised by this serial case. Graph ordering is separate from the application-specific reason to repeat retag or rebuild.

## 4. Native success versus actual delivery

These are predictions, not measured results or population success rates.

| Profile / world | Empty selection | Minimum selection above | Full restart |
|---|---|---|---|
| Tagged / 1 | **Expected native success; wrong delivery (Alpine).** 1 task. | Expected native success; B₀/C₀ delivered. 2 tasks, subject to availability/no-race conditions. | Expected native success; rebuilt application delivered. 4 tasks; original identity conditional. |
| Digest / 1 | **Expected native success; correct B₀/C₀ delivery.** 1 task, subject to original blob availability. | Same as empty. | Expected native success; rebuilt application delivered. 4 tasks; original identity conditional. |
| Tagged / 2 | **Expected native success; wrong delivery (Alpine).** 1 task. | Expected native success; rebuilt application delivered. 3 tasks; original identity conditional. | Expected native success; rebuilt application delivered. 4 tasks; original identity conditional. |
| Digest / 2 | **Expected native failure** fetching unavailable `application@d₀`; original application not delivered. 1 attempted task. | Expected native success; rebuilt application delivered. 2 tasks; original identity conditional. | Expected native success; rebuilt application delivered. 4 tasks; original identity conditional. |

The tagged profile does not fetch d₀ in empty recovery in either world, so loss of d₀ does not prevent it from copying the unrelated image successfully. This is an important distinction from a failure-only oracle.

In tagged recovery, selecting **build alone** may also succeed natively while still delivering Alpine because release was not repaired. In World 2, selecting **retag alone** is expected to fail on missing d₀ before promote. If retag is unnecessarily rerun in the digest profile with a present original digest, it can succeed without changing the delivery requirement; with missing d₀ it fails unless build is also rerun.

For the digest-empty failure, a source-manifest fetch failure is expected before an application image is installed. The registry/task API does not supply a general transaction guarantee over every possible partial destination effect. Destination inspection is needed to establish its exact final state; wrong initial Alpine is not sufficient evidence that it remained untouched.

Full restart runs all four tasks and includes another network checkout/cleaning operation. It can be less available than retaining the known checkout. No task-count prediction is a latency, resource-cost or operator-effort measurement.

## 5. Identity and availability assumptions

### Immutable digest binding

Binding promotion to d₀ prevents a mutable release tag from choosing another manifest, assuming collision-resistant hashes, correct content verification, correct media-type/platform resolution, and correct repository access. A manifest digest binds the manifest bytes and their referenced config/layer descriptors. It does **not** establish that the content is the desired application, certify authorship, keep the manifest/blobs stored, provide access credentials, repair the repository, or prevent destination mutation after copy.

World 2 is explicitly manifest unavailability, not demonstrated garbage collection or disappearance of all copies. The API permits deletion and repository-scoped 404 even if blobs or another repository's manifest still exist. A client with only d₀ cannot recover the content merely from its hash.

### Frozen build inputs in the refined packet

The unchanged Dockerfile names `${BASE_PREFIX}golang:1.18` and `${BASE_PREFIX}alpine:3`. The final build arguments resolve those locators to `registry-source:5000/base/golang:1.18` and `registry-source:5000/base/alpine:3`. The packet states that this private mirror is frozen and remains available in both worlds. Its reference identities are:

- Go 1.18 base: `sha256:50c889275d26f816b5314fc99f55425fa76b18fcaf16af255f5d57f09e1f48da`.
- Alpine 3 base: `sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a`.

These locators still use tag syntax, but the admitted-world invariant fixes their content. Their immutability is supplied by mirror control, not enforced syntactically by a Dockerfile @digest reference. Recording and verifying the mirror's actual manifests/platform/config/layer identities remains necessary later. Deletion of `application@d₀` is scoped to the application repository; it must not change or make the frozen base repositories unavailable.

The application imports only standard-library fmt/time, builds with `-trimpath`, and sets `GOTRACEBACK=single`, with `CMD ["./app"]`. No downloaded application module dependencies are identified in the supplied Go files. With the same build context, base images, platform, compiler settings and fixed executor, re-creating B₀/C₀ is now a well-supported prospective expectation. Manifest creation/history timestamps may differ without changing that identity. Kaniko's `--reproducible` flag is not itself an independent proof of compiler determinism; actual binary and execution-config comparison is still the stipulated verification.

The original assessment's particular counterexample—upstream Go/Alpine tags drifting between source and recovery—is outside the refined packet's admitted worlds. It should not be carried forward as an unresolved input defect. Unexpected binary/configuration mismatch under the frozen-input assumptions would instead need investigation of actual build inputs, platform, context or tool behavior.

### Operational assumptions requiring later evidence

- The retained source Workspace has the pinned application's original relevant bytes and permissions after the first build; it is not replaced by an empty volume. Profile objects declare workspaces but do not include actual PipelineRun PVC bindings or byte snapshots.
- Config and all layer blobs of the selected manifest are accessible, not just its manifest GET. The correct platform is selected consistently; Skopeo defaults to `multiArch=system`, not all-platform copying.
- Retag/build destinations are writable; delivery really is restored; registry DNS/network/authentication and frozen-mirror base-image retrieval work. Runtime-image pin declarations do not demonstrate current availability.
- Recovery supplies saved d₀ when build is inherited and current d₁ when build reruns. Retag/promote retain the original scheduling dependencies even if completed tasks are removed.
- The expected destination state is restored for every arm without helping workflow tasks recover missing content; there are no racing writers or subsequent deletion/repointing.
- The fixed direct Skopeo invocation's exit code is faithfully reported by TaskRun status; no unobserved helper/wrapper substitutes an outcome or source.

## 6. Evidence needed to verify the composition

A later execution audit should preserve the following, independently of any planner's predicted labels:

1. **Source and task identity:** full PipelineRun/TaskRun captures, exact profile spec, actual resolved commands/parameters/runtime image digests, clone's actual commit, and hashes of the relevant checkout/Dockerfile/module/application files. Record native failure at promote and the earlier successful tasks.
2. **Source application reference:** d₀, raw original manifest/config and verified blobs/layers; B₀'s digest and executable mode; C₀'s execution-relevant settings. A code commit or manifest header alone cannot supply B₀.
3. **Initial world observations:** release resolves to the unrelated Alpine content; source `application@d₀` is pullable in World 1 and its exact manifest GET returns 404 in World 2; referenced-blob availability is separately established where needed. Destination starts at the specified wrong image. Record response media types and digest verification.
4. **Actual recovery execution:** generated/submitted manifests, literal or current result bindings, TaskRun names/counts, native terminal conditions and exit errors. These establish the minimal executed sequences predicted above, rather than selected tasks that did not actually run.
5. **Actual delivered content:** read delivery's manifest/config/blobs independently, verify requested-content hashes, reconstruct the effective layered filesystem (including whiteouts, overwrite order, symlinks and mode), and compare the executable `/app` bytes and execution configuration against B₀/C₀. Merely comparing delivery to whatever tag was copied would accept the wrong Alpine arm.
6. **Rebuild qualification:** preserve and verify the frozen mirror's resolved Go/Alpine base manifest/config/layer identities, platform and compiler/build context; establish their equality across source and rebuild and their availability in both worlds. Compare d₁ separately from B₀/C₀ so timestamp/media-format changes are not confused with semantic delivery failure. If checking exact configuration, explicitly distinguish execution fields from creation/history timestamps. Relevant fields include Entrypoint, Cmd, Env, WorkingDir, User and platform; executable access/mode also matters.
7. **Isolation and completeness:** preserve state-before/state-after evidence for each arm, with no intervening administrative restoration or writer that supplies the desired application. On failed copy, inspect destination explicitly rather than infer atomicity from nonzero exit.

A smoke execution that prints the expected message can complement the byte/config comparison but does not replace the stipulated original-binary identity test. Native success and delivery correctness must be reported separately in every cell.

## Locked conclusion

Conditional earlier-task minima are **tagged/World 1 `{retag}`; digest/World 1 `{}`; tagged/World 2 `{build, retag}`; digest/World 2 `{build}`**. Empty tagged recovery can succeed with wrong content in both worlds; empty digest recovery selects the original content when available and fails when that manifest is unavailable. The frozen private mirror removes the original upstream base-drift gap. With the fixed build inputs, preserved checkout, consistent platform/configuration and successful task/registry operations, all four minima are prospectively sufficient for the stated delivery objective; actual B₀/C₀ equality is to be verified from registry content. These conclusions are locked to the refined packet and delivery-only objective, before any actual results or revised manuscript are reviewed.


## 7. Final prospective offline-transport confirmation

This appended confirmation governs the final transport binding and input hashes; earlier sections retain the prior packet history. It was prepared before receiving any source/recovery PipelineRun, plan, policy or outcome for this case.

All **14 files** listed in the final SHA256.json matched downloaded bytes. Final hashes are:

- Tagged profile: `1c780672fac4c634d7d6dd5a1f0cf78a18f2f06d76e27f6d3051d61176f1ca56`.
- Digest profile: `282e42fe3b77650b1d1acce0e9b9e2fd2d351fd4e055ea96f9cc3c32252f93c4`.
- SOURCE_BRIEF.md: `de0cd8f9f2c2b05a2890f1013fe7451b75e2f771d02c4a7abe01c0537f670b7b`.
- offline-transports.json: `53ede5a271053b3314fbbe36d5cbb2b871d51d3a7f5d8d3756f440ed122ffa9d`.

Compared with the frozen-base packet, the entire Pipeline spec changes only clone's URL to `git://git-source:9418/skaffold.git`. The commit stays `035e3ddc3589c1c64cb69c563b178f34b1cccc5e`; every embedded Task spec, build argument, application file, ordering edge, promotion locator and registry-world definition remains unchanged. Updated metadata reflects the new profiles. The source brief additionally states read-only Git service, preloaded/local image transport, retained frozen base availability, and pod fsGroup 65532. The profiles themselves are Pipelines, not evidence of actual pod security contexts.

The new transport record specifies Linux/amd64, Go 1.18.10 base configuration, seven Go base rootfs layer identities, one Alpine base layer identity and one foreign-Alpine layer identity. Origin repository digests agree with runtime-images.json. It records Git/image archive checksums and image archive byte counts. Those archive files and live mirror responses are not included in this packet, so their actual bytes and source-to-mirror equality have **not** been independently recomputed in this pass; the JSON records supplied preparation attestations. Later raw commit/checkout and manifest/config/layer evidence should verify those bindings. Rootfs layer identities must not be confused with archive-file hashes or compressed registry blob digests.

The local transports change retrieval location and availability assumptions, not the repair requirements. Conditional minima remain **tagged/World 1 `{retag}`; digest/World 1 `{}`; tagged/World 2 `{build, retag}`; digest/World 2 `{build}`**. Predicted empty tagged success with wrong content, digest World 1 correct delivery, and digest World 2 missing-manifest failure remain unchanged. Frozen base inputs, preserved source bytes, consistent platform, current-result propagation, complete blob availability and actual B₀/C₀ equality remain explicit qualifications/checks.

Public-network clone/image-pull latency is excluded. Local daemon/mirror availability can support full restart, but this preparation packet establishes no recovery latency, operator effectiveness, security/authenticity guarantee or production deployment result. No native experiment was run and no contact was made for this confirmation.


## 8. Post-calibration entrypoint qualification

The coordinator reports that the first attempted source is a retained **calibration failure**: clone succeeded, but build had PodCreationFailed because the controller could not reach gcr.io to discover the pinned Kaniko image's entrypoint. Build and recovery did not execute. The evaluated profiles subsequently set `build.taskSpec.steps[0].command=["/kaniko/executor"]`, reported to match the same pinned image's inspected Entrypoint. Images, arguments, scripts, application, registry worlds and promotion locators are reported unchanged. Evaluated sources use distinct `source-evaluated-{profile}` names.

This confirmation uses that supplied description, not new raw manifests or Docker inspection: those were not supplied in this message and have **not yet been independently verified**. Native execution has started. This entrypoint binding and this confirmation are therefore post-calibration execution refinements, **not preregistered changes or assessments made before native execution**. The earlier source-only minima remain documented as locked before execution evidence.

**Effect on the four minima:** none, provided the raw image configuration confirms that the explicit command invokes exactly the original image entrypoint with the same effective arguments and execution environment. An explicit `/kaniko/executor` command removes the need for controller entrypoint discovery; it adds no credential, image producer, tag repair, digest locator or application dependency. For the previously stipulated source state—successful clone/build/retag followed by failed promote—the earlier minima remain tagged/World 1 `{retag}`, digest/World 1 `{}`, tagged/World 2 `{build, retag}`, and digest/World 2 `{build}`. Fixed-input reproducibility, correct result binding, complete content availability and actual delivery identity qualifications remain.

The reported calibration run does **not** have the stipulated promote-failure source state: build/retag never succeeded. Its recovery requirements cannot be read from the four promote-failure cells. Recovery from that calibration boundary would still need build and the not-yet-executed suffix, subject to obtaining a runnable Pod. This describes remaining application work, not admission of a recovery plan from the changed calibration source specification. It should be retained separately from evaluated source failures and recovery arms, without counting PodCreationFailed as executed build work or evidence of the derived promotion repairs.

The final audit should retain the calibration PipelineRun/TaskRuns and events, before/after evaluated Pipeline specs, the pinned Kaniko image's raw configuration, effective command/args and runtime image identity, and the two distinct evaluated source captures. Explicit `command` is a real task execution binding adaptation, not merely image pinning or an empty native serialization default. Descriptions of public task-body reuse should disclose it. Raw verification can establish whether the reported sole change is semantic equivalence of entrypoint invocation; no evaluated outcome is inferred here.

No experiment was run, no shared-repository source or future output was read, and no contact was made for this bounded qualification.


## 9. Post-calibration Kaniko/runtime-file qualification

Latest source packet: https://artifacts.internalmeta.com/a/2479895562538338. All **14 listed file hashes** match downloaded bytes. The controlling current hashes are:

- Tagged profile: `fa3bd1cd3c8b19c1183e72db8bbdd44a1f4586c358764140bd26c267ce376037`.
- Digest profile: `7b74739b7b1cbdd7edc112670bef271903ba2c2794cd72864b249cb3d94628ef`.
- SOURCE_BRIEF.md: `ebca8caddc85df1edc10296f72f839ee6a581a9094143f57a213cc504823f09b`.

Relative to the last downloaded transport packet, exactly the two profiles and SOURCE_BRIEF.md changed among the files named by the checksum manifest. The changed Pipeline-spec fields are (1) build step command `[/kaniko/executor]`, (2) appended EXTRA_ARGS `--ignore-path=/product_uuid`, and (3) an explicit BUILDER_IMAGE parameter `gcr.io/kaniko-project/executor@sha256:9e69fd4330ec887829c780f5126dd80edc663df6def362cd22e79bcdf00ac53f`. All step scripts and args templates match the earlier packet. The default old BUILDER_IMAGE remains in the unchanged Catalog Task parameter declaration but is overridden by this explicit PipelineTask parameter. Application files, frozen base declarations, clone commit/transport, scheduling and promotion locators remain unchanged.

The coordinator/source brief report a second calibration source that compiled the application under old Kaniko 1.5.1 but failed cleaning an injected `/product_uuid` file before publishing an image. No recovery arm executed. This report was not independently reconstructed from raw calibration objects/logs in this pass. Together with the earlier controller-entrypoint discovery failure, it is calibration evidence to retain separately, not an evaluated promote-failure arm or a successful source image.

**The four earlier-task minima do not change:** tagged/World 1 `{retag}`; digest/World 1 `{}`; tagged/World 2 `{build, retag}`; digest/World 2 `{build}`. Their necessity follows the fixed image locators and unique producer/retag roles, not the builder version. The new executor still has the same declared role: read retained application context, publish to the application repository and emit IMAGE_DIGEST. The ignore-path is reported to exclude an injected runtime file from cleanup/snapshot. No supplied Dockerfile instruction or application code uses `/product_uuid`, so excluding that path supplies no new application recovery dependency. This reasoning is conditional on that file being environment-only and the final executor implementing the intended build/push/digest contract; source inspection does not validate arbitrary executor-version equivalence.

For identity verification, **B₀/C₀/d₀ must come from the final evaluated source's successful published image**, built with the same pinned Kaniko 1.23.2 binding used for its recoveries. The unpublished calibration build is not that reference. A version change can affect image construction, and no claim is made that Kaniko 1.5.1 and 1.23.2 produce equivalent image bytes. Rebuilding under the final fixed source/base/context/platform/executor inputs remains expected to recover B₀/C₀, to be checked independently from delivered content. Manifest digest equality remains separate from that application predicate.

The current runtime-images.json and offline-transports.json are unchanged and still document the earlier Kaniko 1.5.1 inventory rather than an inspected 1.23.2 configuration. The explicit final profile digest identifies the intended new executor; this packet does **not** include its raw Docker inspection or actual resolved runtime identity. Final evidence should retain those records, confirm `/kaniko/executor` against the **new** pinned image configuration, and verify actual argument support/exit status, relevant source/output bytes and final delivered identity. The runtime-file adaptation should be disclosed as a calibrated execution option, not native serialization or an unchanged executor implementation.

The original application-repair predictions were locked before attempted runs. The explicit command, upgraded builder and ignore-path bindings are later calibration refinements, **not preregistration**, and this confirmation follows those refinements. The source brief's opening prospective wording should be read with its explicit later calibration section; it does not imply that these final bindings preceded every attempted run.

No native experiment, source modification, shared-repository policy/plan/output lookup, contact, or venue adjudication was performed for this confirmation. Full native and delivery evidence remains pending.
