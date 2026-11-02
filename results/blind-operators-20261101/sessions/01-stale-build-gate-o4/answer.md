Rerunning prepare writes a new `source.txt` to the shared PVC. Its content includes `$(context.taskRun.name)`, so the hash will be different.
If build is reused, `artifact-input.sha` and `artifact.txt` still match the old source, and publish exits with STALE-ARTIFACT. So build must rerun after prepare.
Gate runs again anyway because it failed, and it doesn't touch the workspace. Observe only echoes and produces nothing anything else uses, so reusing it is safe.
Publish never ran, so it runs normally.

RERUN: [prepare, build]
