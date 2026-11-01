`init` was the first task, and it failed, so no task had succeeded before the run stopped (failed-run.md lists "Succeeded: (none)").
The recovery tool always reruns the failed task (`init`), and the 15 tasks that never ran will run normally after it in Pipeline order.
That leaves no completed task to reuse or force again, so no workspace or result from the failed run can be stale.

RERUN: []
