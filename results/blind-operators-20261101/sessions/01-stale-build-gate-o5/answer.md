Rerunning prepare rewrites `source.txt` on the shared PVC. The new content includes the new TaskRun name, so its hash changes.
If build is reused, `artifact-input.sha` and `artifact.txt` on the PVC still hold the old hash, and publish fails with STALE-ARTIFACT.
So build must rerun after prepare. Gate reruns anyway because it failed, and it doesn't touch the workspace.
Observe only echoes a line, with no results and no workspace, so it can be reused safely.

RERUN: [prepare, build]
