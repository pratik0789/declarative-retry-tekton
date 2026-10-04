# Declarative Retry for Tekton

Research prototype for planning safe, declarative recovery of failed Tekton
`PipelineRun`s. The prototype is external to Tekton and does not represent an
existing Tekton API.

Published result: [end-to-end feasibility report](docs/feasibility-report.md)
(`feasibility-v0.1.0`).

Measured evaluation: [synthetic Tekton benchmark](docs/synthetic-benchmark.md).
The published dataset contains 60 paired live Tekton trials.

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
Pipeline and PipelineRun fields (params, retries, timeouts, `finally`, ...) are
preserved. When an inherited task and a remaining task share a Workspace bound
by `volumeClaimTemplate` or `emptyDir`, the planner warns that the recovery run
starts with an empty volume, and when a succeeded task that declares
`blocksResume` is pulled into the closure or selected, it warns that the task
will run again; it also warns when a `finally` task, which always runs again,
binds a Workspace that inherited tasks used. Like other inferred findings, these never change or refuse the
author's closure; `blocksResume` refuses only when that task is the one that
failed. Generation fails only when a reference
to an inherited task cannot be expressed in a valid PipelineRun (a missing or
non-string result, or a status reference). Annotations record the source
run, policy name and generation, and the source TaskRun of each inherited task.

When the policy sets `retryOnlyLatest`, pass `--newer-runs` with the
PipelineRuns to compare against (an empty list is allowed); without it,
recovery is refused. `targetParameters` restricts supersession to runs with
equal values for those parameters. `resumeWithin` is measured from the source
run's completion time, falling back to its creation time.

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
