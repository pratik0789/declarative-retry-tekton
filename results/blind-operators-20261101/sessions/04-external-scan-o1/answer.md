Since the run failed, the service's epoch may have moved on (for example, after a credential rotation). If it has, the stored credential, artifact and signature are from an old epoch, and the rerun of scan would fail with 409.
`acquire`, `publish` and `sign` only write records for the current epoch, so they have to run again in order to rebuild a chain that scan will accept.
`warmup` only runs a local `echo` and never calls the service, so it is safe to reuse.

RERUN: [acquire, publish, sign]
