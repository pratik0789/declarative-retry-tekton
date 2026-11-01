The sign failure may have been a 409 from an epoch change or a transient 503, and the service may have changed since. If the epoch advanced, the reused credential and artifact are stale, and sign (and everything after it) would be rejected again.
Re-running acquire and then publish puts the credential and artifact in the current epoch before sign runs. The service checks only records it stores, not Tekton results or workspaces, so the reused records have to be refreshed by calling the service again.
warmup only runs `echo warmed` and never calls the service, so it is safe to reuse.
scan and deploy never ran, so they run normally.

RERUN: [acquire, publish]
