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
