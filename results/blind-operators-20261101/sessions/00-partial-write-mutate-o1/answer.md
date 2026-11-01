Workspace `shared` is on a PVC, so the recovery run sees the files `prepare` wrote. `mutate` only reads `source.txt` and never writes to it, so the file is still the clean value `prepare` left.

Anything `mutate` half-wrote (`artifact-input.sha`, `artifact.txt`) gets overwritten when it reruns: both files are opened with `>`, which empties them first, so no stale partial output can survive into `publish`'s checks.

`observe` only echoes, and `prepare` has nothing new to pick up, so reusing both is safe.

RERUN: []
