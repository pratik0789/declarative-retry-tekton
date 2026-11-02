Rerunning `mutate` cleans up after itself, so no already-successful task needs to run again.

`mutate` only reads `source.txt` and never writes it, so the clean input `prepare` left on the persistentVolumeClaim (PVC) is still there.
Both files `mutate` may have half-written are fully rewritten when it runs again: `artifact-input.sha` with `>`, and `artifact.txt` is emptied by `printf >` before the append.
`observe` only echoes and passes nothing on. `verify` and `publish` never ran, so they run normally and `publish` checks the fresh artifact.

RERUN: []
