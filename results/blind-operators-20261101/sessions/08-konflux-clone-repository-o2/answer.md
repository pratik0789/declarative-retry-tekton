I'd rerun nothing. The only task that succeeded is `init`, and it's safe to reuse.

- **No workspace:** `init` doesn't mount one, so how the recovery run binds workspaces doesn't affect it.
- **Results unused:** no other task reads its results, and copying string results over is harmless.
- **Clone reruns anyway:** `clone-repository` and everything after it run fresh, so the workspace contents and the CHAINS-GIT results get rebuilt correctly.

RERUN: []
