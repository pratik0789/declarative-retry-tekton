Scan only works if the stored signature is from the service's current epoch. The run failed a while ago, and the epoch may have moved since (for example, after a credential rotation). If it has, the old credential, artifact and signature will all be rejected with 409. Rerunning the chain is safe even if the epoch hasn't changed, because each call just records a fresh value for the current epoch.
Warmup only runs a local `echo` and stores nothing on the service, so it can be reused.

RERUN: [acquire, publish, sign]
