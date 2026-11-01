The epoch may have moved since the run failed. If it did, the stored credential (acquire) and artifact (publish) are stale, and sign would get a 409 again.
Running acquire and then publish again creates fresh current-epoch records before sign. scan and deploy will run normally after that.
warmup only echoes locally and has no external state, so it can be reused safely.
Rerunning acquire and publish costs little and is the only way to be sure the run succeeds, because we can't check the epoch first.

RERUN: [acquire, publish]
