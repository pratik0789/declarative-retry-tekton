# Public Tekton failure-history feasibility check

## Result

**No eligible failure-history dataset was established.** This is a bounded source/access check, not a corpus experiment, model execution or venue recommendation. It cannot support a claimed practical-benefit retrospective.

The final check completely listed the public `tekton-releases` GCS prefixes `pipeline/`, `results/` and `chains/`: 439, 69 and 94 objects respectively, with no next-page tokens. The 602 listed objects are release/install manifests or directory-marker objects. No release log, raw PipelineRun or raw TaskRun object was present. Therefore zero existing log objects could be opened under the requested cap of twelve; no failed-run or manual-recovery pair was extracted.

Raw listing responses, pinned publication source and a machine-readable coverage index accompany this report in the evidence ZIP. Counts describe these exact listings; they do not imply a survey of all public Tekton archives.

## Publication source and selection problem

Sources resolved to Tekton plumbing commit **2d1ad9b5ee76d8a543093ef6e112df26063c9008**.

- [Release logging description](https://github.com/tektoncd/plumbing/blob/2d1ad9b5ee76d8a543093ef6e112df26063c9008/tekton/resources/release/README.md), lines 30–38: logs are saved alongside release YAML, and collection is triggered **when the publish Task executes successfully**.
- [Actual collection Task and Pipeline](https://github.com/tektoncd/plumbing/blob/2d1ad9b5ee76d8a543093ef6e112df26063c9008/tekton/resources/release/release-logs/save-release-logs.yaml): the collector runs `tkn pr logs -n "$NAMESPACE" "$PIPELINE_RUN" --all > "$LOG_FILE"`; the uploader stores `logs.txt`. It does not export PipelineRun JSON/YAML, resolved Task definitions, TaskRun objects or native condition/timestamp records.
- [Trigger bindings](https://github.com/tektoncd/plumbing/blob/2d1ad9b5ee76d8a543093ef6e112df26063c9008/tekton/resources/release/release-logs/release-logs-triggers.yaml): source PipelineRun name/namespace and bucket path parameterize the collection workflow. These are source definitions, not captured instances or evidence of a repair choice.

Even if matching logs became available, the publication trigger would miss failures before publication. Concatenated pod text may identify Tasks/steps and some reasons, but would remain ineligible for a **full** retrospective unless augmented with observed immutable/resolved source, TaskRun status/timing and repair linkage. The configured publication mechanism cannot establish an unbiased failure-history denominator.

## Access results and exclusions

- Public GCS listings were readable anonymously. Reproducible listing URLs and complete object-name/size/creation-time responses are in SOURCE_INDEX.json and the ZIP. No object bodies were needed once the scoped names showed only install/release manifests.
- `https://prow.infra.tekton.dev` is reachable; the current [Prow source](https://github.com/tektoncd/plumbing/blob/2d1ad9b5ee76d8a543093ef6e112df26063c9008/prow/control-plane/config.yaml) describes its merge/tide configuration. Prow jobs were **not** treated as Tekton PipelineRun histories. The obsolete `prow.tekton.dev` probe failed TLS.
- `https://results.infra.tekton.dev` failed certificate validation with an expired certificate. Verification was not bypassed. More decisively, [the pinned Results deployment patch](https://github.com/tektoncd/plumbing/blob/2d1ad9b5ee76d8a543093ef6e112df26063c9008/tekton/cd/results/overlays/oci-ci-cd/patch-disable-results.yaml) sets API/watcher/retention/postgres replicas to zero, explaining broken storage-image pulls and finalizer pileups. This is an unavailable public dataset lead, not native historical evidence.
- Directory/index probes of the existing infra release paths returned 404. A prior concrete `pipeline/previous/v0.65.0/logs.txt` probe also returned 404. No guessed log path was promoted to a record.
- The final task explicitly narrowed scope to these release sources; no further broad Konflux/Prow or unrelated deployment search was performed. No access controls were bypassed; no workloads, native tests or source code were executed; no repo/manuscript files were edited; no humans were contacted.

## Minimum usable data for the proposed value study

The following must come from a real deployment or authoritative public archive, not be reconstructed from fixture expectations:

1. Source failed PipelineRun including immutable/resolved Pipeline/Task definitions and concrete params/workspace bindings.
2. Native TaskRuns with completion conditions, start/end times and saved results, joined to the source UID.
3. Observable retained-state facts or source-derived contracts needed to assess reuse; otherwise report the safety uncertainty explicitly.
4. Actual manual recovery choice or a genuinely source-linked subsequent PipelineRun, with its outcomes and output-validity evidence.
5. A documented collection denominator and exclusions covering failures before publication, not success-conditioned release logs alone.

Without these records, source-level warning/closure coding can be a separately scoped descriptive study, but it cannot be relabelled observed operator-error reduction, valid task-time savings or real incident recovery.

## Frozen GitHub issue follow-up

A subsequent bounded screen fixed twelve public GitHub queries before reading bodies and inspected twenty-two unique issues across tektoncd/pipeline, tektoncd/catalog and konflux-ci/build-definitions. It established **zero eligible completed failed PipelineRun case packets**. Issue #9467 records a genuine parent-failure/child-recovery incident, but lacks original native PR/TR snapshots, resolved/versioned definitions, complete sibling/external state and actual recovery linkage. It remains an incident lead, not a retrospective case.

See the added `issue-feasibility.md`, `FROZEN_SCREEN_DECISIONS.json`, and `frozen-public-issue-feasibility-packet.zip` for the exact fixed protocol, raw public responses/bodies, source hashes and every exclusion. No policy labels or benefits were inferred. Packet SHA-256: `bf30bbd625e1a055e4ac626be93877a9fb7a2bc9c0717053a091b7f2904e2266`.
