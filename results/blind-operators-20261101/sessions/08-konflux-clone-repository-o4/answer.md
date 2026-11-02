Only `init` succeeded. It has no workspaces, so the recovery run loses no files by reusing it.

Its output depends only on the `enable-cache-proxy` param, and its string results are copied into the recovery run unchanged. No task in this Pipeline reads those results.

`clone-repository` reruns automatically because it failed, and every task after it never ran, so all of those run fresh.

Re-running `init` would give the same result, so reusing it is safe.

RERUN: []
