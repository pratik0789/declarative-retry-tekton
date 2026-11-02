Rerunning `mutate` on its own is enough, so no already-successful task needs to run again.

- The shared Workspace is a persistentVolumeClaim, so the recovery run still has the `source.txt` that `prepare` wrote. `mutate` only reads that file, so its partial writes can't have changed it.
- Rerunning `mutate` overwrites `artifact-input.sha` and `artifact.txt` completely (with `>`), so nothing it half-wrote survives. `verify` and `publish` never ran, so they run normally.
- No task passes string results to another, and `observe` only echoes a line, so reusing `observe` and `prepare` is safe.

RERUN: []
