# Evaluation status

The offline evaluation contains 20 deterministic cases: nine accepted recovery
plans and eleven policy or structural refusals. Four accepted plans carry
warnings (a shared result, a writable Workspace, a Workspace bound by
`volumeClaimTemplate`, and a succeeded `blocksResume` task pulled into the
closure). All expected decisions, dispositions, refusal reasons,
and warnings pass. These cases establish planner behavior; they are not
production-performance data.

Run them with `make test evaluate`.

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
v1.31.4 / Tekton v1.6.0; all six gave the same four outcomes. Results are in
`results/v0.3.0-kind-v1.31.4-tekton-v1.6.0/stateful/`.

## Konflux replay

`make replay` injects a failure at each task of the Konflux template-build
Pipeline (vendored, unmodified, at upstream commit bfe4d25) and records the plan
under the author policy in `evaluation/konflux/policy.yaml` and under an empty
policy, with the source Workspace bound by a persistent claim or a
`volumeClaimTemplate`. Only tasks upstream of the failure are treated as
completed. This exercises the planner only, not a live Konflux installation.

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

The benchmark reports TaskRuns avoided and observed wall time. Synthetic timing
must be described as controlled experimental evidence, not production savings.
