Rationale:
- `mutate` never writes `source.txt`. It only rewrites `artifact-input.sha` and `artifact.txt` from scratch (`>` truncates each file before writing). So any half-written files left by the failed attempt get replaced when it reruns.
- The shared Workspace is a persistentVolumeClaim, so the clean `source.txt` that `prepare` wrote is still there. Nothing after `prepare` changes it, so `mutate`'s input check and `publish`'s hash check will both see the right input.
- `observe` only prints a line and produces no results or files, so reusing it is safe. Rerunning `prepare` would do no harm, but nothing in the Pipeline needs it.

RERUN: []
