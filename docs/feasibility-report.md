# Declarative Retry for Tekton: Feasibility Result

Date: 2026-10-03  
Published snapshot: `feasibility-v0.1.0`

## Claim supported

The prototype can ingest real Tekton `PipelineRun` and `TaskRun` resources,
derive a conservative partial-recovery plan, generate a valid recovery
`PipelineRun`, and execute that run successfully on Tekton. In the controlled
pipeline below, recovery reused one successful task and executed the remaining
three tasks instead of executing all four tasks again.

This is a feasibility result. It is not evidence of production-scale savings
or statistically significant performance improvement.

## Environment

- kind v0.27.0
- Kubernetes v1.31.4
- Tekton Pipelines v1.6.0
- Four-task pipeline: `clone -> build -> scan -> deploy`
- Injected failure: `scan` in the source PipelineRun
- Recovery policy: rerun `scan` together with `build`; inherit `clone`; continue
  with `deploy`

## Observed execution

| Execution | Outcome | TaskRuns | Wall time |
|---|---:|---:|---:|
| Full restart | Succeeded | 4 | 17.896 s |
| Declarative recovery | Succeeded | 3 | 13.970 s |

The recovery run avoided one of four TaskRuns (25%). The timing values are from
one controlled run and must not be presented as a general speedup.

## Reproducibility and evidence

- Successful independent CI execution: [GitHub Actions run 37163358593](https://github.com/pratik0789/declarative-retry-tekton/actions/runs/37163358593)
- CI artifact: `tekton-integration-evidence`
- Machine-readable summary: [`results/kind-v1.31.4-tekton-v1.6.0/summary.json`](../results/kind-v1.31.4-tekton-v1.6.0/summary.json)
- Generated recovery manifest: [`recovery.yaml`](../results/kind-v1.31.4-tekton-v1.6.0/recovery.yaml)
- Captured PipelineRuns: [`pipelineruns.yaml`](../results/kind-v1.31.4-tekton-v1.6.0/pipelineruns.yaml)
- Captured TaskRuns: [`taskruns.yaml`](../results/kind-v1.31.4-tekton-v1.6.0/taskruns.yaml)
- Kubernetes events: [`events.txt`](../results/kind-v1.31.4-tekton-v1.6.0/events.txt)

Reproduce the experiment with:

```bash
make test evaluate
make cluster install-tekton wait-tekton integration
```

## Suggested paper wording

> We implemented an external prototype that reads Tekton PipelineRun and
> TaskRun state, computes a conservative recovery plan, and generates a new
> PipelineRun. In a four-task kind-based feasibility experiment, both a full
> restart and the generated recovery run completed successfully. The recovery
> inherited one previously successful task and executed three TaskRuns rather
> than four. This experiment demonstrates end-to-end feasibility; broader and
> repeated experiments are required to quantify performance benefits.

## Limitations

- The reported timing is a single observation.
- The tasks use small synthetic shell workloads.
- The result does not measure production behavior or cost.
- The planner intentionally refuses unsupported or ambiguous recovery cases.
- General effectiveness requires repeated trials across additional pipeline
  shapes, failure positions, task durations, workspaces, and result flows.
