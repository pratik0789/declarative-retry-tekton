The only task that succeeded is `init`. It takes just `enable-cache-proxy`, mounts no workspace, and no later task in this Pipeline uses its results, so reusing it can't change the outcome.

`clone-repository` failed, so it runs again anyway and refills the `workspace` from `git-url`/`revision`. Every task after it never ran, so those run fresh on top of that new clone.

There's no environment note here, so nothing else forces a rerun.

RERUN: []
