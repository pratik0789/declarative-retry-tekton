Publish only succeeds if the stored credential is from the service's current epoch. The run failed a while ago, and the failure may have been a 409 from an epoch change rather than a 503. Either way, the old `acquire` credential may now be stale.

Re-running `acquire` stores a credential for the current epoch, so `publish`, `sign`, `scan` and `deploy` then run in one consistent epoch. It's a single cheap call.

`warmup` only echoes text and produces no results, files or service state, so it's safe to reuse.

RERUN: [acquire]
