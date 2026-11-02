Publish needs a stored credential from the service's *current* epoch. The run failed a while ago, and the epoch may have moved on since then (for example after a credential rotation). The publish failure could even have been a 409 caused by exactly that.
If we reuse acquire, its old credential may be rejected, so rerunning acquire is cheap insurance: it stores a fresh credential right before publish.
Warmup only runs `echo` and leaves no service or workspace state, so it's safe to reuse. Sign, scan and deploy never ran, so they will run normally anyway.
RERUN: [acquire]
