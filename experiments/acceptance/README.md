# Recovery-contract revision evidence

This revision adds explicit observable source admission, full-definition string-result advisories, a contract argument, exhaustive structural checks, and an external-state experiment. It does not implement a controller, guarantee semantic policy correctness, or measure practitioner authoring or production savings.

## Reproduce offline

From the repository root, using Go 1.23 or later:

```sh
go test ./...
go build -o /tmp/retryctl ./cmd/retryctl
/tmp/retryctl evaluate --cases evaluation/cases.yaml
python3 experiments/acceptance/audit.py
```

`go test ./...` includes `TestExhaustiveRecoveryProjection`: 1,024 forward DAGs on five vertices × 31 nonempty retained sets = 31,744 generated recovery transformations. The independent graph oracle is Boolean Floyd-Warshall. The test mixes runAfter, parameter-result and when-result edges. It deliberately checks arbitrary abstract partitions; those are not all realizable failure traces or live Kubernetes tests. It takes about 80 seconds on the recorded host. The six field/namespace advisory checks are source-level tests; legal native syntax and the string-result restrictions bound the application claim.

The original full fixture suite remains 45/45, including two target-parameter extension cases outside the original manuscript's 43-case table. Authored source fixtures now declare their terminal PipelineRun condition explicitly. Historical native captures are unchanged. `results/acceptance/scenarios.json` records the revised fixture run; `results/acceptance/tests.jsonl` records the full Go test run.

## Reproduce the native experiment

Prerequisites: a disposable Tekton v1.6.0 cluster, kubectl, Python 3, port 29173 free, and image access for python:3.12-alpine and alpine:3.20.

```sh
python3 experiments/acceptance/live.py --retryctl /tmp/retryctl --context YOUR_TEST_CONTEXT --namespace A_NEW_UNUSED_NAMESPACE
python3 experiments/acceptance/audit.py
```

The harness creates a new namespace and refuses to reuse an existing one. It does not modify existing Pipeline or PipelineRun objects in other namespaces. It retains its namespace for audit; delete that namespace manually when no longer needed. The result directory is overwritten by a new experiment, so preserve or copy existing evidence first. The checked-in run used `kind-retry-paper`, namespace `retry-acceptance-20261030`, and Tekton v1.6.0; exact image digest is recorded in `results/acceptance/live/environment.json`.

Each source has six serial tasks: warmup, acquire, publish, sign, scan, deploy. Three separate source runs fail once at publish, sign, or scan. All tasks call an external service, and none exports a result or mounts a Workspace. Each complete source snapshot is then used unchanged for stable and rotated service-state worlds. Each world has four arms, shuffled with seed 20261030: empty selection, declared closure, policy-informed manual selection using the same generator, and full restart. There is one execution per failure/world/arm, giving 24 native arms, not a population sample. A fixed four-entry policy is reused without editing; the fourth entry covers deploy but deploy failure is not evaluated. Policy-informed manual selections deliberately use the same conservative application requirements, without querying the service epoch, and give the same dispositions. A state-aware minimum comparator is additionally reconstructed from the existing stable empty arms and rotated manual arms; it is not a fifth live arm or a human participant. Its minimum task sets were independently source-adjudicated. No operator-time or mistake-rate advantage follows from the comparison.

The service in `oracle.py` stores and validates actual credential/artifact/signature/scan/deploy epochs. It has no access to Tekton objects, planner output, or labels. Snapshots/restoration pair each arm to the same initial external state. `audit.py` independently recomputes native success from conditions, TaskRun counts from actual lists, and final consistency from external state fields; it does not accept the observations ledger as ground truth. It is project-authored, so separation from the planner is not an independent practitioner oracle. Separate source-only automated adjudications are archived before full-paper re-review, with that distinction explicit.

Finally, the harness captures a failed ordinary task while a finally task remains active. The revised planner refuses the active source, create emits no manifest, and planning is permitted after the source completes unsuccessfully. This checks observable status admission, not atomic capture or absence of surviving remote work.

## File map

- `experiments/acceptance/contract.md`: definitions, proof, and limits.
- `internal/planner/planner.go`: admission and advisory fixes.
- `internal/planner/revision_test.go`: independent projection oracle and regressions.
- `experiments/acceptance/oracle.py`: external application semantics.
- `experiments/acceptance/live.py`: paired native-run harness.
- `experiments/acceptance/audit.py`: raw-evidence recomputation.
- `results/acceptance/live/<failure>/source-{run,taskruns}.json`: identical source inputs used for both worlds.
- `results/acceptance/live/<failure>/<world>/initial-state.json`: shared initial external snapshot for that world's four arms.
- `results/acceptance/live/<failure>/<world>/<arm>/`: plan/generated manifest where applicable, native run/TaskRuns, external final state.
- `results/acceptance/live/active/`: native active-to-terminal source trace and manifest refusal.
- `results/acceptance/audit.json`: recomputed outcomes and input hashes.
- `results/closure-current-20261029/`: unchanged original current-version Workspace evidence.
- `results/`, `evaluation/konflux/`: original timing and versioned replay artifacts, retained unchanged.

The contribution is the executable Tekton contract and tested execution composition, with an explicit boundary between observable planning and application semantics. Reachability algebra, partial execution and informed task selection are credited as existing ideas.
