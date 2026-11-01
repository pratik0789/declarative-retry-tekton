No task needs to rerun. `init` is the only task that succeeded, and reusing it is safe.

`init` mounts no workspace, and no later task reads its results. Reusing it skips a run but leaves nothing stale downstream.

`clone-repository` failed, so it runs again automatically. Every later task never ran, so they all run normally from the fresh clone.

RERUN: []
