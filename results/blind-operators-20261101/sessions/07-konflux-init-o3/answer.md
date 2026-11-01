The first task, `init`, failed, and failed-run.md lists no successful tasks.
So there is no completed work to reuse and nothing whose results or workspace files could be stale.
The recovery run re-runs `init` (the failed task always runs again), and every other task runs normally because none of them ever ran.

RERUN: []
