Rerunning prepare writes a new `source.txt`, because the file includes `$(context.taskRun.name)` and that changes with each run. The workspace is a persistentVolumeClaim, so if build were reused, the old `artifact-input.sha` and `artifact.txt` would stay behind. Publish would then compare them to the new `source.txt` and fail with STALE-ARTIFACT.
So build must run again after prepare. Gate isn't listed because failed tasks rerun anyway. Observe only echoes, has no results and uses no workspace, so it can safely be reused.

RERUN: [prepare, build]
