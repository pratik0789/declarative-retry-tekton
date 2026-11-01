Scan's failure could be a 503 or a 409 from a stale epoch, and the run is old enough that the epoch may have advanced since. If it has, the stored credential, artifact and signature are all rejected.
Each of those records needs the one before it from the same epoch, so `acquire`, `publish` and `sign` must all run again before `scan`.
Running them again is harmless if the epoch hasn't changed: they just write fresh records for the current epoch.
`warmup` only echoes locally, has no results and touches no service state, so it can be safely reused.

RERUN: [acquire, publish, sign]
