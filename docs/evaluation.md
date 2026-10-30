# Evaluation status

The current evaluation contains 45 deterministic cases:
22 accepted recovery plans, 23 refusals, and eight warning-bearing accepted
plans. All specified decisions, dispositions, refusal reasons, and required
warning codes pass. The report is retained in
`results/current-version-audit-20261029/planner-evaluation.json`.

The original suite contained 23 cases and the earlier companion revision 31.
The current revision changes two formerly accepted later-task declarations to
refusals and adds seven cases for earlier and parallel tasks, a transitive
violation through a parallel branch, explicit operator selection, YAML order,
and result references in parameters and conditions. Unit tests also cover
unreachable policy entries, embedded task references, and unresolved execution
dependencies. These authored examples establish finite specification coverage,
not population correctness.

The finally-support revision adds seven cases for failed finally tasks, ordinary
and parallel closure members, explicit selection, `blocksResume`, the phase
ordering restriction, and shared ephemeral Workspaces. The existing finally
case now expects a successful cleanup task to be reused instead of always
running again.

Run them with `make test evaluate`.

The definition guard now requires literal equality of the saved and current
Pipeline specifications. Regression tests cover changed variable references,
literal-to-variable edits, ambiguous resolved values, and deleted tasks. The
45-case planner suite remains the same authored suite; these regressions are
additional unit tests. PipelineRun generation is not Pipeline generation.
Both IV-G mechanisms and all three finally-task recovery cases were rerun with
the current strict guard on October 29, 2026. Their task scripts now use TaskRun
context for source-only failures and fresh input, preserving original
expressions in the saved PipelineSpec. Native saved/current specifications
match without rewriting captures. The IV-G verifier also refuses changed
scripts and context expressions before producing a recovery manifest. All
eight Konflux replay variants regenerate with the current binary and match the
retained results.

Earlier finally and closure captures whose PipelineSpec resolved PipelineRun
context expressions remain conservatively refused by the current guard. These
historical outcomes are retained, but current IV-G and finally-case claims use
the new captures. The basic CI integration also uses TaskRun context.

See [matched closure comparisons](closure-ablation.md) for the new live
three-arm experiment and its independent recomputation of native objects and
Workspace bytes. Its output oracle does not use planner labels; policies and
task assertions nevertheless share authorship.

## Finally-task recovery

Finally tasks use the same policy and task dispositions as ordinary tasks. A
failed finally task seeds its closure; successful finally tasks outside that
closure are reused. Generation keeps tasks in their original ordinary or final
phase. Ordinary tasks precede all finally tasks, and finally tasks on parallel
branches can be named in the same recovery closure.

The live harness runs three source failures, each confined to `notify` in the
final phase, and verifies the generated recoveries:

- Only failed `notify` runs again; all ordinary work and successful `audit` are reused.
- `notify` pulls in `build`, which pulls in `prepare`; the new build finishes before notification.
- `notify` pulls in successful parallel `audit`; both finally tasks run again.

When all ordinary tasks are reused, Tekton v1.6 requires a nonempty ordinary
`tasks` section. The generator adds an always-skipped placeholder, records its
name in an annotation, and creates no TaskRun for it. The first and third runs
therefore execute only finally tasks. Notification
receipts are checked against the actual inherited or newly produced image
result, and native TaskRuns are checked against the plan. Evidence is in
`results/finally-current-20261029/`. These are three synthetic feasibility
cases, rather than a timing or production effectiveness study.

```bash
go build -o /tmp/retryctl ./cmd/retryctl
python3 hack/finally_recovery.py --context kind-retry-paper \
  --namespace retry-finally-new-unique-name --retryctl /tmp/retryctl \
  --output artifacts/finally-recovery
```

## Current-version finally-result regression

The current implementation was tested with native captured inputs whose
Pipeline specification matches exactly. `build` and finally task `audit`
succeed; finally task `notify` fails. The Pipeline exports
`$(finally.audit.results.note)`, whose saved string value is `checked`.
The pre-fix generator removes audit and leaves that reference unresolved; the
Tekton admission webhook rejects the generated run. A literal replacement is
also rejected because Tekton requires Pipeline outputs to reference task results.

The fixed generator reuses build and audit and emits a small ordinary task
that publishes audit's saved value. The recovery runs only that publisher and
notify: two TaskRuns, without repeating the original audit or build work.
Native TaskRun and PipelineRun results confirm the exported value `checked`
and notification receipt `source-image`. The recovery succeeds. This is one
synthetic composition regression, not a timing or production effectiveness
result. The October 29 rerun with the same current binary used for IV-G is in
`results/finally-export-current-20261029/`. An etcd read timeout interrupted the
harness after submission; guarded `--resume` verified existing specifications
and collected the completed recovery without creating another run. The
October 26 evidence remains in `results/finally-export-recovery-20261026/`; the rejected
literal attempt is retained in `results/finally-export-rejected-literal-20261026/`.

```bash
go build -o /tmp/retryctl ./cmd/retryctl
uv run --with pyyaml python hack/finally_export_recovery.py \
  --context kind-retry-paper --namespace retry-finally-export-new-unique-name \
  --retryctl /tmp/retryctl --output artifacts/finally-export
```

After an interrupted read, repeat the same command with `--resume`. It checks
the original namespace, context, generator source, binary, and existing
resource specifications. Only reads are retried automatically.

## Stateful recovery on kind

`make stateful` runs `integration/stateful.yaml`. `clone` writes a random file to
the shared Workspace and emits its commit identifier and SHA-256 digest; `build`
verifies both before emitting an image name; `scan` fails in the source run. The
harness recovers two source runs and then checks the definition guard:

- Persistent claim: recovery inherits `clone`, substitutes its commit result as
  a literal parameter, and `build` verifies the file and digest (`STATE-VERIFIED`).
- `volumeClaimTemplate`: `create` stops for confirmation and reports the
  `ephemeral-workspace-across-closure` warning; after confirmation, `build`
  fails because the file is missing (`STATE-MISSING`).
- After one line of the Pipeline changes, planning the first source run again is
  refused because the definition changed.

On 2026-10-04 the experiment was run six times on kind v0.27.0 / Kubernetes
v1.31.4 / Tekton v1.6.0 with v0.3.0, and once more with v0.3.1; all seven gave
the same four outcomes. Results are in
`results/v0.3.0-kind-v1.31.4-tekton-v1.6.0/stateful/` and
`results/v0.3.1-kind-v1.31.4-tekton-v1.6.0/stateful/`.

## Ordering around inherited tasks

`make ordering` runs two Pipelines on kind and recovers each with a selected
task, so that deploy's only edge points at an inherited task:

- Linear (clone, build, scan, migrate, deploy): `--tasks scan` reruns build and
  scan, migrate is inherited, and the failed deploy reruns.
- Fan-out/fan-in (clone, build-a and build-b, package, publish, deploy):
  `--tasks build-a` reruns build-a; build-b, package, and publish are inherited.

In both, the harness checks Tekton's timestamps that deploy started no earlier
than the rerun task above it completed. The fan-in source run also exercises
`--newer-runs-from-cluster`: planning is allowed before the recovery run
exists and refused after, because retryctl lists the namespace's PipelineRuns
itself. Results are in `results/v0.3.2-kind-v1.31.4-tekton-v1.6.0/ordering/`.

## Konflux replay

`make replay` injects a failure at each task of the Konflux template-build
Pipeline (vendored, unmodified, at upstream commit bfe4d25) and records the plan
under the author policy in `evaluation/konflux/policy.yaml` and under an empty
policy, with the source Workspace bound by a persistent claim or a
`volumeClaimTemplate`. Two completion models are reported: only tasks upstream
of the failure completed (`replay-*.json`), or every task that does not depend
on the failed task completed, including parallel siblings
(`replay-independent-*.json`). This exercises the planner only, not a live
Konflux installation.

The pinned live harness uses kind v0.27.0, Kubernetes v1.31.4, and Tekton
Pipelines v1.6.0. Run `make cluster install-tekton wait-tekton integration`.
The integration injects a source-run failure, exports real PipelineRun and
TaskRun resources, generates a three-task recovery run that inherits one task,
and requires that recovery run to succeed. It also executes a full restart and
records both durations, TaskRun counts, Kubernetes resources, and events under
`artifacts/integration/`. GitHub Actions retains the same evidence as a workflow
artifact.

An initial attempt with Tekton v0.62.3 could not pull its legacy `gcr.io`
images. The harness now pins v1.6.0, whose official images are published through
`ghcr.io`.

## Verified local execution

On 2026-10-03, the pinned harness completed successfully on kind. The full
restart executed four TaskRuns in 17.896 seconds. The generated recovery run
executed three TaskRuns in 13.970 seconds, inherited the successful `clone`
task, and avoided one of four tasks (25%). Both PipelineRuns reached
`Succeeded=True`. The raw resources, events, generated manifest, and summary
are retained in `results/kind-v1.31.4-tekton-v1.6.0/`.

On 2026-10-04 the same harness was rerun with the v0.3.0 planner (stricter
definition check, result substitution, field preservation). The full restart
executed four TaskRuns in 16.899 seconds and the recovery run three TaskRuns in
13.985 seconds; both succeeded. Results are in
`results/v0.3.0-kind-v1.31.4-tekton-v1.6.0/integration/`. One benchmark pair per
configuration was also rerun with v0.3.0 to confirm that every recovery run
succeeds and skips the same TaskRuns (`benchmark-spot-check/` in the same
directory); those single trials are not used as timing results.

This is one controlled feasibility run, not a statistically meaningful
performance result. Timing claims require repeated trials.

## Synthetic performance benchmark

`make benchmark` executes measured full-restart and declarative-recovery trials
on the live Tekton cluster. By default it evaluates 5- and 10-task linear
pipelines, early/middle/late failures, and three repetitions per configuration.
Raw per-trial measurements and grouped means are written under
`artifacts/benchmark/`. Configure larger experiments with `SIZES`, `POSITIONS`,
`REPETITIONS`, and `TASK_SLEEP`; for example:

```bash
SIZES="5 10 20 40" REPETITIONS=20 TASK_SLEEP=1 make benchmark
```

The 60-trial run used in the paper was repeated on GitHub Actions with
artifact-v0.3.0; raw trials and confidence intervals are in
`results/v0.3.0-ci-60-trials-kind-v1.31.4-tekton-v1.6.0/`.

The benchmark reports TaskRuns avoided and observed wall time. Synthetic timing
must be described as controlled experimental evidence, not production savings.
