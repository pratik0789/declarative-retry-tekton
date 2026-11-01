# Recovering a failed Tekton PipelineRun

A recovery tool creates a new PipelineRun for the same Pipeline. You choose which tasks run again:
- the failed task always runs again;
- any already-successful task you list runs again too;
- every other successful task is reused: it does not run, its string results are copied into the recovery run, and the recovery run mounts the same Workspace bindings (a persistentVolumeClaim keeps its files; volumeClaimTemplate/emptyDir start empty);
- tasks that never ran run normally.
Tasks keep the Pipeline's order. The tool does not inspect what tasks do.
