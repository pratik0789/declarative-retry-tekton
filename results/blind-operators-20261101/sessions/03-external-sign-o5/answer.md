Nothing in this directory says why `sign` failed. It could have been a brief 503, or the epoch may have advanced and the service returned a 409. Either way, the run failed a while ago, so the epoch may have moved since.
If it has, the credential from `acquire` and the artifact from `publish` are from an old epoch, and the reran `sign` (then `scan`/`deploy`) would be rejected with a 409. Rerunning both puts every record in the current epoch.
`warmup` only runs `echo` and has no service state or results, so reusing it is safe.

RERUN: [acquire, publish]
