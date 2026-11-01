# PipelineRetryPolicy semantics (Tekton recovery contract)

When a Tekton PipelineRun fails, a recovery tool creates a new PipelineRun for the same Pipeline. Each task gets one action:
- rerun: the failed task(s), any task the operator explicitly selects with --tasks, and every task in their retry closure (below).
- reuse: every other task that succeeded in the failed run. It does not run again. Its string results are copied into the recovery run, and the recovery run mounts the same Workspace bindings (a persistentVolumeClaim keeps its files; a volumeClaimTemplate or emptyDir starts empty).
- continue: tasks that never ran in the failed run run normally.
Tasks keep the order the Pipeline defines.

The author writes a policy:

```yaml
apiVersion: retry.tekton.dev/v1alpha1
kind: PipelineRetryPolicy
metadata: {name: <name>}
spec:
  pipelineRef: <pipeline name>
  tasks:
    <task>:
      retryWith: [<task>, ...]   # if <task> is rerun, these tasks must be rerun with it
      blocksResume: true         # optional: if <task> is the one that failed, refuse partial recovery
```

Rules:
- retryWith is followed transitively: if A names B and B names C, rerunning A reruns B and C.
- retryWith may name only tasks that come before the declaring task in the Pipeline's execution order (runAfter or result references) or that are on a parallel branch. Naming a later task makes the policy invalid.
- Tasks not listed have no retryWith entries.
- The tool trusts the policy; it does not inspect what tasks do.
