Mutate never writes `source.txt`. It only rewrites `artifact-input.sha` (with `>`) and `artifact.txt` (`printf >` then `>>`), so a rerun fully overwrites anything a partial write left behind.

The shared workspace is a persistentVolumeClaim, so the clean `source.txt` from prepare is still there for the recovery run. Prepare has no external input to refresh, so running it again would give the same result.

Observe only echoes a line, and nothing downstream uses it. Mutate reruns anyway because it failed, and verify and publish run normally, so publish will check a freshly rebuilt artifact.

RERUN: []
