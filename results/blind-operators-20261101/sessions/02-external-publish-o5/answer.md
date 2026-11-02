Rerun `acquire`. The run failed a while ago and we can't tell whether the service's epoch has moved on since (for example after a credential rotation). If it has, the stored credential is from an old epoch and `publish` will get a 409 again. That may even be why `publish` failed the first time.

Running `acquire` again just stores a fresh credential for whatever the current epoch is, so it's cheap and safe either way. `publish`, `sign`, `scan` and `deploy` then all run in that same epoch.

`warmup` only runs `echo`. It has no results, workspaces or calls to the service, so reusing it is safe.

RERUN: [acquire]
