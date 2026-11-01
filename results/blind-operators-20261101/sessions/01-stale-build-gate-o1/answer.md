Rerunning `prepare` writes a new `source.txt` to the persistent volume, and its contents change because they include the new TaskRun name (`$(context.taskRun.name)`).
If `build` is reused, `artifact-input.sha` and `artifact.txt` still hold the hash of the old `source.txt`, so `publish` would fail with STALE-ARTIFACT. `build` has to run again.
`observe` only echoes text and writes nothing that later tasks read, so it can be reused safely. `gate` reruns anyway because it failed, and `publish` never ran, so it runs normally.

RERUN: [prepare, build]
