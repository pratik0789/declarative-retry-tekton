# Declarative Retry for Tekton

Research prototype for planning safe, declarative recovery of failed Tekton
`PipelineRun`s. The prototype is external to Tekton and does not represent an
existing Tekton API.

Published result: [end-to-end feasibility report](docs/feasibility-report.md)
(`feasibility-v0.1.0`).

Measured evaluation: [synthetic Tekton benchmark](docs/synthetic-benchmark.md).
The published dataset contains 60 paired live Tekton trials.

Mechanism-specific evaluation: [matched stateful closure comparisons](docs/closure-ablation.md).
These compare an empty policy, a declared closure, and full restart for partial
Workspace writes and stale builds after an operator refreshes earlier input.

## Quick start

```bash
go run ./cmd/retryctl plan \
  --pipeline testdata/simple/pipeline.yaml \
  --run testdata/simple/run.yaml \
  --policy testdata/simple/policy.yaml \
  --taskruns testdata/simple/taskruns.yaml
```

The planner classifies each pipeline task as `rerun`, `inherit`, `continue`, or
`refuse`, and emits a deterministic JSON recovery plan with reasons and source
provenance. The policy author owns the retry closure. When results or writable
Workspaces cross its boundary, the planner reports structured warnings instead
of changing or refusing the declared closure. Structural errors and explicit
policy constraints, such as an expired recovery window or `blocksResume`, still
prevent recovery.

`retryWith` may name only tasks earlier than or parallel to the declaring task.
The same restriction applies to the full transitive closure of every policy
entry, even entries not reached by the current failure. Ordering comes from
`runAfter` and task references, not YAML list position. Invalid declarations
refuse recovery; warning confirmation cannot override them.

For example, `deploy: {retryWith: [clone, build, scan]}` reruns the earlier work
needed by a failed deploy. `clone` cannot name a later build or scan. An operator
can still select successful tasks explicitly with `--tasks`; if refreshing
clone invalidates a completed build and scan, the operator must select them too.
Tasks that never ran continue normally.

Run the reproducible fixture evaluation:

```bash
go run ./cmd/retryctl evaluate --cases evaluation/cases.yaml
```

The report includes decision correctness, tasks avoided, and the mean task
avoidance rate across recoverable cases.

Generate a recovery `PipelineRun` without submitting it:

```bash
go run ./cmd/retryctl create \
  --pipeline testdata/simple/pipeline.yaml \
  --run testdata/simple/run.yaml \
  --policy testdata/simple/policy.yaml \
  --taskruns testdata/simple/taskruns.yaml > recovery-run.yaml
```

If shared-state warnings are present, `create` asks for confirmation. Automated
jobs must acknowledge them explicitly with `--confirm-warnings`. The generated
PipelineRun records the accepted warning codes in annotations.

The recovery run embeds a reduced `pipelineSpec` containing only rerun and
continue tasks. A `runAfter` edge or result reference to an inherited task is
replaced by edges to its nearest ancestors that still run, so ordering survives
around inherited tasks. String results
of inherited tasks are substituted from the source TaskRuns, and all other
Pipeline and PipelineRun fields (params, retries, timeouts, ...) are
preserved. Finally tasks use the same closure, reuse, and continuation rules as
ordinary tasks and stay in the final execution phase. A failed finally task can
seed recovery, pull in ordinary or parallel finally tasks through `retryWith`,
and use string results inherited from earlier successful tasks. Successful
finally tasks outside the closure are omitted from the recovery run and their
TaskRun provenance is recorded. References using either `$(tasks...)` or
`$(finally...)` are substituted or rejected when they cannot be resolved.
If a Pipeline output exports a string result from a reused task, a generated
result-publishing task emits the saved value: Tekton requires such outputs to
reference task results rather than literals. This adds one helper TaskRun per
reused producer whose outputs are exported; it does not repeat the original
task's work. The generated names are recorded in the
`retry.tekton.dev/result-publishers` annotation. If the recovery has no ordinary tasks, an
always-skipped placeholder satisfies Tekton's requirement for an ordinary task;
it creates no TaskRun. When an inherited task and a remaining task share a Workspace bound
by `volumeClaimTemplate` or `emptyDir`, the planner warns that the recovery run
starts with an empty volume, and when a succeeded task that declares
`blocksResume` is pulled into the closure or selected, it warns that the task
will run again. Shared Workspace and result warnings apply to finally tasks
as well. Like other inferred findings, these never change or refuse the
author's closure; `blocksResume` refuses only when that task is the one that
failed. Generation fails only when a reference
to an inherited task cannot be expressed in a valid PipelineRun (a missing or
non-string result, or a status reference). Annotations record the source
run, policy name and generation, and the source TaskRun of each inherited task.

Recovery requires the complete current Pipeline specification to match the
copy saved in the failed PipelineRun. Changed scripts, variable references,
parameters, task lists, and other specification fields refuse recovery. This
check applies independently of `targetParameters` and `retryOnlyLatest`.
Variables are compared literally: if Tekton saved a resolved value instead of
the original expression, this prototype cannot establish an exact match and
refuses recovery. `PipelineRun.metadata.generation` is the run's own version,
not the referenced Pipeline's version, and is not used for this comparison.
A future controller could capture the Pipeline's UID and generation when a run
starts; this prototype does not assume that historical runs contain them.

The current matched closure and finally-recovery harnesses use TaskRun context
for fault injection and fresh input. Those expressions remain in the saved
PipelineSpec, allowing the strict comparison to accept untouched native
captures. The October 29 reruns and source/binary hashes are retained in
`results/closure-current-20261029/` and `results/finally-current-20261029/`.

When the policy sets `retryOnlyLatest`, pass `--newer-runs-from-cluster` (with
`--kube-context` if needed) to have retryctl list every PipelineRun in the
source run's namespace with kubectl, or `--newer-runs` with a list of your own
(an empty list is allowed); without either, recovery is refused. `targetParameters` restricts supersession to runs with
equal effective values for those parameters, including Pipeline defaults and
explicit run overrides. Defaults come from each run's recorded PipelineSpec,
or its embedded pipelineSpec when no recorded definition is available; defaults
from today's referenced Pipeline are not assumed for older runs. If a required
target value cannot be resolved and the newer run cannot otherwise be excluded
as a different target, recovery is refused. Recovery runs count as newer runs of their
source, except a recovery run that itself failed, so the original run can be
recovered again. `resumeWithin` is measured from the source
run's completion time, falling back to its creation time.

When comparison runs are available, a later run that binds the source's
persistent claim produces `later-run-shares-persistent-claim`, even if it is a
failed recovery or belongs to another target. This is an advisory binding check:
it does not observe writes, restore state, or serialize concurrent requests.

## Development

```bash
go test ./...
```

This first milestone is an offline planner; it does not require a Kubernetes or
Tekton installation. A pinned kind/Tekton execution harness is also available:

```bash
make cluster install-tekton wait-tekton integration
```

See `docs/evaluation.md` for the current evidence and limitations.
