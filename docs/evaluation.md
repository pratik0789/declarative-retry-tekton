# Evaluation status

The offline evaluation contains 13 deterministic cases: two accepted recovery
plans and eleven conservative refusals. All expected decisions currently pass.
The recoverable examples avoid one of four source tasks (25%). These synthetic
results establish planner behavior; they are not production-performance data.

Run them with `make test evaluate`.

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

The benchmark reports TaskRuns avoided and observed wall time. Synthetic timing
must be described as controlled experimental evidence, not production savings.
