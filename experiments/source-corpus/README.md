# Public Tekton source corpus: conditional recovery prerequisites

## Evidence and method

Frozen sources: [frozen-source-packet.zip](frozen-source-packet.zip)

Prospective coding protocol and later provenance clarification: [coding-protocol.md](coding-protocol.md)

Original coder A: [coder-a-original-labels.zip](coder-a-original-labels.zip)

Original coder B: [coder-b-original-labels.zip](coder-b-original-labels.zip)

Source lookup resolving the array-extractor and image-binding questions: [source-lookup-evidence.json](source-lookup-evidence.json)

The frozen convenience frame enumerated 614 standalone/embedded Pipeline objects, retained 89 canonical eligibility candidates, and selected 39 distinct bodies after source resolution and duplicate/family exclusions. The predeclared cap was 40, with lexicographic selection before effects were labeled; it was not reached. The packet contains 391 Task occurrences (359 ordinary, 32 finally) and 425 resolved StepAction occurrences. Runtime image/module behavior is not resolved by retaining their YAML.

Two automated coders created separate original records. Both inherited project/session context, including manuscript examples, and share model provenance; neither consulted the other's labels. They are neither human participants nor external validators. Each uses finite literal-operation/interface extraction and source inspection, with unlisted opaque behavior retained as Unknown. The independently created scripts and original outputs are retained. Coverage means the defined source fields were traversed, not complete semantic analysis.

## Counts and adjudication

Post-coding normalization uses an ordered Task pair within a composition. Multiple records on the same pair are combined only for the binary question of whether either record supports a conditional prerequisite. Missing records are distinct from Unknown. This normalization and the agreement calculation are descriptive post-coding analyses, not predeclared semantic endpoints.

There are 625 pairs coded by both agents, one pair coded only by A and seven only by B: 633 pairs in the union. On the common pairs, both support 208, both leave 381 unresolved, and B alone supports 36; A alone supports none. Binary support-status agreement is therefore 589/625. This is not complete effect/event/semantic or timing agreement, and its many related source bodies cannot be treated as independent observations.

All 36 common-pair differences concern an array-artifact extraction interface omitted by A's narrower single-artifact rule. Frozen producer results, Pipeline array bindings and the consumer's `sourceDataArtifacts` extraction interface support a conditional availability prerequisite. The raw labels were preserved, rather than silently overwritten. Adjudication also corrects the registry pair's structural visibility: Kaniko and GKE deployment share a Workspace even though their bound image state lives in an external registry.

The adjudicated union contains 249 Interface-supported conditional relationships and 384 Unknown relationships. Supported relationships occur in 34/39 bodies: 23/24 related release-service bodies, 8/9 Catalog compositions and all 3/3 related OpenShift profiles. None is supported in the one integration or two KubeVirt bodies under the retained coding scope. These counts are source-characterization counts, not recovery success or incident prevalence.

The 249 supported pairs comprise 233 artifact flows, eight checkout/file prerequisites, three cluster-credential prerequisites, three conditional finally-teardown interactions, one authentication-configuration interface, and one registry-image binding. Structural visibility is an explicit result link for 236, common Workspace binding for ten, and phase ordering/shared upstream cluster identity for three lifecycle interactions. No actual planner zero-warning count was measured.

## Interpretation

A retained artifact locator can outlive availability of its bytes; the source interfaces explicitly allow artifact expiry and skipped creation. A shared checkout or authentication directory can disappear or change. Finally teardown can end the resource lifecycle before later recovery, despite saved successful Task records. Source declarations establish those conditional interfaces; the dataset does not establish that a producer rerun is a complete or minimum valid repair. Recreating a resource can also change identity and invalidate retained consumers.

Most evidence concerns already visible wiring and related release workflows. The OpenShift examples run connectivity checks such as `oc whoami`; a release sample has a fake advisory producer. Opaque helpers, runtime parameters, image behavior, remote manifests, API semantics, optional branches and complete application validity remain uncertain. Negative, trivial and unresolved cases are retained. The source coding does not discover a representative population of warning-invisible real failures.

The audit verifies 291 source-file hashes, both coders' 391-occurrence coverage, 10,862 source quotations (allowing one terminal newline), and all reported arithmetic with zero mismatches. Quotation reproduction is evidence integrity, not validation of an interpretation or execution.

## Remaining value question

This corpus grounds a case bank and clarifies recovery conditions in existing public compositions. It does not measure operator effort/errors, real incidents, successful recovery, full repair sets, or an advantage over equally informed manual/persisted selection. The value of persistent recovery knowledge remains an open empirical question. Real failed-run histories with matched recovery choices, or a controlled human study, remain needed for that value claim.

Prepared study protocol: [value-study-protocol.md](value-study-protocol.md)

Bounded public-history feasibility result: [public-history-feasibility.md](public-history-feasibility.md)

## Upstream attribution

The frozen packet preserves source URLs, exact commits and original file headers. Accompanying license texts fetched from those exact upstream revisions are retained under `upstream-licenses/`; its index maps each text to its source. These supplemental files do not change the frozen packet or original coding hashes.

## Post-review structural comparison

`structural-baseline.json`, generated by `hack/corpus_structural_baseline.py`, compares graph-ancestor selection (runAfter, string results and native phase order) and its shared-Workspace expansion. The comparison is offline and was specified after coding; it is not a semantic repair oracle. All 246 non-lifecycle supported pairs already lie on graph-ancestor paths. Workspace expansion adds no supported pair. The other three pairs are finally-teardown invalidations of earlier consumers, not established producer-rerun obligations. Complete valid repairs remain Unknown. The corpus consequently does not establish that graph-invisible prerequisites are common.

Of 249 supported pairs, 236 have direct result references and ten common Workspace bindings; 234 come from the release-service repository. The canonical eligibility ledger excludes 5/5 build-definitions candidates and 36/45 Catalog candidates. Cohen’s kappa over 625 shared binary-support labels is 0.8756878605 (observed agreement 0.9424, chance agreement 0.53665024). The adjudicated positive-pair set equals coder B’s original positive-pair set; source resolution of the 36 disagreements is not independent validation.

For the external-state fixture, graph-ancestor seeds plus unreached continuation select all six tasks. That action set matches the archived full-restart arms: six executions, 36 TaskRuns. Declared and equally informed manual selection each use 30 TaskRuns. This post-hoc reconstruction adds no native execution and establishes no policy-format or operator advantage.

Reproduce this supplementary analysis from the repository root:

```bash
/usr/bin/python3 hack/corpus_structural_baseline.py --corpus experiments/source-corpus --output /tmp/structural-baseline.json --external-evidence results/acceptance/live
```

The baseline treats every source Task as applicable, does not evaluate runtime `when` conditions, and does not infer external identity, availability or repair validity from a selected graph ancestor.
