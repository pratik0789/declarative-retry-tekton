# Declarative Retry for Tekton

Research prototype for planning safe, declarative recovery of failed Tekton
`PipelineRun`s. The prototype is external to Tekton and does not represent an
existing Tekton API.

Published result: [end-to-end feasibility report](docs/feasibility-report.md)
(`feasibility-v0.1.0`).

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
provenance. It conservatively refuses recovery when the pipeline changed, a
rerun producer conflicts with an inherited consumer, mutable workspaces cross
the recovery boundary, or the recovery window has expired.

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
