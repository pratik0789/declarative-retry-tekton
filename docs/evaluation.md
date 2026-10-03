# Evaluation status

The offline evaluation contains 13 deterministic cases: two accepted recovery
plans and eleven conservative refusals. All expected decisions currently pass.
The recoverable examples avoid one of four source tasks (25%). These synthetic
results establish planner behavior; they are not production-performance data.

Run them with `make test evaluate`.

The pinned live harness uses kind v0.27.0, Kubernetes v1.31.4, and Tekton
Pipelines v0.62.3. Run `make cluster install-tekton wait-tekton integration`.
The integration injects a source-run failure, exports real PipelineRun and
TaskRun resources, generates a three-task recovery run that inherits one task,
and requires that recovery run to succeed. It also executes a full restart and
records both durations, TaskRun counts, Kubernetes resources, and events under
`artifacts/integration/`. GitHub Actions retains the same evidence as a workflow
artifact.

On the current devserver, cluster creation and Tekton resource installation
succeeded, but official controller images could not be pulled because requests
to `gcr.io` returned HTTP 403. Consequently, no live execution result should be
claimed from this environment until registry access is available.
